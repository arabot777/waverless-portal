# Project Context

## 项目信息
- 项目名称: waverless-portal
- 后端: Go (Gin + GORM)
- 前端: React + TypeScript + Vite
- 数据库: MySQL

## 开发环境
- 测试域名: portal-test.wavespeed.ai:3000
- 后端构建: `cd /Users/shanliu/work/wavespeedai/waverless-portal && go build -o portal ./cmd/main.go`
- 前端构建: `cd /Users/shanliu/work/wavespeedai/waverless-portal/web-ui && npm run build`

## 代码规范
- 价格单位: int64, 1000000 = 1 USD
- API 返回价格时用 `ToUSD()` 转换为 float64
- API 接收价格时用 `FromUSD()` 转换为 int64
  
## 数据库连接方式
- /opt/homebrew/opt/mysql-client@8.0/bin/mysql -h 127.0.0.1 -P 3306 -u root -p'root@123!23'   
- 数据库名：waverless-portal

## 相关项目
- waverless: /Users/shanliu/work/wavespeedai/waverless (核心调度系统)

