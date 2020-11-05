package services

import (
	"testing"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories/inmemory"

	"github.com/stretchr/testify/assert"
)

func TestEnqueueDeployment(t *testing.T) {
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	svc := NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, nil)

	dep := &domain.Deployment{
		DeploymentName: "d1",
		DefinitionID:   "def1",
		TenantID:       "1",
	}

	yaml := []byte("MyYaml")
	parameters := []byte("MyParameters")

	err := svc.EnqueueDeployment(dep, yaml, parameters)
	assert.Nil(t, err)

	state, err := stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, dep.State)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{dep.DeploymentID}, state.DeploymentIDs))
}

func TestEnqueueTwoDeployments(t *testing.T) {
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	svc := NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, nil)

	depOne := &domain.Deployment{
		DeploymentName: "d1",
		DefinitionID:   "def1",
		TenantID:       "1",
	}

	yaml := []byte("MyYaml")
	parameters := []byte("MyParameters")

	err := svc.EnqueueDeployment(depOne, yaml, parameters)
	assert.Nil(t, err)

	state, err := stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))

	depTwo := &domain.Deployment{
		DeploymentName: "d2",
		DefinitionID:   "def2",
		TenantID:       "1",
	}

	err = svc.EnqueueDeployment(depTwo, yaml, parameters)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	// still only first deployment is running
	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))
}

func TestEnqueueTwoDeploymentsWithMaxParallel2(t *testing.T) {
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	tenant, err := tenantRepo.Get("1")
	assert.Nil(t, err)

	tenant.MaxParallelJobs = 2
	_, err = tenantRepo.Update(tenant)
	assert.Nil(t, err)

	svc := NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, nil)

	depOne := &domain.Deployment{
		DeploymentName: "d1",
		DefinitionID:   "def1",
		TenantID:       "1",
	}

	yaml := []byte("MyYaml")
	parameters := []byte("MyParameters")

	err = svc.EnqueueDeployment(depOne, yaml, parameters)
	assert.Nil(t, err)

	state, err := stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))

	depTwo := &domain.Deployment{
		DeploymentName: "d2",
		DefinitionID:   "def2",
		TenantID:       "1",
	}

	err = svc.EnqueueDeployment(depTwo, yaml, parameters)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depTwo.State)

	// both are running
	assert.Equal(t, 2, state.RunningJobs)
	assert.Equal(t, 2, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID, depTwo.DeploymentID}, state.DeploymentIDs))
}

func TestEnqueueThreeDeploymentsWithMaxParallel2(t *testing.T) {
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	tenant, err := tenantRepo.Get("1")
	assert.Nil(t, err)

	tenant.MaxParallelJobs = 2
	_, err = tenantRepo.Update(tenant)
	assert.Nil(t, err)

	svc := NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, nil)

	depOne := &domain.Deployment{
		DeploymentName: "d1",
		DefinitionID:   "def1",
		TenantID:       "1",
	}

	yaml := []byte("MyYaml")
	parameters := []byte("MyParameters")

	err = svc.EnqueueDeployment(depOne, yaml, parameters)
	assert.Nil(t, err)

	state, err := stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))

	depTwo := &domain.Deployment{
		DeploymentName: "d2",
		DefinitionID:   "def2",
		TenantID:       "1",
	}

	err = svc.EnqueueDeployment(depTwo, yaml, parameters)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depTwo.State)

	// still only first deployment is running
	assert.Equal(t, 2, state.RunningJobs)
	assert.Equal(t, 2, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID, depTwo.DeploymentID}, state.DeploymentIDs))

	depThree := &domain.Deployment{
		DeploymentName: "d3",
		DefinitionID:   "def3",
		TenantID:       "1",
	}

	err = svc.EnqueueDeployment(depThree, yaml, parameters)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depThree.State)

	// still only first and second deployment are running
	assert.Equal(t, 2, state.RunningJobs)
	assert.Equal(t, 2, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID, depTwo.DeploymentID}, state.DeploymentIDs))
}

func TestEnqueueTwoDeploymentsMaxParallel2AndOneFinished(t *testing.T) {
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	svc := NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, nil)

	depOne := &domain.Deployment{
		DeploymentName: "d1",
		DefinitionID:   "def1",
		TenantID:       "1",
	}

	yaml := []byte("MyYaml")
	parameters := []byte("MyParameters")

	err := svc.EnqueueDeployment(depOne, yaml, parameters)
	assert.Nil(t, err)

	state, err := stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))

	depTwo := &domain.Deployment{
		DeploymentName: "d2",
		DefinitionID:   "def2",
		TenantID:       "1",
	}

	err = svc.EnqueueDeployment(depTwo, yaml, parameters)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, domain.Queued, depOne.State)

	// still only first deployment is running
	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depOne.DeploymentID}, state.DeploymentIDs))

	err = svc.DeploymentFinished("1", depOne.DeploymentID, 0)
	assert.Nil(t, err)

	state, err = stateRepo.Get("1")
	assert.Nil(t, err)
	assert.NotNil(t, state)

	assert.Equal(t, 1, state.RunningJobs)
	assert.Equal(t, 1, len(state.DeploymentIDs))
	assert.True(t, assert.ObjectsAreEqualValues([]string{depTwo.DeploymentID}, state.DeploymentIDs))

	_, err = depRepo.Get("1", depOne.DeploymentID)
	assert.NotNil(t, err)
}
