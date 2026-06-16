package webhook

import (
	"encoding/json"
	"os"
)

type Subscriber struct {
	URL              string   `json:"url"`
	Secret           string   `json:"secret"`
	EventTypes       []string `json:"event_types"`
	WorkflowDefNames []string `json:"workflow_def_names"`
}

func LoadRegistry(path string) ([]Subscriber, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var subscribers []Subscriber
	if err := json.Unmarshal(data, &subscribers); err != nil {
		return nil, err
	}
	return subscribers, nil
}

func (s *Subscriber) Matches(eventType, defName string) bool {
	if !containsString(s.EventTypes, eventType) {
		return false
	}
	if len(s.WorkflowDefNames) == 0 {
		return true
	}
	return containsString(s.WorkflowDefNames, defName)
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
