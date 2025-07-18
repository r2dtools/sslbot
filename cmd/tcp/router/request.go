package router

import "strings"

type Request struct {
	Command,
	Token string
	Data any
}

func (r *Request) GetModule() string {
	parts := strings.SplitN(r.Command, ".", 2)

	if len(parts) == 2 {
		return parts[0]
	}

	return "main"
}

func (r *Request) GetAction() string {
	parts := strings.SplitN(r.Command, ".", 2)

	if len(parts) == 2 {
		return parts[1]
	}

	return r.Command
}
