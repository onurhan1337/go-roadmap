package api

import (
	"encoding/json"
	"fmt"
	"github-user-activity/models"
	"io"
	"net/http"
)

func FetchUserEvents(username string) ([]models.GithubEvent, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Request failed with status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var events []models.GithubEvent
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return events, nil
}
