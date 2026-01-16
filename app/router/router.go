package router

import (
	"github.com/wavespeedai/waverless-portal/app/handler"
	"github.com/wavespeedai/waverless-portal/app/middleware"
	"github.com/wavespeedai/waverless-portal/internal/service"

	"github.com/gin-gonic/gin"
)

type Router struct {
	specHandler       *handler.SpecHandler
	endpointHandler   *handler.EndpointHandler
	taskHandler       *handler.TaskHandler
	billingHandler    *handler.BillingHandler
	clusterHandler    *handler.ClusterHandler
	webhookHandler    *handler.WebhookHandler
	monitoringHandler *handler.MonitoringHandler
	userHandler       *handler.UserHandler
	userService       *service.UserService
}

func NewRouter(
	specHandler *handler.SpecHandler,
	endpointHandler *handler.EndpointHandler,
	taskHandler *handler.TaskHandler,
	billingHandler *handler.BillingHandler,
	clusterHandler *handler.ClusterHandler,
	webhookHandler *handler.WebhookHandler,
	monitoringHandler *handler.MonitoringHandler,
	userHandler *handler.UserHandler,
	userService *service.UserService,
) *Router {
	return &Router{
		specHandler:       specHandler,
		endpointHandler:   endpointHandler,
		taskHandler:       taskHandler,
		billingHandler:    billingHandler,
		clusterHandler:    clusterHandler,
		webhookHandler:    webhookHandler,
		monitoringHandler: monitoringHandler,
		userHandler:       userHandler,
		userService:       userService,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// V1 API - 任务提交接口 (RunPod 兼容)
	v1 := engine.Group("/v1")
	v1.Use(middleware.JWTAuthWithUserService(r.userService))
	{
		v1.GET("/status/:task_id", r.taskHandler.GetTaskStatus)
		v1.POST("/cancel/:task_id", r.taskHandler.CancelTask)

		endpoint := v1.Group("/:endpoint")
		{
			endpoint.POST("/run", r.taskHandler.SubmitTask)
			endpoint.POST("/runsync", r.taskHandler.SubmitTaskSync)
		}
	}

	// API v1 - 用户管理接口
	api := engine.Group("/api/v1")
	{
		// 规格查询 (公开)
		api.GET("/specs", r.specHandler.ListSpecs)
		api.GET("/specs/:name", r.specHandler.GetSpec)
		api.POST("/estimate-cost", r.specHandler.EstimateCost)

		// 用户信息
		api.GET("/user", middleware.JWTAuthWithUserService(r.userService), r.userHandler.GetCurrentUser)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.JWTAuthWithUserService(r.userService))
		{
			// 全局任务查询
			if r.monitoringHandler != nil {
				auth.GET("/tasks", r.monitoringHandler.GetAllTasks)
				auth.GET("/tasks/overview", r.monitoringHandler.GetTasksOverview)
			}

			// Endpoint 管理
			endpoints := auth.Group("/endpoints")
			{
				endpoints.POST("", r.endpointHandler.CreateEndpoint)
				endpoints.GET("", r.endpointHandler.ListEndpoints)
				endpoints.GET("/:name", r.endpointHandler.GetEndpoint)
				endpoints.PUT("/:name", r.endpointHandler.UpdateEndpoint)
				endpoints.PUT("/:name/config", r.endpointHandler.UpdateEndpointConfig)
				endpoints.DELETE("/:name", r.endpointHandler.DeleteEndpoint)

				// Endpoint 监控
				if r.monitoringHandler != nil {
					endpoints.GET("/:name/workers", r.monitoringHandler.GetEndpointWorkers)
					endpoints.GET("/:name/workers/exec", r.monitoringHandler.ExecWorker)
					endpoints.GET("/:name/logs", r.monitoringHandler.GetWorkerLogs)
					endpoints.GET("/:name/tasks", r.monitoringHandler.GetWorkerTasks)
					endpoints.GET("/:name/metrics", r.monitoringHandler.GetEndpointMetrics)
					endpoints.GET("/:name/stats", r.monitoringHandler.GetEndpointStats)
					endpoints.GET("/:name/statistics", r.monitoringHandler.GetEndpointStatistics)
				}
			}

			// 计费查询
			billing := auth.Group("/billing")
			{
				billing.GET("/balance", r.billingHandler.GetBalance)
				billing.GET("/usage", r.billingHandler.GetUsage)
				billing.GET("/workers", r.billingHandler.GetWorkerRecords)
				billing.GET("/records", r.billingHandler.GetRechargeRecords)
			}
		}

		// Webhook 接口 (Waverless 调用)
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/worker-created", r.webhookHandler.WorkerCreated)
			webhooks.POST("/worker-terminated", r.webhookHandler.WorkerTerminated)
		}

		// 管理员接口 (暂时不需要认证，方便开发)
		admin := api.Group("/admin")
		{
			admin.GET("/clusters", r.clusterHandler.ListClusters)
			admin.GET("/clusters/:id", r.clusterHandler.GetCluster)
			admin.GET("/clusters/:id/specs", r.clusterHandler.GetClusterSpecs)
			admin.POST("/clusters/:id/specs", r.clusterHandler.CreateClusterSpec)
			admin.PUT("/clusters/:id/specs", r.clusterHandler.UpdateClusterSpec)
			admin.DELETE("/clusters/:id/specs", r.clusterHandler.DeleteClusterSpec)
			admin.POST("/clusters", r.clusterHandler.CreateCluster)
			admin.PUT("/clusters/:id", r.clusterHandler.UpdateCluster)
			admin.DELETE("/clusters/:id", r.clusterHandler.DeleteCluster)

			// 规格管理
			admin.GET("/specs", r.specHandler.ListAllSpecs)
			admin.POST("/specs", r.specHandler.CreateSpec)
			admin.PUT("/specs", r.specHandler.UpdateSpec)
			admin.DELETE("/specs", r.specHandler.DeleteSpec)
		}
	}
}
