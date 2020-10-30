package filter

import (
	"encoding/json"
	"strings"
)

// IsDeploymentStartedFilter deployment finished
var IsDeploymentStartedFilter UnaryPredicateFilterFunc = isDeploymentStartedFilter

func isDeploymentStartedFilter(data []byte) bool {
	evt := struct {
		EventName string `json:"eventName"`
		Started   bool   `json:"started"`
	}{}

	if err := json.Unmarshal(data, &evt); err != nil {
		return false
	}

	return strings.ToLower(evt.EventName) == "deploymentevent" && evt.Started
}
