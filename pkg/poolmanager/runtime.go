package poolmanager

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories/azure"
	"gopkg.in/yaml.v2"

	"github.com/google/uuid"
)

var thePoolManagerRuntime *Runtime

var usageStr = `
Usage: runner [options]
Options:
	-m <mode local | kuberenetes>						Mode, run locally or in Kuberenets
	-storage-mode <mode inmemory | azure> 				Mode in memory or Azure Storages
	-p <port>											API Server port
	-s <url>											NATS server URL(s) (separated by comma)
	-cluster-id <cluster id>							NATS streaming cluster id
	-agent-event-channel <agent streaming channel name>	NATS agent event channel
	-pool-manager-channel <pool manager channel name> 	NATS subscription to publish all events
	-config-filepath <path to file containg config> 	Load secrets from config file
`
var mode, serverURLs, clusterID, agentChannelName, poolManagerChannelName, port, storageMode, configFilePath string

// Runtime PoolManager
type Runtime struct {
	NatsConnectionName     string
	NatsConnectionStrings  []string
	NatsStreamingClusterID string
	NastStreamingClientID  string
	AgentEventChannelName  string
	PoolManagerChannelName string
	APIServerPort          string
	ConfigFilePath         string
	IsModeAzure            bool
	IsModeInMemory         bool
	StorageAccountName     string
	StorageAccountKey      string
	StorageContainer       string
}

// NewPoolManagerRuntime instance
func NewPoolManagerRuntime() *Runtime {
	if nil == thePoolManagerRuntime {
		thePoolManagerRuntime = &Runtime{}
		flag.StringVar(&mode, "m", "", "")
		flag.StringVar(&storageMode, "storage-mode", "", "")
		flag.StringVar(&serverURLs, "s", "", "")
		flag.StringVar(&clusterID, "cluster-id", "", "")
		flag.StringVar(&agentChannelName, "agent-channel", "", "")
		flag.StringVar(&poolManagerChannelName, "pool-manager-channel", "", "")
		flag.StringVar(&port, "p", "", "")
		flag.StringVar(&configFilePath, "config-filepath", "", "")
	}

	return thePoolManagerRuntime
}

// FromFlags parse config from flags
func (runtime *Runtime) FromFlags() {
	flag.Usage = usage
	flag.Parse()

	if mode == "" {
		log.Println("No runtime mode -m speicfied, either local or kuberenetes")
		flag.Usage()
	}

	if storageMode == "" {
		log.Println("Invalid storgae mode: inmemory | azure")
		flag.Usage()
	}

	if strings.ToLower(storageMode) == "azure" {
		if configFilePath == "" {
			log.Println("No config file path specified to run in mode azure")
			flag.Usage()
		}

		runtime.IsModeAzure = true
	} else {
		runtime.IsModeInMemory = true
	}

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
		flag.Usage()
	}

	if port == "" {
		log.Println("No api server port specified")
		flag.Usage()
	}

	clientID := fmt.Sprintf("cs-agpoolmngr-%s", uuid.New().String())

	runtime.NatsConnectionName = "agpoolmngr"
	runtime.NatsConnectionStrings = strings.Split(serverURLs, ",")
	runtime.NatsStreamingClusterID = clusterID
	runtime.PoolManagerChannelName = poolManagerChannelName
	runtime.NastStreamingClientID = clientID
	runtime.AgentEventChannelName = agentChannelName
	runtime.APIServerPort = port
	runtime.ConfigFilePath = configFilePath

	if runtime.IsModeAzure {
		err := runtime.loadConfigFiles()
		if err != nil {
			log.Panic(err)
		}

		if err := azure.SetStorageAccount(runtime.StorageAccountName, runtime.StorageAccountKey); err != nil {
			log.Panic(fmt.Sprintf("Azure storage account error: %s", err))
		}
	}
}

func usage() {
	log.Fatalf(usageStr)
}

func (runtime *Runtime) loadConfigFiles() error {

	config := struct {
		StorageAccountName   string `yaml:"storageAccountName"`
		StorageAccountKey    string `yaml:"storageAccountKey"`
		StorageBlobContainer string `yaml:"storageBlobContainer"`
	}{}

	data, err := ioutil.ReadFile(runtime.ConfigFilePath)
	if err != nil {
		return err
	}

	// running in kubernetes?
	// if strings.ToLower(mode) == "kubernetes" {
	// 	data, err = base64.StdEncoding.DecodeString(string(data))
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	runtime.StorageAccountKey = config.StorageAccountKey
	runtime.StorageAccountName = config.StorageAccountName
	runtime.StorageContainer = config.StorageBlobContainer

	return nil
}
