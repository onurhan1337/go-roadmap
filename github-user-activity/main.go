package main

import (
	"bufio"
	"context"
	"fmt"
	"github-user-activity/api"
	"github-user-activity/formatters"
	"github-user-activity/models"
	"os"
	"strings"
	"time"
)

const (
	maxEvents    = 10
	readTimeout  = 30 * time.Second
	eventTimeout = 10 * time.Second
)

func main() {
	client := api.NewClient(os.Getenv("GITHUB_TOKEN"))
	format := models.SimpleFormat

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--json":
			format = models.JSONFormat
		case "--table":
			format = models.TableFormat
		}
	}

	for {
		if err := run(client, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Print("\nPress Enter to check again (Ctrl+C to exit): ")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func run(client *api.Client, format models.OutputFormat) error {
	printLogo()
	printUsageInfo()

	username, err := promptUsername()
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()

	events, err := client.FetchUserEvents(ctx, username)
	if err != nil {
		return err
	}

	if rateLimit := client.GetRateLimit(); rateLimit != nil {
		fmt.Printf("\nRate limit remaining: %d (resets at %s)\n",
			rateLimit.Remaining,
			rateLimit.ResetAt.Local().Format(time.Kitchen))
	}

	printEvents(events, format)
	return nil
}

func promptUsername() (string, error) {
	fmt.Print("Enter a GitHub username: ")

	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	username = strings.TrimSpace(username)

	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	return username, nil
}

func printEvents(events []models.GithubEvent, format models.OutputFormat) {
	fmt.Print(formatters.FormatOutput(events, format))
}

func formatEventMessage(event models.GithubEvent) string {
	switch event.Type {
	case models.PushEvent:
		return fmt.Sprintf("🔨 Pushed commits to %s", event.Repo.Name)
	case models.IssuesEvent:
		return fmt.Sprintf("🐛 Opened issue in %s", event.Repo.Name)
	case models.WatchEvent:
		return fmt.Sprintf("⭐ Starred %s", event.Repo.Name)
	case models.ForkEvent:
		return fmt.Sprintf("🔱 Forked %s", event.Repo.Name)
	case models.CreateEvent:
		return fmt.Sprintf("📝 Created repository/branch in %s", event.Repo.Name)
	case models.DeleteEvent:
		return fmt.Sprintf("🗑️  Deleted branch/tag in %s", event.Repo.Name)
	case models.PullRequestEvent:
		return fmt.Sprintf("🔄 Pull request activity in %s", event.Repo.Name)
	case models.ReleaseEvent:
		return fmt.Sprintf("📦 Released version in %s", event.Repo.Name)
	default:
		return fmt.Sprintf("❓ %s in %s", string(event.Type), event.Repo.Name)
	}
}

func printLogo() {
	logo := `
	╔════════════════════════════════╗
	║   ╔═╗ ╔╗╔ ╦ ╦ ╦═╗             ║
	║   ║ ║ ║║║ ║ ║ ╠╦╝             ║
	║   ╚═╝ ╝╚╝ ╚═╝ ╩╚═             ║
	║                                ║
	║   GitHub Activity CLI v1.0     ║
	╚════════════════════════════════╝
	`
	fmt.Println(logo)
}

func printUsageInfo() {
	info := `
🔑 USAGE WITH GITHUB TOKEN:
	1. Create a token at: https://github.com/settings/tokens
	2. Export your token:
	  export GITHUB_TOKEN=your_token_here
	3. Run the program

📝 NOTE: Using a token increases API rate limits
`
	fmt.Println(info)
}
