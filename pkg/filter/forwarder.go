package filter

import (
	"log"

	"github.com/andreasM009/nats-library/channel"
)

// Forwarder forwards filtered events to PoolManager
type Forwarder struct {
	natsStreamingChannel *channel.NatsStreamingChannel
}

// NewForwarder new instance
func NewForwarder(channel *channel.NatsStreamingChannel) *Forwarder {
	return &Forwarder{
		natsStreamingChannel: channel,
	}
}

// Forward forwards filtered data to PoolManager
func (f *Forwarder) Forward(data []byte) error {
	log.Println("forwarding event: ", string(data))
	return f.natsStreamingChannel.SnatNativeConnection.Publish(f.natsStreamingChannel.NatsPublishName, data)
}
