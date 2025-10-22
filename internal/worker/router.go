package worker

import (
	"errors"
	"fmt"
)

var ErrHandlerNotFound = errors.New("handler not found")

type EventHandler func([]byte) error

type Router struct {
	Handlers map[string][]EventHandler
}

func NewRouter(handlers map[string][]EventHandler) *Router {
	return &Router{
		Handlers: handlers,
	}
}

func (this *Router) Handle(event string, data []byte) error {
	handlers, ok := this.Handlers[event]
	if !ok {
		return ErrHandlerNotFound
	}

	for _, handler := range handlers {
		err := handler(data)
		if err != nil {
			return fmt.Errorf("error handling event: %w", err)
		}
	}

	return nil
}
