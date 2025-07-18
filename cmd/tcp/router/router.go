package router

import "fmt"

type HandlerInterface interface {
	Handle(request Request) (any, error)
}

type Router struct {
	handlers map[string]HandlerInterface
}

func (r *Router) RegisterHandler(module string, handler HandlerInterface) {
	if r.handlers == nil {
		r.handlers = make(map[string]HandlerInterface)
	}

	r.handlers[module] = handler
}

func (r *Router) GetHandler(request Request) HandlerInterface {
	if r.handlers == nil {
		r.handlers = make(map[string]HandlerInterface)
	}

	return r.handlers[request.GetModule()]
}

func (r *Router) HandleRequest(request Request) (any, error) {
	handler := r.GetHandler(request)

	if handler == nil {
		return nil, fmt.Errorf("could not find handler for the command '%s'", request.Command)
	}

	return handler.Handle(request)
}
