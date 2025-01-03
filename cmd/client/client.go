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

func printTask(client pb.TaskServiceClient, id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ReadTaskRequest{Id: id}
	res, err := client.ReadTask(ctx, req)
	if err != nil {
		log.Fatalf("could not read task: %v", err)
	}

	outputPath := res.Task.Output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}

	fmt.Print(string(content))
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
	catCmd := flag.NewFlagSet("cat", flag.ExitOnError)
	flagSets := map[string]*flag.FlagSet{
		"list": listCmd,
		"new":  newCmd,
		"show": showCmd,
		"cat":  catCmd,
	}

	listN := listCmd.Int("n", 10, "Number of tasks to list")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %v", err)
	}
	newWorkingDir := newCmd.String("w", cwd, "Working directory")

	showID := showCmd.Int64("i", -1, "Task ID")
	showOutput := showCmd.Bool("o", false, "Only print output path")
	showStatus := showCmd.Bool("s", false, "Only print status")
	showExitCode := showCmd.Bool("e", false, "Only print exit code")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("Commands:")
		fmt.Println("  list -n <number>       List tasks")
		fmt.Println("  new -w <directory>     Create a new task")
		fmt.Println("  show -i <task_id>     Show task details")
		fmt.Println("  cat -i <task_id>      Print the task output")
		for _, subCmd := range flagSets {
			subCmd.PrintDefaults()
		}
	}

	catId := catCmd.Int64("i", -1, "Task ID")

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printHelp(flagSets)
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
	case "cat":
		catCmd.Parse(os.Args[2:])
		printTask(client, *catId)
	default:
		printHelp(flagSets)
	}
}

func printHelp(flagSets map[string]*flag.FlagSet) {
	subCmds := make([]string, 0, len(flagSets))
	for c := range flagSets {
		subCmds = append(subCmds, fmt.Sprintf("'%s'", c))
	}
	fmt.Printf("expected %s subcommands", strings.Join(subCmds, " "))
	flag.Usage()
	os.Exit(1)
}
