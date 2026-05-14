package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type ExecutionHost struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Adapter   string          `json:"adapter"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

var (
	ErrHostNotFound = errors.New("host not found")
)
