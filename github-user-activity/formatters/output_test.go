package formatters

import (
	"github-user-activity/models"
	"strings"
	"testing"
)

func TestFormatOutput(t *testing.T) {
	testEvents := []models.GithubEvent{
		{
			Type: models.EventType("PushEvent"),
			Repo: struct {
				Name string `json:"name"`
			}{Name: "test/repo"},
		},
		{
			Type: models.EventType("WatchEvent"),
			Repo: struct {
				Name string `json:"name"`
			}{Name: "another/repo"},
		},
	}

	tests := []struct {
		name     string
		events   []models.GithubEvent
		format   models.OutputFormat
		contains []string
	}{
		{
			name:   "JSON format",
			events: testEvents,
			format: models.JSONFormat,
			contains: []string{
				"PushEvent",
				"test/repo",
				"WatchEvent",
				"another/repo",
			},
		},
		{
			name:   "Table format",
			events: testEvents,
			format: models.TableFormat,
			contains: []string{
				"Event Type",
				"Repository",
				"PushEvent",
				"test/repo",
			},
		},
		{
			name:   "Simple format",
			events: testEvents,
			format: models.SimpleFormat,
			contains: []string{
				"Code Contributions",
				"Repositories Starred",
				"test/repo",
				"another/repo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatOutput(tt.events, tt.format)
			for _, str := range tt.contains {
				if !strings.Contains(output, str) {
					t.Errorf("expected output to contain %q, but it didn't", str)
				}
			}
		})
	}
}
