package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"internal/pb"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50052"
)

func listTasks(client pb.TaskServiceClient, n int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ReadTaskListRequest{Count: int64(n)}
	res, err := client.ReadTaskList(ctx, req)
	if err != nil {
		log.Fatalf("could not list tasks: %v", err)
	}

	tasks, err := json.MarshalIndent(res.Tasks, "", "  ")
	if err != nil {
		log.Fatalf("could not marshal tasks: %v", err)
	}

	fmt.Println(string(tasks))
}

func newTask(client pb.TaskServiceClient, commandline string, workingDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	task := &pb.Task{WorkingDirectory: workingDir, Commandline: commandline}
	req := &pb.CreateTaskRequest{Task: task}
	res, err := client.CreateTask(ctx, req)
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}

	fmt.Printf("Created task with ID: %d\n", res.Task.Id)
}

func showTask(client pb.TaskServiceClient, id int64, onlyOutput, onlyStatus, onlyExitCode bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ReadTaskRequest{Id: id}
	res, err := client.ReadTask(ctx, req)
	if err != nil {
		log.Fatalf("could not show task: %v", err)
	}

	if onlyOutput {
		fmt.Println(res.Task.Output)
	} else if onlyStatus {
		fmt.Println(res.Task.Status)
	} else if onlyExitCode {
		fmt.Println(res.Task.ReturnCode)
	} else {
		task, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			log.Fatalf("could not marshal task: %v", err)
		}
		fmt.Println(string(task))
	}
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTaskServiceClient(conn)

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	showCmd := flag.NewFlagSet("show", flag.ExitOnError)

	listN := listCmd.Int("n", 10, "Number of tasks to list")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %v", err)
	}
	newWorkingDir := newCmd.String("w", cwd, "Working directory")

	showID := showCmd.Int64("id", -1, "Task ID")
	showOutput := showCmd.Bool("o", false, "Only print output path")
	showStatus := showCmd.Bool("s", false, "Only print status")
	showExitCode := showCmd.Bool("e", false, "Only print exit code")

	if len(os.Args) < 2 {
		fmt.Println("expected 'list', 'new' or 'show' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		listTasks(client, *listN)
	case "new":
		newCmd.Parse(os.Args[2:])
		commandline := newCmd.Args()
		if len(commandline) == 0 {
			fmt.Println("expected commandline arguments for new task")
			os.Exit(1)
		}
		newTask(client, strings.Join(commandline, " "), *newWorkingDir)
	case "show":
		showCmd.Parse(os.Args[2:])
		showTask(client, *showID, *showOutput, *showStatus, *showExitCode)
	default:
		fmt.Println("expected 'list', 'new' or 'show' subcommands")
		os.Exit(1)
	}
}
