package middleware

import (
	contextpkg "context"
	"errors"
)

var ErrSessionIDNotSet = errors.New("session_id not set")

func SetSessionID(context contextpkg.Context, id string) contextpkg.Context {
	return contextpkg.WithValue(context, "sessionID", id)
}

func GetSessionID(context contextpkg.Context) (string, error) {
	if val := context.Value("sessionID"); val != nil {
		if id, ok := val.(string); ok {
			return id, nil
		}
	}

	return "", ErrSessionIDNotSet
}
