// Package mapper provides file-to-process mapping functionality.
package mapper

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/thiruk/logmonster/pkg/types"
)

// Mapper maps files to processes.
type Mapper struct{}

// New creates a new Mapper.
func New() *Mapper {
	return &Mapper{}
}

// FindProcessForFile finds the process(es) writing to a file.
func (m *Mapper) FindProcessForFile(filePath string) ([]types.ProcessInfo, error) {
	// Try lsof first
	pids, err := m.findPIDsWithLsof(filePath)
	if err != nil || len(pids) == 0 {
		// Try /proc fallback
		pids, err = m.findPIDsFromProc(filePath)
		if err != nil {
			return nil, err
		}
	}

	if len(pids) == 0 {
		return nil, fmt.Errorf("no process found with file open: %s", filePath)
	}

	var processes []types.ProcessInfo
	for _, pid := range pids {
		info, err := m.GetProcessInfo(pid)
		if err != nil {
			continue // Process may have exited
		}
		processes = append(processes, *info)
	}

	return processes, nil
}

// findPIDsWithLsof uses lsof to find PIDs with a file open.
func (m *Mapper) findPIDsWithLsof(filePath string) ([]int32, error) {
	cmd := exec.Command("lsof", "-t", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pids []int32
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		pid, err := strconv.ParseInt(line, 10, 32)
		if err != nil {
			continue
		}
		pids = append(pids, int32(pid))
	}

	return pids, nil
}

// findPIDsFromProc searches /proc for processes with the file open.
func (m *Mapper) findPIDsFromProc(filePath string) ([]int32, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	var pids []int32

	// Read all /proc entries
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.ParseInt(entry.Name(), 10, 32)
		if err != nil {
			continue // Not a PID directory
		}

		// Check fd directory
		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue // Permission denied or process exited
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == absPath {
				pids = append(pids, int32(pid))
				break
			}
		}
	}

	return pids, nil
}

// GetProcessInfo retrieves detailed information about a process.
func (m *Mapper) GetProcessInfo(pid int32) (*types.ProcessInfo, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}

	name, _ := proc.Name()
	cmdline, _ := proc.Cmdline()
	exe, _ := proc.Exe()
	username, _ := proc.Username()
	createTime, _ := proc.CreateTime()
	cpuPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()

	var memoryMB float64
	if memInfo != nil {
		memoryMB = float64(memInfo.RSS) / (1024 * 1024)
	}

	// Get write bytes from /proc/[pid]/io
	writeBytes := m.getWriteBytes(pid)

	startTime := time.Unix(createTime/1000, 0)

	return &types.ProcessInfo{
		PID:        pid,
		Name:       name,
		Cmdline:    cmdline,
		Exe:        exe,
		User:       username,
		StartTime:  startTime,
		CPUPercent: cpuPercent,
		MemoryMB:   memoryMB,
		WriteBytes: writeBytes,
	}, nil
}

// getWriteBytes reads write_bytes from /proc/[pid]/io.
func (m *Mapper) getWriteBytes(pid int32) int64 {
	path := fmt.Sprintf("/proc/%d/io", pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "write_bytes:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bytes, _ := strconv.ParseInt(parts[1], 10, 64)
				return bytes
			}
		}
	}

	return 0
}
