package processmng

import (
	"fmt"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

type NginxProcessManager struct {
	proc *process.Process
}

func (m *NginxProcessManager) Reload() error {
	err := m.proc.SendSignal(syscall.SIGHUP)

	if err != nil {
		return fmt.Errorf("failed to reload nginx: %v", err)
	}

	return nil
}

func GetNginxProcessManager() (*NginxProcessManager, error) {
	nginxProcess, err := findProcessByName([]string{"nginx", "httpd"})

	if err != nil {
		return nil, err
	}

	if nginxProcess == nil {
		return nil, fmt.Errorf("failed to find nginx process")
	}

	isRunning, err := nginxProcess.IsRunning()

	if err != nil {
		return nil, err
	}

	if !isRunning {
		return nil, fmt.Errorf("nginx process is not running")
	}

	return &NginxProcessManager{nginxProcess}, nil
}
