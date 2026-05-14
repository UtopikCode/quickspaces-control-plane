package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

type HostRepository interface {
	Create(ctx context.Context, host *domain.ExecutionHost) error
	GetByID(ctx context.Context, id string) (*domain.ExecutionHost, error)
	List(ctx context.Context) ([]*domain.ExecutionHost, error)
}

type HostService struct {
	repo HostRepository
}

func NewHostService(repo HostRepository) *HostService {
	return &HostService{repo: repo}
}

type CreateHostRequest struct {
	Name    string                  `json:"name"`
	Adapter string                  `json:"adapter"`
	Config  domain.ExecutionProfile `json:"config"`
}

func (s *HostService) CreateHost(ctx context.Context, request CreateHostRequest) (*domain.ExecutionHost, error) {
	if request.Name == "" || request.Adapter == "" {
		return nil, errors.New("host name and adapter are required")
	}

	now := time.Now().UTC()
	host := &domain.ExecutionHost{
		ID:        generateID(),
		Name:      request.Name,
		Adapter:   strings.ToLower(strings.TrimSpace(request.Adapter)),
		Config:    request.Config,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, host); err != nil {
		return nil, err
	}

	return host, nil
}

func (s *HostService) ListHosts(ctx context.Context) ([]*domain.ExecutionHost, error) {
	return s.repo.List(ctx)
}

func (s *HostService) GetHost(ctx context.Context, id string) (*domain.ExecutionHost, error) {
	return s.repo.GetByID(ctx, id)
}
