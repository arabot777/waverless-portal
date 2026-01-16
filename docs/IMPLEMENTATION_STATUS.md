# Portal 实现状态总结

> 更新时间: 2026-01-16

本文档总结 Portal 项目当前已实现的功能，对照 `PORTAL_DESIGN.md` 设计文档。

---

## 1. 整体架构

### 1.1 已实现

- ✅ Portal 作为中心化控制平面
- ✅ 多集群管理架构
- ✅ JWT Cookie 认证（与主站集成）
- ✅ Waverless 集群注册与心跳
- ✅ 用户 Endpoint 管理
- ✅ 任务提交与状态查询（透传 Waverless）

### 1.2 未实现

- ❌ 计费服务（Worker 级计费）
- ❌ 余额管理与扣费
- ❌ 集群智能调度算法（当前使用简单选择）
- ❌ 对账机制

---

## 2. 数据库

### 2.1 已实现的表

| 表名 | 状态 | 说明 |
|------|------|------|
| `clusters` | ✅ | 集群注册表 |
| `cluster_specs` | ✅ | 集群规格表 |
| `spec_pricing` | ✅ | 规格价格配置表 |
| `user_endpoints` | ✅ | 用户 Endpoint 表 |

### 2.2 未实现的表

| 表名 | 状态 | 说明 |
|------|------|------|
| `user_balances` | ❌ | 用户余额表 |
| `recharge_records` | ❌ | 充值记录表 |
| `billing_transactions` | ❌ | 计费流水表 |
| `worker_billing_state` | ❌ | Worker 计费状态表 |
| `task_routing` | ❌ | 任务路由记录表 |
| `user_preferences` | ❌ | 用户偏好设置表 |

---

## 3. API 实现状态

### 3.1 用户侧 API

#### 认证
| API | 状态 | 说明 |
|-----|------|------|
| JWT Cookie 认证 | ✅ | 从主站获取 token |
| `GET /api/v1/user` | ✅ | 获取当前用户信息 |

#### 规格查询
| API | 状态 | 说明 |
|-----|------|------|
| `GET /api/v1/specs` | ✅ | 查询所有可用规格 |
| `GET /api/v1/specs?type=GPU/CPU` | ✅ | 按类型过滤规格 |
| `POST /api/v1/estimate-cost` | ✅ | 估算成本 |

#### Endpoint 管理
| API | 状态 | 说明 |
|-----|------|------|
| `POST /api/v1/endpoints` | ✅ | 创建 Endpoint |
| `GET /api/v1/endpoints` | ✅ | 列出用户 Endpoints |
| `GET /api/v1/endpoints/:name` | ✅ | 获取 Endpoint 详情 |
| `PUT /api/v1/endpoints/:name` | ✅ | 更新 Endpoint（image, env, replicas） |
| `PUT /api/v1/endpoints/:name/config` | ✅ | 更新 Endpoint 配置（scaling） |
| `DELETE /api/v1/endpoints/:name` | ✅ | 删除 Endpoint |

#### Endpoint 监控
| API | 状态 | 说明 |
|-----|------|------|
| `GET /api/v1/endpoints/:name/workers` | ✅ | 获取 Worker 列表 |
| `GET /api/v1/endpoints/:name/metrics` | ✅ | 获取实时指标 |
| `GET /api/v1/endpoints/:name/stats` | ✅ | 获取统计数据 |
| `GET /api/v1/endpoints/:name/statistics` | ✅ | 获取统计信息 |
| `GET /api/v1/endpoints/:name/logs` | ✅ | 获取 Worker 日志 |
| `GET /api/v1/endpoints/:name/tasks` | ✅ | 获取 Endpoint 任务列表 |
| `GET /api/v1/endpoints/:name/workers/exec` | ✅ | WebSocket 终端 |

#### 任务管理
| API | 状态 | 说明 |
|-----|------|------|
| `POST /v1/:endpoint/run` | ✅ | 异步提交任务 |
| `POST /v1/:endpoint/runsync` | ✅ | 同步提交任务 |
| `GET /v1/status/:task_id` | ✅ | 查询任务状态 |
| `POST /v1/cancel/:task_id` | ✅ | 取消任务 |
| `GET /api/v1/tasks` | ✅ | 获取用户所有任务 |
| `GET /api/v1/tasks/overview` | ✅ | 获取任务总览统计 |

#### 计费查询（未实现）
| API | 状态 | 说明 |
|-----|------|------|
| `GET /api/v1/billing/balance` | ❌ | 查询余额 |
| `GET /api/v1/billing/usage` | ❌ | 查询使用统计 |
| `GET /api/v1/billing/workers` | ❌ | 查询 Worker 费用明细 |
| `POST /api/v1/billing/recharge` | ❌ | 充值 |
| `GET /api/v1/billing/records` | ❌ | 充值记录 |

### 3.2 管理员 API

| API | 状态 | 说明 |
|-----|------|------|
| `GET /api/v1/admin/clusters` | ✅ | 查询所有集群 |
| `GET /api/v1/admin/clusters/:id` | ✅ | 获取集群详情 |
| `POST /api/v1/admin/clusters` | ✅ | 创建集群 |
| `PUT /api/v1/admin/clusters/:id` | ✅ | 更新集群 |
| `DELETE /api/v1/admin/clusters/:id` | ✅ | 删除集群 |
| `GET /api/v1/admin/clusters/:id/specs` | ✅ | 获取集群规格 |
| `POST /api/v1/admin/clusters/:id/specs` | ✅ | 创建集群规格 |
| `PUT /api/v1/admin/clusters/:id/specs` | ✅ | 更新集群规格 |
| `DELETE /api/v1/admin/clusters/:id/specs` | ✅ | 删除集群规格 |
| `GET /api/v1/admin/specs` | ✅ | 获取全局规格 |
| `POST /api/v1/admin/specs` | ✅ | 创建全局规格 |
| `PUT /api/v1/admin/specs` | ✅ | 更新全局规格 |
| `DELETE /api/v1/admin/specs` | ✅ | 删除全局规格 |

---

## 4. Web UI 实现状态

### 4.1 页面

| 页面 | 状态 | 功能 |
|------|------|------|
| Endpoints 列表 | ✅ | 展示用户所有 Endpoints，支持创建、删除 |
| Endpoint 详情 | ✅ | 多 Tab 页面：Overview, Metrics, Workers, Tasks, Settings |
| Tasks 全局页面 | ✅ | 展示用户所有任务，支持搜索、过滤、分页 |
| Admin - Clusters | ✅ | 集群管理（CRUD） |
| Admin - Specs | ✅ | 规格管理（CRUD） |

### 4.2 Endpoint 详情页功能

#### Overview Tab
| 功能 | 状态 | 说明 |
|------|------|------|
| 基本信息展示 | ✅ | 名称、状态、规格、集群、价格等 |
| Quick Start | ✅ | API 调用示例（curl/python/js） |
| 任务提交 | ✅ | 支持 sync/async 模式 |

#### Metrics Tab
| 功能 | 状态 | 说明 |
|------|------|------|
| 请求数趋势图 | ✅ | ECharts 折线图 |
| 执行时间趋势图 | ✅ | ECharts 折线图 |
| Worker 数量趋势图 | ✅ | ECharts 折线图 |
| 时间范围选择 | ✅ | 1h/6h/24h/7d |

#### Workers Tab
| 功能 | 状态 | 说明 |
|------|------|------|
| Worker 卡片列表 | ✅ | 展示 Worker 状态、任务数、运行时长 |
| Worker 详情抽屉 | ✅ | 多 Tab：Overview, Tasks, Logs, Exec |
| Worker 日志查看 | ✅ | 实时日志，支持刷新 |
| Worker 终端 | ✅ | xterm.js WebSocket 终端 |
| Worker 任务列表 | ✅ | 该 Worker 执行的任务 |

#### Tasks Tab
| 功能 | 状态 | 说明 |
|------|------|------|
| 任务列表 | ✅ | 表格展示，支持分页 |
| 任务搜索 | ✅ | 按 Task ID 搜索 |
| 状态过滤 | ✅ | PENDING/IN_PROGRESS/COMPLETED/FAILED |
| 任务详情 | ✅ | Modal 展示 Input/Output/Error |

#### Settings Tab
| 功能 | 状态 | 说明 |
|------|------|------|
| Scaling 配置 | ✅ | minReplicas, maxReplicas, taskTimeout |
| 环境变量 | ✅ | 添加/删除/保存 key-value |

### 4.3 Tasks 全局页面

| 功能 | 状态 | 说明 |
|------|------|------|
| 任务总览统计 | ✅ | Completed/In Progress/Pending/Failed 数量 |
| 任务列表 | ✅ | 表格展示，支持分页 |
| 任务搜索 | ✅ | 按 Task ID 搜索 |
| 状态过滤 | ✅ | PENDING/IN_PROGRESS/COMPLETED/FAILED |
| Endpoint 过滤 | ✅ | 按 Endpoint 过滤 |
| 取消任务 | ✅ | 取消 PENDING/IN_PROGRESS 任务 |
| 任务详情 | ✅ | Modal 展示完整任务信息 |

---

## 5. 核心流程实现状态

| 流程 | 状态 | 说明 |
|------|------|------|
| 用户认证（JWT Cookie） | ✅ | 从主站获取 token，自动解析用户信息 |
| 集群注册 | ✅ | Waverless 主动注册到 Portal |
| 集群心跳 | ⚠️ | 结构已实现，需 Waverless 配合 |
| 创建 Endpoint | ✅ | 自动选择集群，生成 physicalName，调用 Waverless |
| 任务提交 | ✅ | 透传到 Waverless |
| 任务状态查询 | ✅ | 透传到 Waverless |
| Worker 计费 | ❌ | 未实现 |
| 余额扣费 | ❌ | 未实现 |
| 对账机制 | ❌ | 未实现 |

---

## 6. 技术栈

### 6.1 后端
- Go + Gin 框架
- GORM (MySQL)
- JWT 认证
- Gorilla WebSocket

### 6.2 前端
- React + TypeScript
- Vite 构建
- Ant Design 组件库
- ECharts 图表
- xterm.js 终端
- React Query 数据管理

---

## 7. 下一步计划

### 7.1 计费模块（优先级高）

1. 创建计费相关数据库表
2. 实现 Worker 生命周期 Webhook
3. 实现定时计费任务
4. 实现余额查询和扣费 API
5. 实现充值功能

### 7.2 功能完善

1. 集群智能调度算法
2. 集群健康检查定时任务
3. 余额预警机制
4. 对账机制

### 7.3 UI 优化

1. Dashboard 总览页
2. 计费页面
3. 用户设置页面

---

## 8. 文件结构

```
waverless-portal/
├── cmd/
│   └── main.go                 # 入口
├── app/
│   ├── handler/                # HTTP 处理器
│   │   ├── endpoint_handler.go
│   │   ├── monitoring_handler.go
│   │   ├── admin_handler.go
│   │   └── user_handler.go
│   ├── middleware/             # 中间件
│   │   └── auth.go
│   └── router/                 # 路由
│       └── router.go
├── internal/
│   ├── config/                 # 配置
│   ├── service/                # 业务逻辑
│   │   ├── endpoint_service.go
│   │   └── cluster_service.go
│   └── jobs/                   # 定时任务
│       └── endpoint_sync.go
├── pkg/
│   ├── store/mysql/            # 数据库
│   │   └── model/
│   │       ├── endpoint.go
│   │       ├── cluster.go
│   │       └── spec.go
│   └── waverless/              # Waverless 客户端
│       └── client.go
├── migrations/                 # 数据库迁移
│   └── 001_initial_schema.sql
├── web-ui/                     # 前端
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Endpoints.tsx
│   │   │   ├── EndpointDetail.tsx
│   │   │   ├── Tasks.tsx
│   │   │   └── admin/
│   │   ├── components/
│   │   │   └── Terminal.tsx
│   │   ├── api/
│   │   │   └── client.ts
│   │   └── styles/
│   │       └── portal.css
│   └── vite.config.ts
└── docs/
    ├── PORTAL_DESIGN.md        # 设计文档
    └── IMPLEMENTATION_STATUS.md # 本文档
```

---

## 9. 已知问题

1. **physicalName 大小写**: 已修复，使用 `strings.ToLower()` 确保 k8s 兼容
2. **WebSocket 代理**: 已配置 vite `ws: true`
3. **任务 API 路径**: waverless 使用 `/v1/tasks` 而非 `/api/v1/tasks`

---

## 10. 参考文档

- [PORTAL_DESIGN.md](./PORTAL_DESIGN.md) - 完整设计文档
- [Waverless API 文档](../waverless/docs/) - Waverless API 参考
