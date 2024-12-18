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
	if len(args) < 2 {
		return fmt.Errorf("usage: task-cli add <description> <priority>")
	}
	description := args[0]

	priority := 0
	if _, err := fmt.Sscanf(args[1], "%d", &priority); err != nil {
		return fmt.Errorf("invalid priority format: must be 1 (Low), 2 (Medium), or 3 (High)")
	}

	if priority < model.PriorityLow || priority > model.PriorityHigh {
		return fmt.Errorf("invalid priority: must be 1 (Low), 2 (Medium), or 3 (High)")
	}

	newID := 1
	if len(h.tasks.Tasks) > 0 {
		newID = h.tasks.Tasks[len(h.tasks.Tasks)-1].ID + 1
	}

	task := model.Task{
		ID:          newID,
		Description: description,
		Status:      model.StatusTodo,
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	h.tasks.Tasks = append(h.tasks.Tasks, task)

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving task: %v", err)
	}

	priorityStr := getPriorityString(priority)
	fmt.Printf("Task added successfully (ID: %d) with priority %s\n",
		newID, priorityStr)
	return nil
}

func (h *CommandHandler) HandleSetPriority(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: task-cli set-priority <id> <priority>")
	}

	id := 0
	priority := 0
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid ID format")
	}
	if _, err := fmt.Sscanf(args[1], "%d", &priority); err != nil {
		return fmt.Errorf("invalid priority format: must be 1 (Low), 2 (Medium), or 3 (High)")
	}

	if priority < 1 || priority > 3 {
		return fmt.Errorf("invalid priority: must be 1 (Low), 2 (Medium), or 3 (High)")
	}

	_, task := storage.FindTaskByID(h.tasks, id)
	if task == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	task.Priority = priority
	task.UpdatedAt = time.Now()

	if err := storage.SaveTasks(h.tasks); err != nil {
		return fmt.Errorf("error saving task: %v", err)
	}
	fmt.Printf("Task %d priority set to %s\n", id, getPriorityString(priority))
	return nil
}

func (h *CommandHandler) HandleUpdate(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: task-cli update <id> <description> <priority>")
	}

	id := 0
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid ID format")
	}
	description := args[1]
	priority := 0
	if _, err := fmt.Sscanf(args[2], "%d", &priority); err != nil {
		return fmt.Errorf("invalid priority format")
	}

	_, task := storage.FindTaskByID(h.tasks, id)
	if task == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	task.Description = description
	task.Priority = priority
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

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

func getPriorityString(priority int) string {
	switch priority {
	case model.PriorityHigh:
		return colorRed + "High" + colorReset
	case model.PriorityMedium:
		return colorYellow + "Medium" + colorReset
	case model.PriorityLow:
		return colorGreen + "Low" + colorReset
	default:
		return colorGreen + "Low" + colorReset
	}
}

func getStatusColor(status string) string {
	switch status {
	case model.StatusDone:
		return colorGreen + status + colorReset
	case model.StatusInProgress:
		return colorYellow + status + colorReset
	default:
		return colorBlue + status + colorReset
	}
}

type Table struct {
	Headers    []string
	Rows       [][]string
	ColWidths  []int
	TotalWidth int
}

func createTable(tasks []model.Task, filterStatus string) *Table {
	headers := []string{"ID", "Status", "Description", "Created", "Updated", "Priority"}
	var rows [][]string

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, task := range tasks {
		if filterStatus == "" || task.Status == filterStatus ||
			(filterStatus == model.StatusTodo && task.Status == model.StatusTodo) {
			row := []string{
				fmt.Sprintf("%d", task.ID),
				getStatusColor(task.Status),
				truncateString(task.Description, 40),
				task.CreatedAt.Format("2006-01-02"),
				task.UpdatedAt.Format("2006-01-02"),
				getPriorityString(task.Priority),
			}

			for i, cell := range row {
				cleanCell := strings.ReplaceAll(cell, colorReset, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorRed, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorGreen, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorYellow, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorBlue, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorPurple, "")
				cleanCell = strings.ReplaceAll(cleanCell, colorCyan, "")

				if len(cleanCell) > colWidths[i] {
					colWidths[i] = len(cleanCell)
				}
			}
			rows = append(rows, row)
		}
	}

	totalWidth := 1
	for _, w := range colWidths {
		totalWidth += w + 3
	}
	totalWidth++

	return &Table{
		Headers:    headers,
		Rows:       rows,
		ColWidths:  colWidths,
		TotalWidth: totalWidth,
	}
}

func (t *Table) print() {
	fmt.Print("\n╭")
	for i, width := range t.ColWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.ColWidths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("╮")

	fmt.Print("│")
	for i, header := range t.Headers {
		fmt.Printf(" %s%-*s%s │", colorCyan, t.ColWidths[i], header, colorReset)
	}
	fmt.Println()

	fmt.Print("├")
	for i, width := range t.ColWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.ColWidths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	for _, row := range t.Rows {
		fmt.Print("│")
		for i, cell := range row {
			cleanCell := strings.ReplaceAll(cell, colorReset, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorRed, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorGreen, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorYellow, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorBlue, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorPurple, "")
			cleanCell = strings.ReplaceAll(cleanCell, colorCyan, "")

			padding := strings.Repeat(" ", t.ColWidths[i]-len(cleanCell))
			fmt.Printf(" %s%s │", cell, padding)
		}
		fmt.Println()
	}

	fmt.Print("╰")
	for i, width := range t.ColWidths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.ColWidths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("╯")
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

	table := createTable(h.tasks.Tasks, filterStatus)
	table.print()

	return nil
}

func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	lastSpace := strings.LastIndex(str[:length-3], " ")
	if lastSpace > 0 {
		return str[:lastSpace] + "..."
	}
	return str[:length-3] + "..."
}
