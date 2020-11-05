package nats

import (
	"encoding/json"
	"log"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	base "github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/services"
	"github.com/andreasM009/nats-library/channel"
	"gopkg.in/yaml.v2"
)

// EnqueueDeploymentService enqueues jobs for agents
type enqueueDeploymentService struct {
	agentChannel *channel.NatsStreamingChannel
}

// NewEnqueueDeploymentService new instance
func NewEnqueueDeploymentService(channel *channel.NatsStreamingChannel) base.EnqueueDeploymentService {
	return &enqueueDeploymentService{
		agentChannel: channel,
	}
}

// Enqueue implements EnqueueDeploymentService
func (svc *enqueueDeploymentService) Enqueue(deployment *domain.Deployment, yamldef, parameters []byte) error {
	// todo: split up creation of data to send and the sending of data through channel
	p := map[string]string{}
	if err := yaml.Unmarshal(parameters, &p); err != nil {
		log.Println(err)
	}

	evt := base.EnqueueDeploymentEvent{
		ID:             deployment.DeploymentID,
		DefinitionID:   deployment.DefinitionID,
		DeploymentName: deployment.DeploymentName,
		TenantID:       deployment.TenantID,
		Yaml:           string(yamldef),
		Parameters:     p,
		LiveStreamName: deployment.LiveStream,
	}

	data, err := json.Marshal(&evt)
	if err != nil {
		return err
	}

	sub := svc.agentChannel.NatsPublishName
	if err := svc.agentChannel.SnatNativeConnection.Publish(sub, data); err != nil {
		return err
	}

	return nil
}
