package formatters

import (
	"encoding/json"
	"fmt"
	"github-user-activity/models"
	"strings"
)

func FormatOutput(events []models.GithubEvent, format models.OutputFormat) string {
	switch format {
	case models.JSONFormat:
		return formatJSON(events)
	case models.TableFormat:
		return formatTable(events)
	case models.SimpleFormat:
		return formatSimple(events)
	default:
		return ""
	}
}

func formatJSON(events []models.GithubEvent) string {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(data)
}

func formatTable(events []models.GithubEvent) string {
	if len(events) == 0 {
		return "No events found"
	}

	var sb strings.Builder
	sb.WriteString("\n┌──────────────┬────────────────────────────────┐\n")
	sb.WriteString("│ Event Type    │ Repository                     │\n")
	sb.WriteString("├──────────────┼────────────────────────────────┤\n")

	for _, event := range events {
		eventType := padRight(string(event.Type), 12)
		repoName := padRight(event.Repo.Name, 30)
		sb.WriteString(fmt.Sprintf("│ %s │ %s │\n", eventType, repoName))
	}

	sb.WriteString("└──────────────┴────────────────────────────────┘\n")

	return sb.String()
}

func formatSimple(events []models.GithubEvent) string {
	var sb strings.Builder
	sb.WriteString("\nRecent Activity:\n")
	sb.WriteString("═══════════════\n")

	eventGroups := make(map[models.EventType][]models.GithubEvent)
	for _, event := range events {
		eventGroups[event.Type] = append(eventGroups[event.Type], event)
	}

	for eventType, typeEvents := range eventGroups {
		switch eventType {
		case models.WatchEvent:
			sb.WriteString("\n⭐ Repositories Starred:")
		case models.PushEvent:
			sb.WriteString("\n🔨 Code Contributions:")
		case models.ForkEvent:
			sb.WriteString("\n🔱 Forked Repositories:")
		case models.CreateEvent:
			sb.WriteString("\n📝 Created Repositories/Branches:")
		case models.DeleteEvent:
			sb.WriteString("\n🗑️  Deleted Branches/Tags:")
		case models.PullRequestEvent:
			sb.WriteString("\n🔄 Pull Request Activity:")
		case models.ReleaseEvent:
			sb.WriteString("\n📦 Released Versions:")
		case models.IssueCommentEvent:
			sb.WriteString("\n💬 Issue Comments:")
		case models.CommitCommentEvent:
			sb.WriteString("\n💭 Commit Comments:")
		case models.PublicEvent:
			sb.WriteString("\n🌟 Made Public:")
		case models.MemberEvent:
			sb.WriteString("\n👥 Collaborator Activity:")
		default:
			sb.WriteString(fmt.Sprintf("\n❓ Other Activity (%s):", eventType))
		}

		sb.WriteString("\n")
		for i, event := range typeEvents {
			if i < 3 {
				sb.WriteString(fmt.Sprintf("  • %s\n", event.Repo.Name))
			}
		}

		if len(typeEvents) > 3 {
			sb.WriteString(fmt.Sprintf("  └─ and %d more...\n", len(typeEvents)-3))
		}
	}

	return sb.String()
}

func padRight(str string, length int) string {
	if len(str) >= length {
		return str[:length]
	}
	return str + strings.Repeat(" ", length-len(str))
}
