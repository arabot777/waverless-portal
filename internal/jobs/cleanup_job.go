package jobs

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// CleanupJob 清理过期数据
type CleanupJob struct {
	db *gorm.DB
}

func NewCleanupJob(db *gorm.DB) *CleanupJob {
	return &CleanupJob{db: db}
}

// Start 启动清理任务 (每天执行一次)
func (j *CleanupJob) Start(ctx context.Context) {
	// 启动时执行一次
	j.cleanup()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.cleanup()
		}
	}
}

func (j *CleanupJob) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -7)

	// 清理 billing_transactions
	result := j.db.Exec("DELETE FROM billing_transactions WHERE created_at < ?", cutoff)
	if result.Error != nil {
		log.Printf("[Cleanup] failed to clean billing_transactions: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("[Cleanup] deleted %d billing_transactions older than 7 days", result.RowsAffected)
	}

	// 清理 task_routing
	result = j.db.Exec("DELETE FROM task_routing WHERE submitted_at < ?", cutoff)
	if result.Error != nil {
		log.Printf("[Cleanup] failed to clean task_routing: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("[Cleanup] deleted %d task_routing older than 7 days", result.RowsAffected)
	}
}
