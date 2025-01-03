package runner

import (
	"log"
	"os"
	"os/exec"
	"time"

	"internal/pb"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Run(task *pb.Task) (<-chan *pb.Task, error) {
	cmd := exec.Command("sh", "-c", task.Commandline)
	cmd.Dir = task.WorkingDirectory

	outputPathAbsolute := task.GetOutput()
	outputFile, err := os.OpenFile(outputPathAbsolute, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	log.Printf("running task %s", task.AsJsonString())
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	cmd.Stdout = outputFile
	cmd.Stderr = outputFile

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	ch := make(chan *pb.Task)

	go func() {
		defer close(ch)
		startTime := time.Now()
		task.StartTime = timestamppb.New(startTime)
		task.Status = pb.TaskStatus_RUNNING
		ch <- task
		err := cmd.Wait()
		if err != nil {
			ch <- nil
			return
		}
		finishTime := time.Now()
		task.FinishTime = timestamppb.New(finishTime)
		task.ReturnCode = int32(cmd.ProcessState.ExitCode())
		task.ExecutionTime = durationpb.New(finishTime.Sub(startTime))
		task.Status = pb.TaskStatus_FINISHED
		log.Printf("Task %d finished with return code %d", task.Id, task.ReturnCode)
		ch <- task
	}()

	return ch, nil
}
