package rocketmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
)

const (
	TopicMeteringBilling = "TOPIC_METERING_BILLING"
	TagDeductExec        = "DEDUCT_EXEC"
)

var globalProducer rocketmq.Producer

// BillingMessage 计费消息
type BillingMessage struct {
	UserID      string    `json:"user_id"`
	OrgID       string    `json:"org_id"`
	RequestID   string    `json:"request_id"`   // 幂等 key
	EndpointID  int64     `json:"endpoint_id"`
	WorkerID    string    `json:"worker_id"`
	Amount      int64     `json:"amount"`       // 微美元
	DurationSec int64     `json:"duration_sec"` // 计费时长
	Service     string    `json:"service"`
	Timestamp   time.Time `json:"timestamp"`
}

// Init 初始化 RocketMQ Producer
func Init() error {
	cfg := config.GlobalConfig.RocketMQ
	if cfg.NameServer == "" {
		logger.Infof("[RocketMQ] not configured, skip init")
		return nil
	}

	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{cfg.NameServer}),
		producer.WithGroupName(cfg.ProducerGroup),
		producer.WithRetry(2),
	)
	if err != nil {
		return err
	}

	if err := p.Start(); err != nil {
		return err
	}

	globalProducer = p
	logger.Infof("[RocketMQ] producer started: %s", cfg.NameServer)
	return nil
}

// Close 关闭 Producer
func Close() {
	if globalProducer != nil {
		globalProducer.Shutdown()
	}
}

// SendBillingMessage 发送计费消息
func SendBillingMessage(ctx context.Context, msg *BillingMessage) error {
	if globalProducer == nil {
		return nil // MQ 未初始化，跳过
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	mqMsg := &primitive.Message{
		Topic: TopicMeteringBilling,
		Body:  body,
	}
	mqMsg.WithTag(TagDeductExec)
	mqMsg.WithKeys([]string{msg.RequestID})

	_, err = globalProducer.SendSync(ctx, mqMsg)
	if err != nil {
		logger.ErrorCtx(ctx, "[RocketMQ] send billing message failed: %v", err)
		return err
	}

	logger.InfoCtx(ctx, "[RocketMQ] billing message sent: request_id=%s, org_id=%s, amount=%d",
		msg.RequestID, msg.OrgID, msg.Amount)
	return nil
}
