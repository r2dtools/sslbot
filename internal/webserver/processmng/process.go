package processmng

import (
	"strings"

	"github.com/shirou/gopsutil/process"
)

func findProcessByName(names []string) (*process.Process, error) {
	processes, err := process.Processes()

	if err != nil {
		return nil, err
	}

	var requiredProcess *process.Process

	for _, p := range processes {
		cmd, err := p.Name()

		if err != nil {
			return nil, err
		}

		if !isRequiredProcess(cmd, names) {
			continue
		}

		requiredProcess = p
		pp, err := p.Parent()

		if err != nil {
			return nil, err
		}

		if pp == nil {
			break
		}

		ppCmd, err := pp.Name()

		if err != nil {
			return nil, err
		}

		if !isRequiredProcess(ppCmd, names) {
			break
		}
	}

	return requiredProcess, nil
}

func isRequiredProcess(cmd string, names []string) bool {
	for _, name := range names {
		if strings.Contains(cmd, name) {
			return true
		}
	}

	return false
}
