package filter

import (
	"encoding/json"
	"strings"
)

// IsDeploymentFinishedFilter deployment finished
var IsDeploymentFinishedFilter UnaryPredicateFilterFunc = isDeploymentFinishedFilter

func isDeploymentFinishedFilter(data []byte) bool {
	evt := struct {
		EventName string `json:"eventName"`
		Finished  bool   `json:"finished"`
	}{}

	if err := json.Unmarshal(data, &evt); err != nil {
		return false
	}

	return strings.ToLower(evt.EventName) == "deploymentevent" && evt.Finished
}
