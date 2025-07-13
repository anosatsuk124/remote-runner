package rsync

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/anosatsuk124/remote-runner/internal/config"
)

func Sync(cfg *config.Config) error {
	args := []string{
		"-avz",
		"--delete",
	}

	for _, exclude := range cfg.Sync.Exclude {
		args = append(args, "--exclude", exclude)
	}

	source := cfg.Sync.Source
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}

	destination := fmt.Sprintf("%s:%s/", cfg.Remote.Host, cfg.Remote.Path)
	
	args = append(args, source, destination)

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync command failed: %w", err)
	}

	return nil
}