package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	base "github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/services"
	"github.com/andreasM009/nats-library/channel"
	snatsio "github.com/nats-io/stan.go"
)

// DeploymentAgentListener processes filtered agent stream
type deploymentAgentListener struct {
	filteredChannel   *channel.NatsStreamingChannel
	done              chan bool
	deploymentService *base.DeploymentService
}

// NewNatsDeploymentAgentListener new instance
func NewNatsDeploymentAgentListener(stream *channel.NatsStreamingChannel, deploymentService *base.DeploymentService) base.DeploymentAgentListener {
	return &deploymentAgentListener{
		filteredChannel:   stream,
		done:              make(chan bool, 1),
		deploymentService: deploymentService,
	}
}

func (listener *deploymentAgentListener) ListenAsync(ctx context.Context) error {
	var wg sync.WaitGroup
	var mustStop bool = false
	var isJobRunning bool = false
	var mtx sync.Mutex

	_, err := listener.filteredChannel.SnatNativeConnection.QueueSubscribe(listener.filteredChannel.NatsPublishName, "poolmanager", func(msg *snatsio.Msg) {
		// check if listener must stop or not
		mtx.Lock()

		if mustStop {
			mtx.Unlock()
			return
		}

		isJobRunning = true
		wg.Add(1)
		mtx.Unlock()

		defer func() {
			mtx.Lock()
			defer mtx.Unlock()
			isJobRunning = false
		}()

		// now we can run the job, and notify waiter when it is finished
		defer wg.Done()

		// the event we expect to receive
		evt := struct {
			EventName    string `json:"eventName"`
			Finished     bool   `json:"finished"`
			Started      bool   `json:"started"`
			TenantID     string `json:"tenantId"`
			DeploymentID string `json:"deplyomentId"`
		}{}

		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Println(err)
			return
		}

		// only deploymentEvent
		if strings.ToLower(evt.EventName) != "deploymentevent" {
			msg.Ack()
			return
		}

		if evt.Started {
			if err := listener.deploymentService.DeploymentStarted(evt.TenantID, evt.DeploymentID); err != nil {
				log.Println(err)
				msg.Ack()
				return
			}

			msg.Ack()
		} else if evt.Finished {
			if err := listener.deploymentService.DeploymentFinished(evt.TenantID, evt.DeploymentID, 0); err != nil {
				log.Println(err)
				msg.Ack()
				return
			}

			msg.Ack()
		} else {
			// unknown event
			log.Println(fmt.Sprintf("Received unknown event: %s", evt.EventName))
			msg.Ack()
		}

	}, snatsio.DurableName("poolmanager"), snatsio.SetManualAckMode(), snatsio.MaxInflight(1), snatsio.AckWait(time.Second*60))

	if err != nil {
		return err
	}

	// wait until poolmanager must stop
	go func() {
		<-ctx.Done()
		mtx.Lock()
		mustStop = true
		if isJobRunning {
			mtx.Unlock()
			wg.Wait()
		} else {
			mtx.Unlock()
		}

		listener.done <- true
	}()

	return nil
}

func (listener *deploymentAgentListener) Done() <-chan bool {
	return listener.done
}
