package execution

import (
	"encoding/json"
	"errors"
	"strings"

	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type AdapterFactory func(hostConfig json.RawMessage) (contracts.ExecutionAdapter, error)

type AdapterResolver interface {
	Resolve(provider string, hostConfig json.RawMessage) (contracts.ExecutionAdapter, error)
}

type AdapterRegistry struct {
	factories map[string]AdapterFactory
}

func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{factories: make(map[string]AdapterFactory)}
}

func (r *AdapterRegistry) Register(provider string, factory AdapterFactory) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || factory == nil {
		return
	}
	r.factories[provider] = factory
}

func (r *AdapterRegistry) Resolve(provider string, hostConfig json.RawMessage) (contracts.ExecutionAdapter, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	factory, ok := r.factories[provider]
	if !ok {
		return nil, errors.New("unsupported execution provider")
	}
	return factory(hostConfig)
}
