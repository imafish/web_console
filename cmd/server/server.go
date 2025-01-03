package main

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"google.golang.org/grpc"

	"internal/db"
	"internal/pb"
	"internal/runner"
	"internal/service"
)

const (
	dbPath     = "tasks.db"
	outputPath = "output"
	tmpDir     = "tmp"
	listenAddr = ":50052"
	protocol   = "tcp"
)

func main() {
	tmpDir := filepath.Join(os.Getenv("HOME"), tmpDir)

	// Initialize the database
	taskDB, err := db.NewTaskDatabase(filepath.Join(tmpDir, dbPath))
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	err = taskDB.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer taskDB.Uninit()

	var wg sync.WaitGroup
	wg.Add(2)

	// Initialize the runner service
	runnerDaemon := runner.NewRunnerDaemon(taskDB)
	go func() {
		defer runnerDaemon.Close()
		defer wg.Done()
		log.Printf("Runner service started")
		runnerDaemon.Run()
		log.Printf("Runner service stopped")
	}()

	// Initialize the gRPC server
	server := grpc.NewServer()
	taskService := service.NewTaskServiceServer(taskDB)

	// create a listner to receive task update events
	taskListener := runner.NewTaskListener(runnerDaemon.IncomingChan)
	taskService.RegisterListener(taskListener)

	// Register the TaskServiceServer with the gRPC server
	pb.RegisterTaskServiceServer(server, taskService)

	// Start the gRPC server
	listener, err := net.Listen(protocol, listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	go func() {
		defer func() { runnerDaemon.ExitChan <- true }()
		defer wg.Done()
		log.Printf("gRPC server is listening on %s", listenAddr)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	go func() {
		for t := range runnerDaemon.TaskChan {
			log.Printf("Updated: ID: %d, CMD: %s, Status %s\n", t.Id, t.Commandline, t.GetStatus().String())
		}
	}()

	// Wait for the goroutines to complete
	wg.Wait()
	log.Print("Server stopped")
}
