package services

import "context"

// DeploymentAgentListener intarface to implement stream processing of agent events
type DeploymentAgentListener interface {
	ListenAsync(ctx context.Context) error
	Done() <-chan bool
}
