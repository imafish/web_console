package runner

import (
	"internal/pb"
)

// TaskListener writes a message to the pipe when updates to Tasks occur
type TaskListener struct {
	incomingChan chan<- bool
}

func NewTaskListener(incomingChan chan<- bool) *TaskListener {
	return &TaskListener{incomingChan: incomingChan}
}

func (tl *TaskListener) OnTaskCreated(task *pb.Task) {
	tl.incomingChan <- true
}

func (tl *TaskListener) OnTaskUpdated(task *pb.Task) {
	tl.incomingChan <- true
}

func (tl *TaskListener) OnTaskDeleted(task *pb.Task) {
	tl.incomingChan <- true
}

func (tl *TaskListener) OnTaskExecuted(task *pb.Task) {
	tl.incomingChan <- true
}
