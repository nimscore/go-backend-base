package middleware

import (
	contextpkg "context"
	"errors"
	"fmt"

	uuidpkg "github.com/google/uuid"
)

var ErrSessionIDNotSet = errors.New("session_id not set")
var ErrUserIDNotSet = errors.New("user_id not set")

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

func SetUserID(context contextpkg.Context, id string) contextpkg.Context {
	return contextpkg.WithValue(context, "userID", id)
}

func GetUserID(context contextpkg.Context) (string, error) {
	val := context.Value("userID")
	if val == nil {
		return "", ErrUserIDNotSet
	}

	id, ok := val.(string)
	if !ok {
		return "", ErrUserIDNotSet
	}

	return id, nil
}

func GetUserUUID(context contextpkg.Context) (uuidpkg.UUID, error) {
	val := context.Value("userID")
	if val == nil {
		return uuidpkg.Max, ErrUserIDNotSet
	}

	id, ok := val.(string)
	if !ok {
		return uuidpkg.Max, ErrUserIDNotSet
	}

	uuid, err := uuidpkg.Parse(id)
	if err != nil {
		return uuidpkg.Max, fmt.Errorf("failed to parse UUID: %w", err)
	}

	return uuid, nil
}
