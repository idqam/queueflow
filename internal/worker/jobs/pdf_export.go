package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func PDFExportHandler(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Template     string `json:"template"`
		OutputBucket string `json:"output_bucket"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	sleepDuration := time.Second + time.Duration(rand.Intn(2000))*time.Millisecond
	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	bucket := p.OutputBucket
	if bucket == "" {
		bucket = "queueflow-outputs"
	}

	result := map[string]any{
		"output_url": fmt.Sprintf("s3://%s/export-%d.pdf", bucket, rand.Int63()),
		"pages":      rand.Intn(12) + 1,
	}
	return json.Marshal(result)
}
