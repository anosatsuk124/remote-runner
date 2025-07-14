package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/anosatsuk124/remote-runner/internal/signal"
)

type ExecuteRequest struct {
	Executable string `json:"executable"`
	WorkDir    string `json:"workdir"`
}

type Daemon struct {
	port         int
	currentProc  *exec.Cmd
	procMutex    sync.Mutex
	server       *http.Server
	signalHandle *signal.Handler
}

func New(port int) *Daemon {
	return &Daemon{
		port: port,
	}
}

func (d *Daemon) Start() error {
	d.signalHandle = signal.NewHandler(d.killCurrentProcess, d.gracefulShutdown)
	
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", d.handleExecute)
	
	d.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: mux,
	}

	fmt.Printf("Daemon listening on port %d\n", d.port)
	
	go d.signalHandle.Start()
	
	if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	
	return nil
}

func (d *Daemon) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := d.executeCommand(req.Executable, req.WorkDir); err != nil {
		http.Error(w, fmt.Sprintf("Execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *Daemon) executeCommand(executable, workdir string) error {
	d.procMutex.Lock()
	defer d.procMutex.Unlock()

	if d.currentProc != nil && d.currentProc.Process != nil {
		d.currentProc.Process.Kill()
		d.currentProc.Wait()
	}

	execPath := executable
	if !strings.HasPrefix(executable, "/") {
		execPath = fmt.Sprintf("%s/%s", workdir, executable)
	}

	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return fmt.Errorf("executable '%s' not found in workdir '%s'", executable, workdir)
	}

	cmd := exec.Command(execPath)
	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	d.currentProc = cmd

	fmt.Printf("Executing: %s (in %s)\n", execPath, workdir)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start '%s' in '%s': %w", execPath, workdir, err)
	}

	go func() {
		cmd.Wait()
		d.procMutex.Lock()
		if d.currentProc == cmd {
			d.currentProc = nil
		}
		d.procMutex.Unlock()
		fmt.Printf("Process '%s' finished\n", execPath)
	}()

	return nil
}

func (d *Daemon) killCurrentProcess() {
	d.procMutex.Lock()
	defer d.procMutex.Unlock()

	if d.currentProc != nil && d.currentProc.Process != nil {
		fmt.Println("Killing current process...")
		
		pgid, err := syscall.Getpgid(d.currentProc.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		} else {
			d.currentProc.Process.Kill()
		}
		
		d.currentProc.Wait()
		d.currentProc = nil
		fmt.Println("Process killed, daemon ready for new requests")
	}
}

func (d *Daemon) Stop() error {
	if d.server != nil {
		return d.server.Shutdown(context.Background())
	}
	return nil
}

func (d *Daemon) gracefulShutdown() {
	fmt.Println("Ctrl+C pressed twice - shutting down daemon...")
	
	d.procMutex.Lock()
	if d.currentProc != nil && d.currentProc.Process != nil {
		fmt.Println("Killing current process before shutdown...")
		pgid, err := syscall.Getpgid(d.currentProc.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		} else {
			d.currentProc.Process.Kill()
		}
		d.currentProc.Wait()
	}
	d.procMutex.Unlock()
	
	if d.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		d.server.Shutdown(ctx)
	}
	
	os.Exit(0)
}