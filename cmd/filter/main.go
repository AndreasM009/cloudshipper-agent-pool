package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/filter"
	"github.com/andreasM009/nats-library/channel"
)

func main() {
	runtime := filter.NewRuntime()
	runtime.FromFlags()

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
	fwdChannel, err := channel.NewNatsStreamingChannelFromPool(runtime.PoolManagerChannelName, runtime.NatsStreamingClusterID, runtime.NastStreamingClientID)
	if err != nil {
		log.Panic(err)
	}
	// forwarder
	forwarder := filter.NewForwarder(fwdChannel)

	// Input agent steam
	agentChannel, err := channel.NewNatsStreamingChannelFromPool(runtime.AgentEventChannelName, runtime.NatsStreamingClusterID, runtime.NastStreamingClientID)
	if err != nil {
		log.Panic(err)
	}

	// set filters
	or := filter.Or(filter.IsDeploymentFinishedFilter, filter.IsDeploymentStartedFilter)

	processor := filter.NewProcessor(agentChannel, forwarder, or)

	if err := processor.ProcessAsync(ctx); err != nil {
		log.Panic(err)
	}

	signalchannel := make(chan os.Signal, 1)
	signal.Notify(signalchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case sig := <-signalchannel:
		log.Println(fmt.Sprintf("Received signal: %d, shutting down", sig))
		cancel()
		<-processor.Done()
	case <-processor.Done():
		log.Println("processor ended")
	}
}
