package scheduler

import (
	"github.com/gin-gonic/gin"
)

// Add Scheduler apis to the gin router object
func AddApis(r *gin.Engine, s *Scheduler, ui *SchedulerUI, workers int) {
	for w := 0; w < workers; w++ {
		go s.Start(ui)
	}

	r.Static("/static", "static")
	r.GET("/", ui.HomeHandler)
	r.POST("/schedule", ui.ScheduleHandler)
	r.GET("/status/:id", ui.StatusHandler)
	r.GET("/jobs", ui.JobsHandler)
	r.GET("/cancel/:id", ui.CancelHandler)
}
