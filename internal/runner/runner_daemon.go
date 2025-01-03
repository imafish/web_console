package runner

import (
	"log"
	"time"

	"internal/db"
	"internal/pb"
)

type RunnerDaemon struct {
	TaskChan     <-chan *pb.Task
	IncomingChan chan<- bool
	ExitChan     chan<- bool

	taskChan     chan *pb.Task
	incomingChan chan bool
	exitChan     chan bool
	db           db.TaskDatabase
}

func (rd *RunnerDaemon) Close() {
	close(rd.exitChan)
}

func NewRunnerDaemon(db db.TaskDatabase) *RunnerDaemon {
	taskChan := make(chan *pb.Task)
	incomingChan := make(chan bool)
	exitChan := make(chan bool, 2)
	return &RunnerDaemon{
		TaskChan:     taskChan,
		IncomingChan: incomingChan,
		ExitChan:     exitChan,

		taskChan:     taskChan,
		incomingChan: incomingChan,
		exitChan:     exitChan,
		db:           db,
	}
}

// Run waits on the channel
func (rd *RunnerDaemon) Run() {
	for {
		select {
		case <-rd.exitChan:
			return
		case <-rd.incomingChan:
			task, err := rd.db.GetLatestTask()
			if err != nil {
				log.Printf("failed to get latest task to execute: %v", err)
				// wait for 5 seconds and try again.
				go func() {
					time.Sleep(5 * time.Second)
					rd.incomingChan <- true
				}()
				continue
			}
			ch, err := Run(task)
			if err != nil {
				log.Printf("failed to execute task %v, error: %v", task, err)
				// wait for 5 seconds and try again.
				go func() {
					time.Sleep(5 * time.Second)
					rd.incomingChan <- true
				}()
				continue
			}
			task = <-ch
			rd.taskChan <- task
		}
	}
}
