# Task Tracker CLI

A simple command-line interface (CLI) to track and manage your tasks.

## Project Structure

```
task-tracker/
├── cmd/
│   └── task-cli/
│       └── main.go         # Main application entry point
├── internal/
│   ├── model/
│   │   └── task.go        # Task data models
│   ├── storage/
│   │   └── storage.go     # Task storage operations
│   └── handler/
│       └── commands.go     # Command handlers
├── go.mod
└── README.md
```

## Features

- Add, Update, and Delete tasks
- Mark tasks as in progress or done
- List all tasks
- List tasks by status (todo, in-progress, done)
- Persistent storage using JSON file

## Building

```bash
go build -o task-cli ./cmd/task-cli
```

## Usage

```bash
# Adding a new task
./task-cli add "Buy groceries"

# Updating a task
./task-cli update 1 "Buy groceries and cook dinner"

# Deleting a task
./task-cli delete 1

# Marking a task as in progress or done
./task-cli mark-in-progress 1
./task-cli mark-done 1

# Listing tasks
./task-cli list           # List all tasks
./task-cli list done      # List completed tasks
./task-cli list todo      # List todo tasks
./task-cli list in-progress  # List in-progress tasks
```

## Data Storage

Tasks are stored in a `tasks.json` file in the current directory. The file is created automatically when you add your first task. 