package jobs

import (
	"context"
	"sync"
	"time"
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

type Job struct {
	Mu           sync.Mutex
	ID           string
	Func         func(ctx context.Context, cancel context.CancelFunc)
	ScheduledAt  time.Time
	Timeout      time.Duration
	CurrentState string
	Context      *Context
}

type Context struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
