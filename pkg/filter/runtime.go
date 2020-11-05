package filter

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

var theRuntimeInstance *Runtime

var usageStr = `
Usage: runner [options]
Options:
	-s <url>											NATS server URL(s) (separated by comma)
	-cluster-id <cluster id>							NATS streaming cluster id
	-agent-event-channel <agent streaming channel name>	NATS agent event channel
	-pool-manager-channel <pool manager channel name> 	NATS subscription to publish all events
`
var serverURLs, clusterID, agentChannelName, poolManagerChannelName string

// Runtime holds runtime information
type Runtime struct {
	NatsConnectionName     string
	NatsConnectionStrings  []string
	NatsStreamingClusterID string
	NastStreamingClientID  string
	AgentEventChannelName  string
	PoolManagerChannelName string
}

// NewRuntime creates or gets instance
func NewRuntime() *Runtime {
	if nil == theRuntimeInstance {
		theRuntimeInstance = &Runtime{}
		flag.StringVar(&serverURLs, "s", "", "")
		flag.StringVar(&clusterID, "cluster-id", "", "")
		flag.StringVar(&agentChannelName, "agent-channel", "", "")
		flag.StringVar(&poolManagerChannelName, "pool-manager-channel", "", "")
	}

	return theRuntimeInstance
}

// FromFlags runtime from flags
func (runtime *Runtime) FromFlags() {
	flag.Usage = usage
	flag.Parse()

	if serverURLs == "" {
		log.Println("No NATS server URL specified")
		flag.Usage()
	}

	if clusterID == "" {
		log.Println("No cluster ID specified")
		flag.Usage()
	}

	if agentChannelName == "" {
		log.Println("No agent channel name specified for input")
		flag.Usage()
	}

	if poolManagerChannelName == "" {
		log.Println("No pool manager channel specified")
	}

	clientID := fmt.Sprintf("cs-agpool-filter-%s", uuid.New().String())

	runtime.NatsConnectionName = "agentfilter"
	runtime.NatsConnectionStrings = strings.Split(serverURLs, ",")
	runtime.NatsStreamingClusterID = clusterID
	runtime.PoolManagerChannelName = poolManagerChannelName
	runtime.NastStreamingClientID = clientID
	runtime.AgentEventChannelName = agentChannelName
}

func usage() {
	log.Fatalf(usageStr)
}
