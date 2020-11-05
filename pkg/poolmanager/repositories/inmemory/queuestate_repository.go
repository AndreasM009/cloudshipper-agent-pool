package inmemory

import (
	"errors"
	"strconv"
	"sync"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

type inMemQueueStateRepository struct {
	states map[string]*domain.QueueState
	lock   sync.Mutex
}

// NewInMemoryQueueStateRepository new instance
func NewInMemoryQueueStateRepository() repositories.QueueStateRepository {
	return &inMemQueueStateRepository{
		states: make(map[string]*domain.QueueState),
	}
}

func (repo *inMemQueueStateRepository) Update(state *domain.QueueState) (*domain.QueueState, error) {
	if state.TenantID == "" {
		return nil, errors.New("tenant id can not be empty")
	}

	repo.lock.Lock()
	defer repo.lock.Unlock()

	if v, ok := repo.states[state.TenantID]; ok {
		if v.ETag != state.ETag {
			return nil, repositories.NewPreconditionFailedError(errors.New("ETag missmatch"))
		}
	}

	val, err := strconv.Atoi(state.ETag)
	if err != nil {
		return nil, err
	}

	val++
	state.ETag = strconv.Itoa(val)

	repo.states[state.TenantID] = state
	return state, nil
}

func (repo *inMemQueueStateRepository) Get(tenantID string) (*domain.QueueState, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if val, ok := repo.states[tenantID]; ok {
		return val, nil
	}

	s := &domain.QueueState{
		TenantID:    tenantID,
		RunningJobs: 0,
		ETag:        "1",
	}

	repo.states[tenantID] = s
	return s, nil
}
