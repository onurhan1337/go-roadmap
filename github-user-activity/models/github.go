package models

import "encoding/json"

type EventType string

const (
	PushEvent   EventType = "PushEvent"
	IssuesEvent EventType = "IssuesEvent"
	WatchEvent  EventType = "WatchEvent"
	OtherEvent  EventType = "OtherEvent"
)

type GithubEvent struct {
	Type EventType `json:"type"`
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
}

func (e *EventType) UnmarshalJSON(data []byte) error {
	var typeString string
	if err := json.Unmarshal(data, &typeString); err != nil {
		return err
	}

	switch typeString {
	case "PushEvent":
		*e = PushEvent
	case "IssuesEvent":
		*e = IssuesEvent
	case "WatchEvent":
		*e = WatchEvent
	case "OtherEvent":
		*e = OtherEvent
	}

	return nil
}
