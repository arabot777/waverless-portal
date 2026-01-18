package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	HeaderServiceName = "X-Service-Name"
	HeaderTimestamp   = "X-Timestamp"
	HeaderSignature   = "X-Signature"
	SignatureExpiry   = 5 * time.Minute
)

// SignRequest 为请求添加内部服务认证签名
func SignRequest(req *http.Request, serviceName, secretKey string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := GenerateSignature(req.Method, req.URL.Path, timestamp, secretKey)

	req.Header.Set(HeaderServiceName, serviceName)
	req.Header.Set(HeaderTimestamp, timestamp)
	req.Header.Set(HeaderSignature, signature)
}

// GenerateSignature 生成 HMAC-SHA256 签名
// 签名内容: METHOD + PATH + TIMESTAMP
func GenerateSignature(method, path, timestamp, secretKey string) string {
	data := fmt.Sprintf("%s\n%s\n%s", method, path, timestamp)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名 (主站使用)
func VerifySignature(method, path, timestamp, signature, secretKey string) error {
	// 检查时间戳
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > SignatureExpiry {
		return fmt.Errorf("signature expired")
	}

	// 验证签名
	expected := GenerateSignature(method, path, timestamp, secretKey)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
