package service

import (
	"context"
	"internal/db"
	"internal/pb"
	"log"
)

type TaskStatusProxy interface {
	RegisterListener(listener TaskServiceListener)
	RemoveListener(listener TaskServiceListener)
}

type TaskServiceServer struct {
	pb.UnimplementedTaskServiceServer // Embedding for forward compatibility
	taskDB                            db.TaskDatabase
	listeners                         []TaskServiceListener
}

// NewTaskServiceServer creates a new TaskServiceServer
func NewTaskServiceServer(taskDB db.TaskDatabase) *TaskServiceServer {
	return &TaskServiceServer{
		taskDB:    taskDB,
		listeners: make([]TaskServiceListener, 0)}
}

// RegisterListener implements the TaskStatusProxy interface
func (s *TaskServiceServer) RegisterListener(listener TaskServiceListener) {
	s.listeners = append(s.listeners, listener)
}

// RemoveListener implements the TaskStatusProxy interface
func (s *TaskServiceServer) RemoveListener(listener TaskServiceListener) {
	for i, l := range s.listeners {
		if l == listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return
		}
	}
}

// ReadTask implements the ReadTask gRPC method
func (s *TaskServiceServer) ReadTask(ctx context.Context, req *pb.ReadTaskRequest) (*pb.TaskResponse, error) {
	task, err := s.taskDB.GetTask(req.Id)
	if err != nil {
		log.Printf("Readtask: Failed to get task: %v", err)
		return nil, err
	}
	return &pb.TaskResponse{Task: task}, nil
}

// DeleteTask implements the DeleteTask gRPC method
func (s *TaskServiceServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.TaskResponse, error) {
	// Retrieve the task before deleting it
	task, err := s.taskDB.GetTask(req.Id)
	if err != nil {
		log.Printf("DeleteTask: Failed to get task: %v", err)
		return nil, err
	}

	err = s.taskDB.DeleteTask(req.Id)
	if err != nil {
		log.Printf("DeleteTask: Failed to delete task: %v", err)
		return nil, err
	}

	go func() {
		for _, l := range s.listeners {
			l.OnTaskDeleted(task)
		}
	}()

	return &pb.TaskResponse{Task: task}, nil
}

// ReadTaskList implements the ReadTaskList gRPC method
func (s *TaskServiceServer) ReadTaskList(ctx context.Context, req *pb.ReadTaskListRequest) (*pb.TaskListResponse, error) {
	tasks, err := s.taskDB.GetTasks()
	if err != nil {
		log.Printf("ReadTaskList: Failed to get tasks: %v", err)
		return nil, err
	}
	return &pb.TaskListResponse{Tasks: tasks}, nil
}

func (s *TaskServiceServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	task, err := s.taskDB.CreateTask(req.GetTask())
	if err != nil {
		log.Printf("CreateTask: Failed to create task: %v", err)
		return nil, err
	}

	go func() {
		for _, l := range s.listeners {
			l.OnTaskCreated(task)
		}
	}()

	return &pb.TaskResponse{Task: task}, nil
}
