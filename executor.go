package main

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

type ShellExecutor struct {
	Timeout time.Duration
}

func (e *ShellExecutor) Exec(ctx context.Context, command string, out io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	command = strings.TrimSpace(command)

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
