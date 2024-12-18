package main

import (
	"fmt"
	"os"
	"task-tracker/internal/handler"
	"task-tracker/internal/model"
	"task-tracker/internal/storage"
	"task-tracker/internal/ui"
)

func printUsage() {
	ui.PrintLogo()
	fmt.Println("Usage: task-cli <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  add <description> <priority> [tags...]    Add a new task with optional tags")
	fmt.Println("  update <id> <description> <priority>      Update an existing task")
	fmt.Println("  set-priority <id> <priority>             Set the priority of a task")
	fmt.Println("  delete <id>                              Delete a task")
	fmt.Println("  mark-in-progress <id>                    Mark a task as in progress")
	fmt.Println("  mark-done <id>                           Mark a task as done")
	fmt.Println("  list [done|todo|in-progress|tag <tag>]   List tasks with optional status or tag filter")
	fmt.Println("  add-tag <tag>                            Add a new tag to the system")
	fmt.Println("  list-tags                                List all available tags")
	fmt.Println("  add-task-tag <task-id> <tag>            Add a tag to a task")
	fmt.Println("  remove-task-tag <task-id> <tag>         Remove a tag from a task")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	tasks, err := storage.LoadTasks()
	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	h := handler.NewCommandHandler(tasks)
	command := os.Args[1]
	args := os.Args[2:]

	var cmdErr error
	switch command {
	case "add":
		cmdErr = h.HandleAdd(args)
	case "set-priority":
		cmdErr = h.HandleSetPriority(args)
	case "update":
		cmdErr = h.HandleUpdate(args)
	case "delete":
		cmdErr = h.HandleDelete(args)
	case "mark-in-progress":
		cmdErr = h.HandleMarkStatus(args, model.StatusInProgress)
	case "mark-done":
		cmdErr = h.HandleMarkStatus(args, model.StatusDone)
	case "list":
		cmdErr = h.HandleList(args)
	case "add-tag":
		cmdErr = h.HandleAddTag(args)
	case "list-tags":
		cmdErr = h.HandleListTags(args)
	case "add-task-tag":
		cmdErr = h.HandleAddTaskTag(args)
	case "remove-task-tag":
		cmdErr = h.HandleRemoveTaskTag(args)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if cmdErr != nil {
		fmt.Println(cmdErr)
		os.Exit(1)
	}
}
