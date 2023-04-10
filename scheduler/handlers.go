package scheduler

import (
	"context"
	"main/jobs"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
)

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
	var form jobs.JobForm
	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	t, err := time.ParseInLocation("2006-01-02T15:04", form.Time, time.Local)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(form.Timeout)*time.Second)

	job := &jobs.Job{
		ID:          strconv.FormatInt(time.Now().UnixNano(), 10),
		Func:        FuncOptions[form.Func],
		ScheduledAt: t,
		Timeout:     time.Duration(form.Timeout) * time.Second,
		Context: &jobs.Context{
			Ctx:    ctx,
			Cancel: cancel,
		},
	}

	ui.Scheduler.JobRequests <- job
	ui.JobStatus[job.ID] = &jobs.JobStatus{ID: job.ID, Status: "scheduled"}

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
	var jobs []*jobs.Job
	for _, job := range ui.Scheduler.AllJobs {
		jobs = append(jobs, job)
	}
	render(c, "jobs", gin.H{"jobs": jobs, "jobStatus": ui.JobStatus})
}

func (ui *SchedulerUI) CancelHandler(c *gin.Context) {
	id := c.Param("id")
	ui.Scheduler.JobCancellations <- &id
	c.Redirect(http.StatusSeeOther, "/jobs")
}
