package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/anosatsuk124/remote-runner/internal/config"
	"github.com/anosatsuk124/remote-runner/internal/rsync"
	"github.com/anosatsuk124/remote-runner/internal/client"
	"github.com/anosatsuk124/remote-runner/internal/daemon"
)

var (
	daemonMode bool
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "remote-run [executable]",
	Short: "Remote file sync and execution tool",
	Long:  "A tool for syncing files to remote host and executing programs remotely",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if daemonMode {
			runDaemon()
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: executable name is required")
			os.Exit(1)
		}

		executable := args[0]
		runClient(executable)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in daemon mode")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "remote-run.yaml", "Config file path")
}

func runClient(executable string) {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Syncing files to %s:%s\n", cfg.Remote.Host, cfg.Remote.Path)
	if err := rsync.Sync(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Sync failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Executing %s on remote host\n", executable)
	if err := client.Execute(cfg, executable); err != nil {
		fmt.Fprintf(os.Stderr, "Remote execution failed: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon() {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	daemon := daemon.New(cfg.Remote.Port)
	if err := daemon.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Daemon failed: %v\n", err)
		os.Exit(1)
	}
}