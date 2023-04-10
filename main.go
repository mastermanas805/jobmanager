package main

import (
	"main/jobs"
	"main/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	s := &scheduler.Scheduler{
		Jobs:             make(scheduler.PriorityQueue, 0),
		JobRequests:      make(chan *jobs.Job),
		JobCancellations: make(chan *string),
		JobStatus:        make(map[string]string),
	}
	ui := &scheduler.SchedulerUI{
		Scheduler: s,
		JobStatus: make(map[string]*jobs.JobStatus),
	}

	r := gin.Default()
	scheduler.AddApis(r, s, ui, 5)

	r.Run("127.0.0.1:8080")
}
