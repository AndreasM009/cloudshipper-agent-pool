package filter

import (
	"context"
	"sync"

	"github.com/andreasM009/nats-library/channel"
	"github.com/nats-io/stan.go"
)

// Processor processes agent's event steam
type Processor struct {
	natsStreamingChannel *channel.NatsStreamingChannel
	forwarder            *Forwarder
	filter               UnaryPredicateFilter
	done                 chan bool
}

// NewProcessor new instance
func NewProcessor(channel *channel.NatsStreamingChannel, forwarder *Forwarder, filter UnaryPredicateFilter) *Processor {
	return &Processor{
		natsStreamingChannel: channel,
		done:                 make(chan bool, 1),
		forwarder:            forwarder,
		filter:               filter,
	}
}

// ProcessAsync processes stream async
func (processor *Processor) ProcessAsync(ctx context.Context) error {
	var wg sync.WaitGroup
	var mustStop bool = false
	var isJobRunning bool = false
	var mtx sync.Mutex

	_, err := processor.natsStreamingChannel.SnatNativeConnection.QueueSubscribe(processor.natsStreamingChannel.NatsPublishName, "agentpoolfilters", func(msg *stan.Msg) {
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

		if processor.filter.Filter(msg.Data) {
			if err := processor.forwarder.Forward(msg.Data); err == nil {
				msg.Ack()
			}
		} else {
			msg.Ack()
		}
	}, stan.DurableName("agent"), stan.SetManualAckMode(), stan.MaxInflight(5))

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

		processor.done <- true

	}()

	return err
}

// Done has processor finished
func (processor *Processor) Done() <-chan bool {
	return processor.done
}
