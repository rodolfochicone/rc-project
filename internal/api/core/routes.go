package core

import "github.com/gin-gonic/gin"

// RegisterRoutes registers the shared daemon API routes on the supplied router.
func RegisterRoutes(router gin.IRouter, handlers *Handlers) {
	if router == nil || handlers == nil {
		return
	}

	api := router.Group("/api")
	registerDaemonRoutes(api, handlers)
	registerWorkspaceRoutes(api, handlers)
	registerTaskRoutes(api, handlers)
	registerReviewRoutes(api, handlers)
	registerRunRoutes(api, handlers)
	registerConfigRoutes(api, handlers)
	registerCatalogRoutes(api, handlers)
	registerSetupRoutes(api, handlers)

	api.GET("/ui/dashboard", handlers.GetDashboard)
	api.POST("/sync", handlers.SyncWorkflow)
	api.POST("/exec", handlers.StartExecRun)
}

func registerDaemonRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/daemon")
	g.GET("/status", h.DaemonStatus)
	g.GET("/health", h.DaemonHealth)
	g.GET("/metrics", h.DaemonMetrics)
	g.POST("/stop", h.StopDaemon)
}

func registerWorkspaceRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/workspaces")
	g.POST("", h.RegisterWorkspace)
	g.GET("", h.ListWorkspaces)
	g.POST("/sync", h.SyncWorkspaces)
	g.GET("/:id/ws", h.StreamWorkspaceSocket)
	g.GET("/:id", h.GetWorkspace)
	g.PATCH("/:id", h.UpdateWorkspace)
	g.DELETE("/:id", h.DeleteWorkspace)
	g.POST("/resolve", h.ResolveWorkspace)
}

func registerTaskRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/tasks")
	g.GET("", h.ListTaskWorkflows)
	g.GET("/:slug", h.GetTaskWorkflow)
	g.GET("/:slug/spec", h.GetWorkflowSpec)
	g.GET("/:slug/memory", h.GetWorkflowMemory)
	g.GET("/:slug/memory/files/:file_id", h.GetWorkflowMemoryFile)
	g.GET("/:slug/board", h.GetTaskBoard)
	g.GET("/:slug/items", h.ListTaskItems)
	g.GET("/:slug/items/:task_id", h.GetTaskItemDetail)
	g.POST("/:slug/validate", h.ValidateTaskWorkflow)
	g.POST("/:slug/runs", h.StartTaskRun)
	g.POST("/:slug/archive", h.ArchiveTaskWorkflow)
}

func registerReviewRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/reviews")
	g.POST("/:slug/fetch", h.FetchReview)
	g.POST("/:slug/watch", h.StartReviewWatch)
	g.GET("/:slug/rounds/:round/issues", h.ListReviewIssues)
	g.GET("/:slug/rounds/:round/issues/:issue_id", h.GetReviewIssue)
	g.GET("/:slug/rounds/:round", h.GetReviewRound)
	g.POST("/:slug/rounds/:round/runs", h.StartReviewRun)
	g.GET("/:slug", h.GetLatestReview)
}

func registerRunRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/runs")
	g.GET("", h.ListRuns)
	g.GET("/:run_id", h.GetRun)
	g.GET("/:run_id/snapshot", h.GetRunSnapshot)
	g.GET("/:run_id/transcript", h.GetRunTranscript)
	g.GET("/:run_id/events", h.ListRunEvents)
	g.GET("/:run_id/stream", h.StreamRun)
	g.POST("/:run_id/cancel", h.CancelRun)
	g.POST("/:run_id/input", h.SendInput)
}

func registerConfigRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/config")
	g.GET("/global", h.GetGlobalConfig)
	g.PUT("/global", h.PutGlobalConfig)
	g.GET("/workspace", h.GetWorkspaceConfig)
	g.PUT("/workspace", h.PutWorkspaceConfig)
}

func registerCatalogRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/catalog")
	g.GET("/extensions", h.ListCatalogExtensions)
	g.GET("/agents", h.ListCatalogAgents)
}

func registerSetupRoutes(api *gin.RouterGroup, h *Handlers) {
	g := api.Group("/setup")
	g.GET("/options", h.GetSetupOptions)
	g.POST("", h.RunSetup)
}
