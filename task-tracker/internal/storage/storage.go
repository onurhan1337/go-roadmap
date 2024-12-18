package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"task-tracker/internal/model"
)

func LoadTasks() (*model.TaskList, error) {
	data, err := os.ReadFile("tasks.json")
	if os.IsNotExist(err) {
		return &model.TaskList{Tasks: []model.Task{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading tasks file: %v", err)
	}

	var tasks model.TaskList
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("error parsing tasks file: %v", err)
	}
	return &tasks, nil
}

func SaveTasks(tasks *model.TaskList) error {
	data, err := json.MarshalIndent(tasks, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding tasks: %v", err)
	}

	if err := os.WriteFile("tasks.json", data, 0644); err != nil {
		return fmt.Errorf("error saving tasks file: %v", err)
	}
	return nil
}

func FindTaskByID(tasks *model.TaskList, id int) (int, *model.Task) {
	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == id {
			return i, &tasks.Tasks[i]
		}
	}
	return -1, nil
}
