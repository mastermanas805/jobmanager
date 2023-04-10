package scheduler

import (
	"main/jobs"
	"time"
)

func ExecuteJob(job *jobs.Job, ui *SchedulerUI) {
	job.CurrentState = "processing"

	go func(job *jobs.Job, ui *SchedulerUI) {
		done := make(chan bool)
		defer close(done)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if job.CurrentState != "cancelled" {
						job.CurrentState = "failed"
					}
					ui.JobStatus[job.ID].Status = job.CurrentState
					ui.JobStatus[job.ID].Error = r
					done <- true
				}
			}()
			job.Func(job.Context.Ctx, job.Context.Cancel)
			job.CurrentState = "completed"
			ui.JobStatus[job.ID].Status = job.CurrentState
			done <- true
		}()
		select {
		case <-time.After(job.Timeout):
			job.CurrentState = "timeout"
			ui.JobStatus[job.ID].Status = job.CurrentState
		case <-done:
			return
		}
	}(job, ui)
}
