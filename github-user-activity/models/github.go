package models

import (
	"encoding/json"
	"fmt"
)

type EventType string

const (
	PushEvent    EventType = "PushEvent"
	IssuesEvent  EventType = "IssuesEvent"
	WatchEvent   EventType = "WatchEvent"
	UnknownEvent EventType = "UnknownEvent"
)

type GithubEvent struct {
	Type EventType `json:"type"`
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
}

func (e EventType) String() string {
	return string(e)
}

func (e *EventType) UnmarshalJSON(data []byte) error {
	var typeString string
	if err := json.Unmarshal(data, &typeString); err != nil {
		return fmt.Errorf("failed to unmarshal event type: %w", err)
	}

	switch typeString {
	case string(PushEvent):
		*e = PushEvent
	case string(IssuesEvent):
		*e = IssuesEvent
	case string(WatchEvent):
		*e = WatchEvent
	default:
		*e = UnknownEvent
	}

	return nil
}
