package execution

import (
	"errors"
	"strings"

	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

type AdapterRegistry struct {
	adapters map[string]contracts.ExecutionAdapter
}

func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{adapters: make(map[string]contracts.ExecutionAdapter)}
}

func (r *AdapterRegistry) Register(provider string, adapter contracts.ExecutionAdapter) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || adapter == nil {
		return
	}
	r.adapters[provider] = adapter
}

func (r *AdapterRegistry) Resolve(provider string) (contracts.ExecutionAdapter, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	adapter, ok := r.adapters[provider]
	if !ok {
		return nil, errors.New("unsupported execution provider")
	}
	return adapter, nil
}
