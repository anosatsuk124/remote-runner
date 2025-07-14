package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anosatsuk124/remote-runner/internal/config"
)

type ExecuteRequest struct {
	Executable string `json:"executable"`
	WorkDir    string `json:"workdir"`
}

func Execute(cfg *config.Config, executable string) error {
	req := ExecuteRequest{
		Executable: executable,
		WorkDir:    cfg.Remote.Path,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/execute", cfg.Remote.HttpHost, cfg.Remote.Port)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}