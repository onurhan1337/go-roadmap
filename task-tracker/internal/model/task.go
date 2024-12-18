package model

import "time"

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TaskList struct {
	Tasks []Task   `json:"tasks"`
	Tags  []string `json:"tags"` // Available tags in the system
}

const (
	StatusTodo       = "todo"
	StatusInProgress = "in-progress"
	StatusDone       = "done"
)

const (
	PriorityLow    = 1
	PriorityMedium = 2
	PriorityHigh   = 3
)

// Tag-related functions
func (tl *TaskList) AddTag(tag string) bool {
	// Check if tag already exists
	for _, t := range tl.Tags {
		if t == tag {
			return false
		}
	}
	tl.Tags = append(tl.Tags, tag)
	return true
}

func (tl *TaskList) HasTag(tag string) bool {
	for _, t := range tl.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
