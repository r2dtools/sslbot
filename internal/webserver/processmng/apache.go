package processmng

import (
	"fmt"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

type ApacheProcessManager struct {
	proc *process.Process
}

func (m *ApacheProcessManager) Reload() error {
	err := m.proc.SendSignal(syscall.SIGHUP)

	if err != nil {
		return fmt.Errorf("failed to reload apache: %v", err)
	}

	return nil
}

func GetApacheProcessManager() (*ApacheProcessManager, error) {
	apacheProcess, err := findProcessByName([]string{"apache2", "httpd"})

	if err != nil {
		return nil, err
	}

	if apacheProcess == nil {
		return nil, fmt.Errorf("apache process not found")
	}

	isRunning, err := apacheProcess.IsRunning()

	if err != nil {
		return nil, err
	}

	if !isRunning {
		return nil, fmt.Errorf("apache process is not running")
	}

	return &ApacheProcessManager{apacheProcess}, nil
}
