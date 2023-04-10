package scheduler

import "main/jobs"

type PriorityQueue []*jobs.Job

type SchedulerUI struct {
	Scheduler *Scheduler
	JobStatus map[string]*jobs.JobStatus
}

type Scheduler struct {
	Jobs             PriorityQueue
	AllJobs          PriorityQueue
	JobRequests      chan *jobs.Job
	JobCancellations chan *string
	JobStatus        map[string]string
}
