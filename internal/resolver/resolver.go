// Package resolver provides PID-to-service resolution functionality.
package resolver

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/thiruk/logmonster/pkg/types"
)

// Resolver resolves PIDs to systemd services.
type Resolver struct {
	conn *dbus.Conn
}

// New creates a new Resolver.
func New() (*Resolver, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		// D-Bus not available, will use fallback
		return &Resolver{conn: nil}, nil
	}
	return &Resolver{conn: conn}, nil
}

// Close closes the D-Bus connection.
func (r *Resolver) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

// ResolveService resolves a PID to its systemd service.
func (r *Resolver) ResolveService(pid int32) (*types.ServiceInfo, error) {
	if r.conn != nil {
		// Try systemd first
		info, err := r.resolveWithSystemd(pid)
		if err == nil && info != nil {
			return info, nil
		}
	}

	// Fallback to process tree analysis
	return r.resolveFromProcessTree(pid)
}

// resolveWithSystemd uses D-Bus to query systemd.
func (r *Resolver) resolveWithSystemd(pid int32) (*types.ServiceInfo, error) {
	obj := r.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	var unitPath dbus.ObjectPath
	err := obj.Call("org.freedesktop.systemd1.Manager.GetUnitByPID", 0, uint32(pid)).Store(&unitPath)
	if err != nil {
		return nil, err
	}

	// Get unit properties
	unitObj := r.conn.Object("org.freedesktop.systemd1", unitPath)

	var unitName string
	err = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Unit", "Id").Store(&unitName)
	if err != nil {
		unitName = string(unitPath)
	}

	var activeState string
	_ = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Unit", "ActiveState").Store(&activeState)

	var mainPID uint32
	_ = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Service", "MainPID").Store(&mainPID)

	var description string
	_ = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Unit", "Description").Store(&description)

	return &types.ServiceInfo{
		Unit:        unitName,
		Status:      activeState,
		MainPID:     int32(mainPID),
		Description: description,
	}, nil
}

// resolveFromProcessTree walks the process tree to find a service.
func (r *Resolver) resolveFromProcessTree(pid int32) (*types.ServiceInfo, error) {
	// Walk up the process tree
	currentPID := pid
	for currentPID > 1 {
		// Check if this process is a service
		serviceName := r.getServiceNameFromComm(currentPID)
		if serviceName != "" {
			return &types.ServiceInfo{
				Unit:    serviceName,
				Status:  "unknown (fallback)",
				MainPID: currentPID,
			}, nil
		}

		// Get parent PID
		parentPID, err := r.getParentPID(currentPID)
		if err != nil || parentPID <= 1 {
			break
		}
		currentPID = parentPID
	}

	return nil, fmt.Errorf("could not resolve PID %d to a service", pid)
}

// getServiceNameFromComm tries to determine service name from /proc/[pid]/comm.
func (r *Resolver) getServiceNameFromComm(pid int32) string {
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := os.ReadFile(commPath)
	if err != nil {
		return ""
	}
	comm := strings.TrimSpace(string(data))

	// Common service patterns
	serviceNames := []string{
		"apache2", "nginx", "mysql", "postgres", "redis",
		"docker", "containerd", "tomcat", "java", "node",
		"python", "php", "ruby", "mongod", "elasticsearch",
	}

	for _, svc := range serviceNames {
		if strings.Contains(strings.ToLower(comm), svc) {
			return comm + ".service"
		}
	}

	return ""
}

// getParentPID reads the parent PID from /proc/[pid]/stat.
func (r *Resolver) getParentPID(pid int32) (int32, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0, err
	}

	// Parse stat file: pid (comm) state ppid ...
	content := string(data)

	// Find the closing parenthesis of comm
	lastParen := strings.LastIndex(content, ")")
	if lastParen == -1 {
		return 0, fmt.Errorf("invalid stat format")
	}

	// Fields after the parenthesis
	fields := strings.Fields(content[lastParen+1:])
	if len(fields) < 2 {
		return 0, fmt.Errorf("invalid stat format")
	}

	ppid, err := strconv.ParseInt(fields[1], 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(ppid), nil
}

// GetServiceStatus returns the status of a systemd service.
func (r *Resolver) GetServiceStatus(unitName string) (string, error) {
	if r.conn == nil {
		return "unknown", fmt.Errorf("D-Bus not available")
	}

	obj := r.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	var unitPath dbus.ObjectPath
	err := obj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, unitName).Store(&unitPath)
	if err != nil {
		return "unknown", err
	}

	unitObj := r.conn.Object("org.freedesktop.systemd1", unitPath)

	var activeState string
	err = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Unit", "ActiveState").Store(&activeState)
	if err != nil {
		return "unknown", err
	}

	return activeState, nil
}

// GetServiceStartTime returns when a service was started.
func (r *Resolver) GetServiceStartTime(unitName string) (time.Time, error) {
	if r.conn == nil {
		return time.Time{}, fmt.Errorf("D-Bus not available")
	}

	obj := r.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	var unitPath dbus.ObjectPath
	err := obj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, unitName).Store(&unitPath)
	if err != nil {
		return time.Time{}, err
	}

	unitObj := r.conn.Object("org.freedesktop.systemd1", unitPath)

	var timestamp uint64
	err = unitObj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.systemd1.Unit", "ActiveEnterTimestamp").Store(&timestamp)
	if err != nil {
		return time.Time{}, err
	}

	// Timestamp is in microseconds since epoch
	return time.Unix(int64(timestamp/1000000), int64(timestamp%1000000)*1000), nil
}
