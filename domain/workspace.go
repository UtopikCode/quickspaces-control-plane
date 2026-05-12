package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type ExecutionProfile = json.RawMessage

type Workspace struct {
	ID               string           `json:"id"`
	Repo             string           `json:"repo"`
	Owner            string           `json:"owner"`
	Ref              string           `json:"ref"`
	DesiredState     string           `json:"desiredState"`
	ActualState      string           `json:"actualState"`
	ExecutionProfile ExecutionProfile `json:"executionProfile"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
)
