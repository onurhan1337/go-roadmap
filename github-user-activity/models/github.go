package models

import (
	"encoding/json"
	"fmt"
)

type EventType string
type OutputFormat string

const (
	SimpleFormat OutputFormat = "simple"
	JSONFormat   OutputFormat = "json"
	TableFormat  OutputFormat = "table"
)

const (
	PushEvent          EventType = "PushEvent"
	IssuesEvent        EventType = "IssuesEvent"
	WatchEvent         EventType = "WatchEvent"
	ForkEvent          EventType = "ForkEvent"
	CreateEvent        EventType = "CreateEvent"
	DeleteEvent        EventType = "DeleteEvent"
	PullRequestEvent   EventType = "PullRequestEvent"
	ReleaseEvent       EventType = "ReleaseEvent"
	IssueCommentEvent  EventType = "IssueCommentEvent"
	CommitCommentEvent EventType = "CommitCommentEvent"
	PublicEvent        EventType = "PublicEvent"
	MemberEvent        EventType = "MemberEvent"
	UnknownEvent       EventType = "UnknownEvent"
)

type GithubEvent struct {
	Type EventType `json:"type"`
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
}

type UsernameRules struct {
	MaxLength    int
	MinLength    int
	AllowedChars string
}

var GithubRules = UsernameRules{
	MaxLength:    39,
	MinLength:    1,
	AllowedChars: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-",
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
	case string(ForkEvent):
		*e = ForkEvent
	case string(CreateEvent):
		*e = CreateEvent
	case string(DeleteEvent):
		*e = DeleteEvent
	case string(PullRequestEvent):
		*e = PullRequestEvent
	case string(ReleaseEvent):
		*e = ReleaseEvent
	case string(IssueCommentEvent):
		*e = IssueCommentEvent
	case string(CommitCommentEvent):
		*e = CommitCommentEvent
	case string(PublicEvent):
		*e = PublicEvent
	case string(MemberEvent):
		*e = MemberEvent
	default:
		*e = UnknownEvent
	}

	return nil
}
