package services

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// TerminalService manages shell sessions used by the live terminal UI.
type TerminalService struct {
	shell   string
	workDir string
}

// NewTerminalService creates a terminal service using the current shell and working directory.
func NewTerminalService() *TerminalService {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	if _, err := exec.LookPath(shell); err != nil {
		shell = "/bin/sh"
	}

	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	return &TerminalService{
		shell:   shell,
		workDir: wd,
	}
}

// StartSession starts a new interactive shell session backed by a PTY.
func (s *TerminalService) StartSession(ctx context.Context, cols, rows int) (*TerminalSession, error) {
	cmd := exec.CommandContext(ctx, s.shell)
	cmd.Dir = s.workDir
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	session := &TerminalSession{
		cmd:   cmd,
		pty:   ptmx,
		done:  make(chan error, 1),
		close: sync.Once{},
	}

	if cols > 0 && rows > 0 {
		_ = pty.Setsize(ptmx, &pty.Winsize{ // best-effort resize before client handshake
			Rows: uint16(rows),
			Cols: uint16(cols),
		})
	}

	go func() {
		session.done <- cmd.Wait()
		close(session.done)
	}()

	go func() {
		<-ctx.Done()
		_ = session.Close()
	}()

	return session, nil
}

// TerminalSession contains the running shell process information.
type TerminalSession struct {
	cmd   *exec.Cmd
	pty   *os.File
	done  chan error
	close sync.Once
}

// PTY returns the PTY backing the session.
func (s *TerminalSession) PTY() *os.File {
	return s.pty
}

// Resize updates the pseudo terminal window size.
func (s *TerminalSession) Resize(cols, rows int) error {
	if s.pty == nil {
		return errors.New("terminal session not initialized")
	}

	if cols <= 0 || rows <= 0 {
		return nil
	}

	return pty.Setsize(s.pty, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Close terminates the running session and releases PTY resources.
func (s *TerminalSession) Close() error {
	var err error
	s.close.Do(func() {
		if s.pty != nil {
			err = s.pty.Close()
		}
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
	})
	return err
}

// Wait blocks until the shell process exits.
func (s *TerminalSession) Wait() error {
	if s.done == nil {
		return nil
	}

	err, ok := <-s.done
	if !ok {
		return nil
	}

	return err
}
