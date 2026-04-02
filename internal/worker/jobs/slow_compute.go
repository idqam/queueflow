package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func SlowComputeHandler(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Size       int   `json:"size"`
		Iterations int   `json:"iterations"`
		Seed       int64 `json:"seed"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if p.Size <= 0 {
		p.Size = 500
	}
	if p.Iterations <= 0 {
		p.Iterations = 5
	}

	start := time.Now()
	rng := rand.New(rand.NewSource(p.Seed))

	a := make([]float64, p.Size*p.Size)
	b := make([]float64, p.Size*p.Size)
	for i := range a {
		a[i] = rng.Float64()
		b[i] = rng.Float64()
	}

	var checksum float64
	for iter := 0; iter < p.Iterations; iter++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		c := make([]float64, p.Size*p.Size)
		for i := 0; i < p.Size; i++ {
			for j := 0; j < p.Size; j++ {
				var sum float64
				for k := 0; k < p.Size; k++ {
					sum += a[i*p.Size+k] * b[k*p.Size+j]
				}
				c[i*p.Size+j] = sum
			}
		}
		for _, v := range c {
			checksum += v
		}
		a = c
	}

	result := map[string]any{
		"result_checksum": math.Round(checksum*1000) / 1000,
		"duration_ms":     time.Since(start).Milliseconds(),
	}
	return json.Marshal(result)
}
