package main

import (
	"fmt"
	"github-user-activity/api"
	"github-user-activity/models"
)

func main() {
	fmt.Println("Enter a GitHub username:")
	var username string
	fmt.Scanln(&username)

	if username == "" {
		fmt.Println("Error: Username cannot be empty!")
		return
	}

	events, err := api.FetchUserEvents(username)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Recently Activity:")
	for i, event := range events {
		if i >= 10 {
			break
		}

		switch event.Type {
		case models.PushEvent:
			fmt.Printf("- Pushed commits to %s\n", event.Repo.Name)
		case models.IssuesEvent:
			fmt.Printf("- Opened issue in %s\n", event.Repo.Name)
		case models.WatchEvent:
			fmt.Printf("- Starred %s\n", event.Repo.Name)
		default:
			fmt.Printf("- %s in %s\n", "Other", event.Repo.Name)
		}
	}

}
