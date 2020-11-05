package inmemory

import (
	"fmt"
	"sync"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

type inMemoryDefinitionRepository struct {
	yaml       map[string][]byte
	parameters map[string][]byte
	lock       sync.Mutex
}

// NewInMemoryDefinitionRepository new instance
func NewInMemoryDefinitionRepository() repositories.DefinitionRepository {
	return &inMemoryDefinitionRepository{
		yaml:       make(map[string][]byte),
		parameters: make(map[string][]byte),
	}
}

func (repo *inMemoryDefinitionRepository) Add(deploymentID string, yaml []byte, parameters []byte) error {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.yaml[deploymentID] = yaml
	repo.parameters[deploymentID] = parameters
	return nil
}

func (repo *inMemoryDefinitionRepository) Get(deploymentID string) ([]byte, []byte, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	var yaml, parameters []byte

	if val, ok := repo.yaml[deploymentID]; ok {
		yaml = val
	} else {
		return nil, nil, fmt.Errorf("No yaml for deployment id %s found", deploymentID)
	}

	if val, ok := repo.parameters[deploymentID]; ok {
		parameters = val
	} else {
		return nil, nil, fmt.Errorf("No parameters for deployment id %s found", deploymentID)
	}

	return yaml, parameters, nil
}

func (repo *inMemoryDefinitionRepository) Delete(deploymentID string) error {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	delete(repo.yaml, deploymentID)
	delete(repo.parameters, deploymentID)
	return nil
}
