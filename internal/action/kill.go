// Package action provides process control functionality.
package action

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Killer handles process termination.
type Killer struct {
	Timeout time.Duration
}

// NewKiller creates a new Killer.
func NewKiller(timeout time.Duration) *Killer {
	return &Killer{Timeout: timeout}
}

// Kill terminates a process gracefully, then forcefully if needed.
func (k *Killer) Kill(pid int32) error {
	// Check if process exists
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Errorf("process not found: %d", pid)
	}

	// Send SIGTERM
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if err == os.ErrProcessDone {
			return nil // Already dead
		}
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for process to exit
	done := make(chan bool, 1)
	go func() {
		for {
			if !k.processExists(pid) {
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(k.Timeout):
		// Send SIGKILL
		if err := proc.Signal(syscall.SIGKILL); err != nil {
			if err == os.ErrProcessDone {
				return nil
			}
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
		// Wait a bit more for SIGKILL
		time.Sleep(500 * time.Millisecond)
		if k.processExists(pid) {
			return fmt.Errorf("process %d still running after SIGKILL", pid)
		}
		return nil
	}
}

// processExists checks if a process is still running.
func (k *Killer) processExists(pid int32) bool {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without sending a signal
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// SendSignal sends a specific signal to a process.
func (k *Killer) SendSignal(pid int32, sig syscall.Signal) error {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Errorf("process not found: %d", pid)
	}

	if err := proc.Signal(sig); err != nil {
		return fmt.Errorf("failed to send signal %d: %w", sig, err)
	}

	return nil
}
