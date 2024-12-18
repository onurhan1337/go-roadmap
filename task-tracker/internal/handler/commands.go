package handler

import (
	"fmt"
	"strings"
	"task-tracker/internal/model"
	"task-tracker/internal/storage"
	"time"
)

type CommandHandler struct {
	tasks *model.TaskList
}

func NewCommandHandler(tasks *model.TaskList) *CommandHandler {
	return &CommandHandler{tasks: tasks}
}

func (h *CommandHandler) HandleAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task-cli add <description>")
	}
	description := args[0]
	newID := 1
	if len(h.tasks.Tasks) > 0 {
		newID = h.tasks.Tasks[len(h.tasks.Tasks)-1].ID + 1
	}

	task := model.Task{
		ID:          newID,
		Description: description,
		Status:      model.StatusTodo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	h.tasks.Tasks = append(h.tasks.Tasks, task)

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving task: %v", err)
	}
	fmt.Printf("Task added successfully (ID: %d)\n", newID)
	return nil
}

func (h *CommandHandler) HandleUpdate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: task-cli update <id> <description>")
	}

	id := 0
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid ID format")
	}
	description := args[1]

	_, task := storage.FindTaskByID(h.tasks, id)
	if task == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	task.Description = description
	task.UpdatedAt = time.Now()

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving task: %v", err)
	}
	fmt.Printf("Task %d updated successfully\n", id)
	return nil
}

func (h *CommandHandler) HandleDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task-cli delete <id>")
	}

	id := 0
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid ID format")
	}

	index, _ := storage.FindTaskByID(h.tasks, id)
	if index == -1 {
		return fmt.Errorf("task with ID %d not found", id)
	}

	h.tasks.Tasks = append(h.tasks.Tasks[:index], h.tasks.Tasks[index+1:]...)

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving tasks: %v", err)
	}
	fmt.Printf("Task %d deleted successfully\n", id)
	return nil
}

func (h *CommandHandler) HandleMarkStatus(args []string, status string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task-cli mark-%s <id>", status)
	}

	id := 0
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid ID format")
	}

	_, task := storage.FindTaskByID(h.tasks, id)
	if task == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving task: %v", err)
	}
	fmt.Printf("Task %d marked as %s\n", id, status)
	return nil
}

func (h *CommandHandler) HandleList(args []string) error {
	filterStatus := ""
	if len(args) > 0 {
		filterStatus = args[0]
		if filterStatus != model.StatusDone && filterStatus != model.StatusTodo && filterStatus != model.StatusInProgress {
			return fmt.Errorf("invalid status filter. Use: done, todo, or in-progress")
		}
	}

	if len(h.tasks.Tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Print header
	fmt.Println("\n╭" + strings.Repeat("─", 90) + "╮")
	fmt.Printf("│ %-4s │ %-11s │ %-40s │ %-14s │ %-12s │\n", "ID", "Status", "Description", "Created", "Updated")
	fmt.Println("├" + strings.Repeat("─", 6) + "┼" + strings.Repeat("─", 13) + "┼" + strings.Repeat("─", 42) + "┼" + strings.Repeat("─", 16) + "┼" + strings.Repeat("─", 14) + "┤")

	// Print tasks
	for _, task := range h.tasks.Tasks {
		if filterStatus == "" || task.Status == filterStatus ||
			(filterStatus == model.StatusTodo && task.Status == model.StatusTodo) {
			fmt.Printf("│ %-4d │ %-11s │ %-40s │ %-14s │ %-12s │\n",
				task.ID,
				task.Status,
				truncateString(task.Description, 40),
				task.CreatedAt.Format("2006-01-02"),
				task.UpdatedAt.Format("2006-01-02"))
		}
	}

	// Print footer
	fmt.Println("╰" + strings.Repeat("─", 90) + "╯")
	return nil
}

func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length-3] + "..."
}
