package execution

import (
	"errors"
	"strings"

	"github.com/UtopikCode/quickspaces-control-plane/execution/adapters"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
)

func NewAdapter(provider string) (contracts.ExecutionAdapter, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "aws":
		return adapters.NewAWSExecutionAdapter(), nil
	case "truenas":
		return adapters.NewLocalExecutionAdapter(), nil
	default:
		return nil, errors.New("unsupported execution provider")
	}
}
