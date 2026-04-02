package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func EmailSendHandler(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p struct {
		To       string `json:"to"`
		Template string `json:"template"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	sleepDuration := 5*time.Millisecond + time.Duration(rand.Intn(45))*time.Millisecond
	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	result := map[string]string{
		"message_id":   fmt.Sprintf("msg-%d", rand.Int63()),
		"delivered_at": time.Now().UTC().Format(time.RFC3339),
	}
	return json.Marshal(result)
}
