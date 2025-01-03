package db_test

import (
	"fmt"
	"internal/db"
	"internal/pb"
	"log"
	"os"
	"testing"
)

const (
	db_path = "_"
)

func clean() {
	if _, err := os.Stat(db_path); err == nil {
		// File exists, delete it
		err = os.Remove(db_path)
		if err != nil {
			log.Fatalf("cannot delete db file: %v", err)
		}
		log.Printf("deleted file `%s\"", db_path)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error checking file:", err)
	}
}

func TestMain(m *testing.M) {
	clean()
	code := m.Run()
	clean()
	os.Exit(code)
}

func TestSmokeTest(t *testing.T) {
	database, err := db.NewTaskDatabase("_")
	if err != nil {
		t.Fatalf("db.NewTaskDatabase() should not return error, but got %v", err)
	}

	err = database.Init()
	if err != nil {
		t.Fatalf("db.Init() should not return error, but got %v", err)
	}

	tasks, err := database.GetTasks()
	if err != nil {
		t.Errorf("db.GetTasks() should not return error, but got %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expect 0 tasks, but got %d", len(tasks))
	}

	task := &pb.Task{
		Id:               0,
		Status:           pb.TaskStatus_NEW,
		WorkingDirectory: os.Getenv("HOME"),
		Commandline:      "ls",
	}

	task, err = database.CreateTask(task)
	if err != nil {
		t.Errorf("should create task but got error: %v", err)
	}
	if task.Id != 1 {
		t.Errorf("expect the ID to be 1, but got %d", task.Id)
	}

	tasks, err = database.GetTasks()
	if err != nil {
		t.Errorf("should get list of tasks, but got error: %v", err)
	} else {
		length := len(tasks)
		if length != 1 {
			t.Errorf("expect to have 1 task, but got %d", length)
		}
		if length > 0 {
			if tasks[0].Id != 1 {
				t.Errorf("expect the ID to be 1, but got %d", task.Id)
			}
		}
		task2, err := database.GetTask(task.Id)
		if err != nil {
			t.Errorf("expect to get a task, but get error: %v", err)
		} else {
			if task2.Id != 1 {
				t.Errorf("expect id to be 1, but got %d", task2.Id)
			}
			if task2.Status != pb.TaskStatus_NEW {
				t.Errorf("expect status to be NEW, but got %s", task2.Status.String())
			}

			task2.Status = pb.TaskStatus_FINISHED
			task2.ReturnCode = 1

			task2, err = database.UpdateTask(task2)
			if err != nil {
				t.Errorf("expect to update a task, but get error: %v", err)
			}
			if task2.Id != 1 {
				t.Errorf("expect id to be 1, but got %d", task2.Id)
			}
			if task2.Status != pb.TaskStatus_FINISHED {
				t.Errorf("expect status to be FINISHED, but got %s", task2.Status.String())
			}

			err = database.DeleteTask(task2.Id)
			if err != nil {
				t.Errorf("expect to delete a task, but get error: %v", err)
			}
			tasks, err = database.GetTasks()
			if err != nil {
				t.Errorf("expect to get tasks, but get error: %v", err)
			}
			length := len(tasks)
			if length != 0 {
				t.Errorf("expect to have 0 tasks, but got %d", length)
			}
			task, err = database.GetTask(1)
			if err == nil {
				t.Error("expect to return error, but got none")
			}
			if task != nil {
				t.Errorf("expect to get nil task, but got %v", task)
			}
		}
	}

}
