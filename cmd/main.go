package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wavespeedai/waverless-portal/app/handler"
	"github.com/wavespeedai/waverless-portal/app/router"
	"github.com/wavespeedai/waverless-portal/internal/jobs"
	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/rocketmq"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Infof("Starting Portal server...")

	mysqlRepo, err := mysql.NewRepository()
	if err != nil {
		logger.Fatalf("Failed to init MySQL: %v", err)
	}
	defer mysqlRepo.Close()

	redisClient, err := redis.NewClient()
	if err != nil {
		logger.Fatalf("Failed to init Redis: %v", err)
	}
	defer redisClient.Close()

	// RocketMQ
	if err := rocketmq.Init(); err != nil {
		logger.Warnf("Failed to init RocketMQ: %v", err)
	}
	defer rocketmq.Close()

	// Repositories
	userRepo := mysql.NewUserRepo(mysqlRepo.DB)
	clusterRepo := mysql.NewClusterRepo(mysqlRepo.DB)
	endpointRepo := mysql.NewEndpointRepo(mysqlRepo.DB)
	billingRepo := mysql.NewBillingRepo(mysqlRepo.DB)
	specRepo := mysql.NewSpecRepo(mysqlRepo.DB)
	workerRepo := mysql.NewWorkerRepo(mysqlRepo.DB)
	taskRepo := mysql.NewTaskRepo(mysqlRepo.DB)
	registryCredentialRepo := mysql.NewRegistryCredentialRepo(mysqlRepo.DB)

	// Services
	userService := service.NewUserService(userRepo)
	clusterService := service.NewClusterService(clusterRepo)
	endpointService := service.NewEndpointService(endpointRepo, clusterRepo, specRepo, registryCredentialRepo, clusterService)
	taskService := service.NewTaskService(taskRepo, endpointService, clusterService)
	specService := service.NewSpecService(specRepo)
	billingService := service.NewBillingService(billingRepo, userRepo, endpointRepo)

	// Handlers
	specHandler := handler.NewSpecHandler(specService)
	endpointHandler := handler.NewEndpointHandler(endpointService, userService)
	taskHandler := handler.NewTaskHandler(taskService)
	billingHandler := handler.NewBillingHandler(billingService, userService)
	clusterHandler := handler.NewClusterHandler(clusterService)
	webhookHandler := handler.NewWebhookHandler(billingService)
	monitoringHandler := handler.NewMonitoringHandler(endpointService, clusterService, taskRepo, workerRepo, nil)
	userHandler := handler.NewUserHandler(userService)
	registryCredentialHandler := handler.NewRegistryCredentialHandler(registryCredentialRepo)

	// Router
	r := router.NewRouter(
		specHandler,
		endpointHandler,
		taskHandler,
		billingHandler,
		clusterHandler,
		webhookHandler,
		monitoringHandler,
		userHandler,
		registryCredentialHandler,
		userService,
	)

	if config.GlobalConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	r.Setup(engine)

	// Background jobs
	jobsManager := jobs.NewManager(billingService, clusterService)
	jobsManager.Start()

	// Endpoint sync service
	endpointSyncService := jobs.NewEndpointSyncService(endpointRepo, clusterService, endpointService)
	endpointSyncService.Start()

	// Worker sync job
	workerSyncJob := jobs.NewWorkerSyncJob(mysqlRepo.DB, clusterService, endpointService)
	go workerSyncJob.Start(context.Background())

	// Metrics sync job
	metricsSyncJob := jobs.NewMetricsSyncJob(mysqlRepo.DB, clusterService, endpointService)
	go metricsSyncJob.Start(context.Background())

	// Task sync job
	taskSyncJob := jobs.NewTaskSyncJob(mysqlRepo.DB, clusterService, endpointService)
	go taskSyncJob.Start(context.Background())

	// Billing job
	billingJob := jobs.NewBillingJob(mysqlRepo.DB, workerRepo, billingRepo, endpointRepo, endpointService)
	go billingJob.Start(context.Background())

	// Cleanup job (delete data older than 7 days)
	cleanupJob := jobs.NewCleanupJob(mysqlRepo.DB)
	go cleanupJob.Start(context.Background())

	// Update monitoringHandler with workerSyncJob
	monitoringHandler.SetWorkerSyncJob(workerSyncJob)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port),
		Handler: engine,
	}

	go func() {
		logger.Infof("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Shutting down server...")
	jobsManager.Stop()
	endpointSyncService.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	jobsManager.Wait()
	logger.Infof("Server stopped")
}
