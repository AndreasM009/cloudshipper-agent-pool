package services

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
	"github.com/google/uuid"
)

// DeploymentService service
type DeploymentService struct {
	queueStateRepo           repositories.QueueStateRepository
	deploymentRepo           repositories.DeploymentRepository
	definitionRepo           repositories.DefinitionRepository
	tenantRepo               repositories.TenantRepository
	enqueueDeploymentService EnqueueDeploymentService
}

// NewDeploymentService new instance
func NewDeploymentService(
	queueStateRepo repositories.QueueStateRepository,
	deploymentRepo repositories.DeploymentRepository,
	definitionRepo repositories.DefinitionRepository,
	tenantRepo repositories.TenantRepository,
	enqeueDeploymentServic EnqueueDeploymentService) *DeploymentService {
	return &DeploymentService{
		queueStateRepo:           queueStateRepo,
		deploymentRepo:           deploymentRepo,
		definitionRepo:           definitionRepo,
		tenantRepo:               tenantRepo,
		enqueueDeploymentService: enqeueDeploymentServic,
	}
}

// EnqueueDeployment enqueue
func (svc *DeploymentService) EnqueueDeployment(deployment *domain.Deployment, yaml, parameters []byte) error {
	deployment.DeploymentID = uuid.New().String()
	deployment.LiveStream = fmt.Sprintf("livestream-%s", deployment.DeploymentID)
	deployment.State = domain.Queued
	deployment.Timestamp = time.Now()

	// first add yaml definition and parameters
	if err := svc.definitionRepo.Add(
		deployment.DeploymentID,
		yaml,
		parameters); err != nil {
		return err
	}

	// third add deployment
	if _, err := svc.deploymentRepo.Add(deployment); err != nil {
		return err
	}

	return svc.enqueueNext(deployment.TenantID)
}

// DeploymentStarted deployment started
func (svc *DeploymentService) DeploymentStarted(tenantID string, deploymentID string) error {
	deployment, err := svc.deploymentRepo.Get(tenantID, deploymentID)
	if err != nil {
		return err
	}

	deployment.State = domain.Running
	_, err = svc.deploymentRepo.Update(deployment)
	if err != nil {
		return err
	}

	return nil
}

// DeploymentFinished deployment finished
func (svc *DeploymentService) DeploymentFinished(tenantID string, deploymentID string, exitCode int) error {
	_, err := svc.deploymentRepo.Delete(tenantID, deploymentID)
	if err != nil {
		return err
	}

	if err := svc.definitionRepo.Delete(deploymentID); err != nil {
		log.Println(fmt.Sprintf("deployment definitions for %s not deleted", deploymentID))
	}

	for {
		// now update state
		state, err := svc.queueStateRepo.Get(tenantID)
		if err != nil {
			return err
		}

		state.RunningJobs--
		state.DeploymentIDs = removeDeployment(state.DeploymentIDs, deploymentID)
		_, err = svc.queueStateRepo.Update(state)
		if err != nil {
			if _, ok := err.(*repositories.PreconditionFailedError); ok {
				continue
			} else {
				<-time.After(time.Second * 2)
				continue
			}
		} else {
			break
		}
	}

	return svc.enqueueNext(tenantID)
}

func (svc *DeploymentService) enqueueNext(tenantID string) error {
	tenant, err := svc.tenantRepo.Get(tenantID)
	if err != nil {
		return err
	}

	for {
		// find next deployments to enqueue
		deployments, err := svc.deploymentRepo.GetAll(tenantID)
		if err != nil {
			return err
		}

		if len(deployments) == 0 {
			return nil
		}

		if len(deployments) > 1 {
			sort.SliceStable(deployments, func(i, k int) bool {
				return deployments[i].Timestamp.Before(deployments[k].Timestamp)
			})
		}

		// load current state
		state, err := svc.queueStateRepo.Get(tenant.ID)
		if err != nil {
			return err
		}

		if state.RunningJobs < tenant.MaxParallelJobs {
			var next *domain.Deployment

			// find next that is not queued already
			for _, d := range deployments {
				if !containsDeployment(state.DeploymentIDs, d.DeploymentID) {
					next = d
					break
				}
			}

			if nil == next {
				return nil
			}

			yaml, param, err := svc.definitionRepo.Get(next.DeploymentID)
			if err != nil {
				log.Println(err)
				continue
			}

			state.RunningJobs++
			state.DeploymentIDs = append(state.DeploymentIDs, next.DeploymentID)
			// try to update state and check if another instance has already updated the state
			state, err = svc.queueStateRepo.Update(state)
			if err != nil {
				if _, ok := err.(*repositories.PreconditionFailedError); ok {
					continue
				} else {
					<-time.After(time.Second * 2)
					continue
				}
			}

			if svc.enqueueDeploymentService != nil {
				if err := svc.enqueueDeploymentService.Enqueue(next, yaml, param); err != nil {
					log.Println(err)
					// todo: split up creation of data to send and the sending of data through channel
				}
			}
		} else {
			// nothing todo, wait until a running job is completed
			return nil
		}
	}
}

func containsDeployment(deploymentIds []string, id string) bool {
	for _, v := range deploymentIds {
		if v == id {
			return true
		}
	}

	return false
}

func removeDeployment(deploymentIds []string, id string) []string {
	index := -1
	for i, v := range deploymentIds {
		if v == id {
			index = i
		}
	}

	if index == -1 {
		return deploymentIds
	}

	res := append(deploymentIds[:index], deploymentIds[index+1:]...)
	return res
}
