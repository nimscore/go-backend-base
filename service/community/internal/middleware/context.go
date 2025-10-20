package middleware

import (
	contextpkg "context"
	"errors"
)

var ErrUserIDNotSet = errors.New("user_id not set")

// Исключает коллизии ключей в контексте
type typedKey struct{ name string }

var (
	userIDKey = typedKey{name: "userID"}
)

func SetUserID(context contextpkg.Context, id string) contextpkg.Context {
	return contextpkg.WithValue(context, userIDKey, id)
}

func GetUserID(context contextpkg.Context) (string, error) {
	if val := context.Value(userIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id, nil
		}
	}

	return "", ErrUserIDNotSet
}
