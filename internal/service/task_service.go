package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/waverless"
	"gorm.io/datatypes"
)

type TaskService struct {
	taskRepo        *mysql.TaskRepo
	endpointService *EndpointService
	clusterService  *ClusterService
}

func NewTaskService(taskRepo *mysql.TaskRepo, endpointService *EndpointService, clusterService *ClusterService) *TaskService {
	return &TaskService{taskRepo: taskRepo, endpointService: endpointService, clusterService: clusterService}
}

func (s *TaskService) SubmitTask(ctx context.Context, userID, orgID, logicalName string, input map[string]interface{}) (*waverless.TaskResponse, error) {
	endpoint, err := s.endpointService.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return nil, err
	}
	resp, err := s.endpointService.GetWaverlessClient(cluster).SubmitTask(ctx, endpoint.PhysicalName, input)
	if err != nil {
		return nil, err
	}
	inputJSON, _ := json.Marshal(input)
	now := time.Now()
	s.taskRepo.Create(ctx, &model.TaskRouting{
		TaskID: resp.ID, UserID: userID, OrgID: orgID, EndpointID: endpoint.ID, ClusterID: endpoint.ClusterID,
		Input: datatypes.JSON(inputJSON), Status: "PENDING", SubmittedAt: now, CreatedAt: &now,
	})
	return resp, nil
}

func (s *TaskService) SubmitTaskSync(ctx context.Context, userID, orgID, logicalName string, input map[string]interface{}) (*waverless.TaskResponse, error) {
	endpoint, err := s.endpointService.GetByLogicalName(ctx, userID, logicalName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, endpoint.ClusterID)
	if err != nil {
		return nil, err
	}
	resp, err := s.endpointService.GetWaverlessClient(cluster).SubmitTaskSync(ctx, endpoint.PhysicalName, input)
	if err != nil {
		return nil, err
	}
	inputJSON, _ := json.Marshal(input)
	now := time.Now()
	s.taskRepo.Create(ctx, &model.TaskRouting{
		TaskID: resp.ID, UserID: userID, OrgID: orgID, EndpointID: endpoint.ID, ClusterID: endpoint.ClusterID,
		Input: datatypes.JSON(inputJSON), Status: resp.Status, WorkerID: resp.WorkerID, SubmittedAt: now, CreatedAt: &now,
		ExecutionTimeMs: resp.ExecutionTime,
	})
	return resp, nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, userID, taskID string) (*waverless.TaskResponse, error) {
	routing, err := s.taskRepo.GetByTaskID(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, routing.ClusterID)
	if err != nil {
		return nil, err
	}
	resp, err := s.endpointService.GetWaverlessClient(cluster).GetTaskStatus(ctx, taskID)
	if err != nil {
		return nil, err
	}
	s.taskRepo.Update(ctx, taskID, resp.Status, resp.WorkerID, resp.ExecutionTime)
	return resp, nil
}

func (s *TaskService) CancelTask(ctx context.Context, userID, taskID string) error {
	routing, err := s.taskRepo.GetByTaskID(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	cluster, err := s.clusterService.GetCluster(ctx, routing.ClusterID)
	if err != nil {
		return err
	}
	return s.endpointService.GetWaverlessClient(cluster).CancelTask(ctx, taskID)
}
