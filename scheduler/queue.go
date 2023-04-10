package scheduler

import (
	"container/heap"
	"main/jobs"
	"time"
)

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].ScheduledAt.Before(pq[j].ScheduledAt)
}
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*jobs.Job)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq PriorityQueue) Peek() *jobs.Job {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}

func (pq *PriorityQueue) AddJob(job *jobs.Job) {
	heap.Push(pq, job)
}

func (s *Scheduler) Start(ui *SchedulerUI) {
	for {
		select {
		case job := <-s.JobRequests:
			s.Jobs.AddJob(job)
			s.AllJobs.AddJob(job)
			s.JobStatus[job.ID] = "scheduled"
			if job := ui.JobStatus[job.ID]; job != nil {
				job.Status = "scheduled"
			}
		case id := <-s.JobCancellations:
			for i, job := range s.Jobs {
				if job.ID == *id {
					heap.Remove(&s.Jobs, i)
					s.JobStatus[job.ID] = "cancelled"
					ui.JobStatus[job.ID].Status = "cancelled"
				}
			}
			//cancel context
			for _, job := range s.AllJobs {
				if job.ID == *id {
					job.Context.Cancel()
				}
			}
		default:
			nextJob := s.Jobs.Peek()
			if nextJob == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			nextJob.Mu.Lock()
			if nextJob.ScheduledAt.After(time.Now()) {
				if nextJob.ScheduledAt == time.Now() {
					heap.Pop(&s.Jobs)
					nextJob.Mu.Unlock()
				} else if time.Until(nextJob.ScheduledAt) < 30*time.Second {
					heap.Pop(&s.Jobs)
					nextJob.Mu.Unlock()
					time.Sleep(time.Until(nextJob.ScheduledAt))
				} else {
					nextJob.Mu.Unlock()
					time.Sleep(1 * time.Second)
					continue
				}
			}
			s.JobStatus[nextJob.ID] = "processing"
			if job := ui.JobStatus[nextJob.ID]; job != nil {
				job.Status = "processing"
			}
			go func(job *jobs.Job) {
				ExecuteJob(job, ui)
			}(nextJob)

		}
	}
}
