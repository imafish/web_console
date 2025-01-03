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
	defer taskDB.Uninit()

	var wg sync.WaitGroup
	wg.Add(2)

	runnerService := runner.NewRunnerDaemon(taskDB)
	go func() {
		defer runnerService.Close()
		defer wg.Done()
		runnerService.Run()
	}()

	// Initialize the gRPC server
	server := grpc.NewServer()
	taskService := service.NewTaskServiceServer(taskDB)

	// Register the TaskServiceServer with the gRPC server
	pb.RegisterTaskServiceServer(server, taskService)

	// Start listening on a port
	listener, err := net.Listen(protocol, listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("gRPC server is listening on port 50051...")

	go func() {
		defer wg.Done()
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for the goroutines to complete
	wg.Wait()
	log.Print("Server stopped")
}
