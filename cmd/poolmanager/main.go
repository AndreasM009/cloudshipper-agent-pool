package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories/azure"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/http"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/services/nats"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/services"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories/inmemory"
	"github.com/andreasM009/nats-library/channel"
)

func main() {
	runtime := poolmanager.NewPoolManagerRuntime()
	runtime.FromFlags()

	// cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// NATS
	// Connection to Nats Server
	natsConnection, err := channel.NewNatsConnection(
		runtime.NatsConnectionStrings, runtime.NatsConnectionName)
	if err != nil {
		log.Panic(err)
	}

	if err := channel.GetNatsConnectionPoolInstance().Add(
		runtime.NatsConnectionName, natsConnection); err != nil {
		log.Panic(err)
	}

	// connection to Nats Streaming server
	snatsConnection, err := channel.NewNatsStreamingConnectionWithPooledConnection(
		runtime.NatsConnectionName, runtime.NatsStreamingClusterID, runtime.NastStreamingClientID)

	if err != nil {
		log.Panic(err)
	}

	if err := channel.GetNatsStreamingConnectionPoolInstance().Add(
		runtime.NatsStreamingClusterID, runtime.NastStreamingClientID, snatsConnection); err != nil {
		log.Panic(err)
	}

	// PoolManager channel
	poolMngrChannel, err := channel.NewNatsStreamingChannelFromPool(runtime.PoolManagerChannelName, runtime.NatsStreamingClusterID, runtime.NastStreamingClientID)
	if err != nil {
		log.Panic(err)
	}

	// Input agent steam
	agentChannel, err := channel.NewNatsStreamingChannelFromPool(runtime.AgentEventChannelName, runtime.NatsStreamingClusterID, runtime.NastStreamingClientID)
	if err != nil {
		log.Panic(err)
	}

	// repositories, default inMemory
	tenantRepo := inmemory.NewInMemoryTenantRepository()
	defRepo := inmemory.NewInMemoryDefinitionRepository()
	depRepo := inmemory.NewInMemoryDeploymentRepository()
	stateRepo := inmemory.NewInMemoryQueueStateRepository()

	// run in mode Azure?
	if runtime.IsModeAzure {
		defRepo, err = azure.NewAzureDefinitionRepository(runtime.StorageContainer)
		if err != nil {
			log.Panic(err)
		}

		depRepo, err = azure.NewAzureDeploymentRepository()
		if err != nil {
			log.Panic(err)
		}

		tenantRepo, err = azure.NewAzureTenantRepository()
		if err != nil {
			log.Panic(err)
		}

		stateRepo, err = azure.NewAzureQueueStateRepository()
		if err != nil {
			log.Panic(err)
		}
	}

	// enqeue deployments
	enqService := nats.NewEnqueueDeploymentService(agentChannel)
	// deploymentService
	deploymentService := services.NewDeploymentService(stateRepo, depRepo, defRepo, tenantRepo, enqService)
	// listener
	listener := nats.NewNatsDeploymentAgentListener(poolMngrChannel, deploymentService)

	if err := listener.ListenAsync(ctx); err != nil {
		log.Panic(err)
	}

	// API Server
	apiserver := http.NewHTTPServer(runtime.APIServerPort, deploymentService, depRepo)
	apiserver.ListentAndServeAsync(ctx)

	signalchannel := make(chan os.Signal, 1)
	signal.Notify(signalchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case sig := <-signalchannel:
		log.Println(fmt.Sprintf("Received signal: %d, shutting down", sig))
		cancel()
		<-listener.Done()
		<-apiserver.Done()
	case <-listener.Done():
		log.Println("processor ended unexpected")
	case <-apiserver.Done():
		log.Println("apiserver ended unexpected")
	}
}
