package builder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Builder struct {
	cmd    *exec.Cmd
	runCmd string
}

func New(runCmd string) *Builder {
	return &Builder{
		cmd:    nil,
		runCmd: runCmd,
	}
}

func (b *Builder) Start() error {
	if b.cmd != nil {
		return fmt.Errorf("Process already started")
	}

	parts := strings.Fields(b.runCmd)

	if len(parts) == 0 {
		return fmt.Errorf("Invalid command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()

	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	b.cmd = cmd

	return nil
}

func (b *Builder) Stop() error {
	if b.cmd == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(b.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGKILL)
	} else {
		b.cmd.Process.Kill()
	}

	b.cmd.Wait()
	b.cmd = nil

	return nil
}

func (b *Builder) Build(buildCmd string) error {
	parts := strings.Fields(buildCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid build command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func KillProcessOnPort(port int) error {
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()

	if err != nil {
		return nil
	}

	pid := strings.TrimSpace(string(output))
	if pid == "" {
		return nil
	}

	killCmd := exec.Command("kill", "-9", pid)
	return killCmd.Run()
}
