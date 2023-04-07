package main

import (
	"container/heap"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type JobForm struct {
	Func    string `form:"func" binding:"required"`
	Timeout int    `form:"timeout" binding:"required"`
	Time    string `form:"time" binding:"required"`
}

type JobStatus struct {
	ID     string
	Status string
	Error  interface{}
}

type SchedulerUI struct {
	Scheduler *Scheduler
	JobStatus map[string]*JobStatus
}

func render(c *gin.Context, name string, data interface{}) {
	t, err := template.ParseFiles("templates/"+name+".html", "templates/base.html")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = t.ExecuteTemplate(c.Writer, "base.html", data)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (ui *SchedulerUI) HomeHandler(c *gin.Context) {
	render(c, "home", gin.H{})
}

func (ui *SchedulerUI) ScheduleHandler(c *gin.Context) {
	var form JobForm
	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	t, err := time.ParseInLocation("2006-01-02T15:04", form.Time, time.Local)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	job := &Job{
		ID: strconv.FormatInt(time.Now().UnixNano(), 10),
		Func: func() {
			fmt.Println(form.Func)
			slice := []int{1, 2, 3}
			fmt.Println(slice[10])
			time.Sleep(10 * time.Second)
		},
		ScheduledAt: t,
		Timeout:     time.Duration(form.Timeout) * time.Second,
	}

	ui.Scheduler.JobRequests <- job
	ui.JobStatus[job.ID] = &JobStatus{ID: job.ID, Status: "scheduled"}

	c.Redirect(http.StatusSeeOther, "/")
}

func (ui *SchedulerUI) StatusHandler(c *gin.Context) {
	id := c.Param("id")
	status := ui.JobStatus[id]
	if status == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	render(c, "status", gin.H{"status": status})
}

func (ui *SchedulerUI) JobsHandler(c *gin.Context) {
	var jobs []*Job
	for _, job := range ui.Scheduler.AllJobs {
		jobs = append(jobs, job)
	}
	render(c, "jobs", gin.H{"jobs": jobs, "jobStatus": ui.JobStatus})
}

func (ui *SchedulerUI) CancelHandler(c *gin.Context) {
	id := c.Param("id")
	ui.Scheduler.JobCancellations <- id
	ui.JobStatus[id].Status = "cancelled"
	c.Redirect(http.StatusSeeOther, "/jobs")
}

func main() {
	s := &Scheduler{
		Jobs:             make(PriorityQueue, 0),
		JobRequests:      make(chan *Job),
		JobCancellations: make(chan string),
		JobStatus:        make(map[string]string),
	}
	ui := &SchedulerUI{
		Scheduler: s,
		JobStatus: make(map[string]*JobStatus),
	}

	go s.Start(ui)

	r := gin.Default()

	r.Static("/static", "static")

	r.GET("/", ui.HomeHandler)
	r.POST("/schedule", ui.ScheduleHandler)
	r.GET("/status/:id", ui.StatusHandler)
	r.GET("/jobs", ui.JobsHandler)
	r.GET("/cancel/:id", ui.CancelHandler)

	r.Run("127.0.0.1:8080")
}

// Struct

type Job struct {
	ID           string
	Func         func()
	ScheduledAt  time.Time
	Timeout      time.Duration
	CurrentState string
}

type Scheduler struct {
	Jobs             PriorityQueue
	AllJobs          PriorityQueue
	JobRequests      chan *Job
	JobCancellations chan string
	JobStatus        map[string]string
}

// Priority Queue

type PriorityQueue []*Job

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].ScheduledAt.Before(pq[j].ScheduledAt)
}
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Job)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq PriorityQueue) Peek() *Job {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}

func (pq *PriorityQueue) AddJob(job *Job) {
	heap.Push(pq, job)
}

// Scheduler

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
				if job.ID == id {
					heap.Remove(&s.Jobs, i)
					s.JobStatus[job.ID] = "cancelled"
					ui.JobStatus[job.ID].Status = "cancelled"
				}
			}
		default:
			nextJob := s.Jobs.Peek()
			if nextJob == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if nextJob.ScheduledAt.After(time.Now()) {
				time.Sleep(time.Until(nextJob.ScheduledAt))
				continue
			}
			heap.Pop(&s.Jobs)
			s.JobStatus[nextJob.ID] = "processing"
			if job := ui.JobStatus[nextJob.ID]; job != nil {
				job.Status = "processing"
			}
			go func(job *Job) {
				ExecuteJob(job, ui)
			}(nextJob)
		}
	}
}

// ExecuteJob

func ExecuteJob(job *Job, ui *SchedulerUI) {
	job.CurrentState = "processing"

	go func(job *Job, ui *SchedulerUI) {

		timeout := time.After(job.Timeout)
		done := make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Job Failed", r)
					job.CurrentState = "failed"
					ui.JobStatus[job.ID].Status = job.CurrentState
					ui.JobStatus[job.ID].Error = r
				}
			}()
			job.Func()
			job.CurrentState = "completed"
			ui.JobStatus[job.ID].Status = job.CurrentState
			close(done)
		}()
		select {
		case <-timeout:
			job.CurrentState = "timeout"
			ui.JobStatus[job.ID].Status = job.CurrentState
		case <-done:
		}
	}(job, ui)
}
