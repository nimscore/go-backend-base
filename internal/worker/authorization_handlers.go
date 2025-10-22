package worker

import (
	"encoding/json"
	"fmt"

	eventpkg "github.com/stormhead-org/backend/internal/event"
)

func (this *Worker) AuthorizationLoginHandler(data []byte) error {
	var message eventpkg.AuthorizationLoginMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}

	return nil
}
