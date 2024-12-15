package main

import (
	"bufio"
	"context"
	"fmt"
	"github-user-activity/api"
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

	for {
		if err := run(client); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Print("\nPress Enter to check again (Ctrl+C to exit): ")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func run(client *api.Client) error {
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

	printEvents(events)
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

func printEvents(events []models.GithubEvent) {
	if len(events) == 0 {
		fmt.Println("No recent activity found")
		return
	}

	fmt.Println("Recent Activity:")
	for i, event := range events {
		if i >= maxEvents {
			break
		}

		message := formatEventMessage(event)
		fmt.Println(message)
	}
}

func formatEventMessage(event models.GithubEvent) string {
	switch event.Type {
	case models.PushEvent:
		return fmt.Sprintf("- Pushed commits to %s\n", event.Repo.Name)
	case models.IssuesEvent:
		return fmt.Sprintf("- Opened issue in %s\n", event.Repo.Name)
	case models.WatchEvent:
		return fmt.Sprintf("- Starred %s\n", event.Repo.Name)
	default:
		return fmt.Sprintf("- %s in %s\n", "Other", event.Repo.Name)
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
