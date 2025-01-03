package runner

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"internal/db"
	"internal/pb"
)

const (
	outputDir = "tmp/output"
)

type RunnerDaemon struct {
	TaskChan     <-chan *pb.Task
	IncomingChan chan<- bool
	ExitChan     chan<- bool

	taskChan     chan *pb.Task
	incomingChan chan bool
	exitChan     chan bool
	db           db.TaskDatabase
	outputDir    string
}

func (rd *RunnerDaemon) Close() {
	close(rd.exitChan)
}

func NewRunnerDaemon(db db.TaskDatabase) *RunnerDaemon {
	taskChan := make(chan *pb.Task)
	incomingChan := make(chan bool)
	exitChan := make(chan bool, 2)
	dir := outputDir
	if !filepath.IsAbs(outputDir) {
		dir = filepath.Join(os.Getenv("HOME"), outputDir)
	}
	log.Printf("output directory: %v", dir)
	return &RunnerDaemon{
		TaskChan:     taskChan,
		IncomingChan: incomingChan,
		ExitChan:     exitChan,

		taskChan:     taskChan,
		incomingChan: incomingChan,
		exitChan:     exitChan,
		db:           db,
		outputDir:    dir,
	}
}

// Run waits on the channel
func (rd *RunnerDaemon) Run() {
	for {
		select {
		case <-rd.exitChan:
			return
		case <-rd.incomingChan:
			rd.runTask()
		}
	}
}

func (rd *RunnerDaemon) runTask() {
	task, err := rd.db.GetLatestTask()
	if err != nil {
		log.Printf("failed to get latest task to execute: %v", err)
		if _, ok := err.(*db.ErrNoRows); !ok {
			rd.retry()
		}
		return
	}

	log.Printf("got task %s", task.AsJsonString())

	output := task.GetOutput()
	if len(output) == 0 {
		tempFile, err := os.CreateTemp(rd.outputDir, "task_output_*.log")
		if err != nil {
			log.Printf("failed to create temporary file: %v", err)
			rd.retry()
			return
		}
		tempFile.Close()
		log.Printf("tempFile path is: %s", tempFile.Name())
		task.Output = tempFile.Name()
	}

	log.Printf("Executing task %v", task.AsJsonString())
	receivingChan, err := Run(task)
	if err != nil {
		log.Printf("failed to execute task %v, error: %v", task, err)
		rd.retry()
		return
	}
	task2 := <-receivingChan

	log.Printf("Updating task status to RUNNING: %v", task2.AsJsonString())
	_, err = rd.db.UpdateTask(task2)
	if err != nil {
		log.Printf("Failed to update task status to RUNNING: %v", err)
	}

	task3 := <-receivingChan
	if task3 == nil {
		log.Printf("task failed to execute: %v", task)
		_, err = rd.db.UpdateTask(task)
		if err != nil {
			log.Printf("Failed to update task status to NEW: %v", err)
		}
		rd.retry()
		return
	}
	log.Printf("Updating task status to FINISHED: %v", task3.AsJsonString())
	_, err = rd.db.UpdateTask(task3)
	if err != nil {
		log.Printf("Failed to update task status to FINISHED: %v", err)
	}

	rd.taskChan <- task
}

func (rd *RunnerDaemon) retry() {
	go func() {
		time.Sleep(5 * time.Second)
		rd.incomingChan <- true
	}()
}
