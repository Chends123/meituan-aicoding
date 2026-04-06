# 美团评价 AI 分析看板

基于 **Vue2 + Go + Gemini** 构建的餐饮商家评价智能分析系统，支持评价列表浏览、近 7 日评分趋势展示，以及由大模型驱动的聚合分析与差评回复建议，全程通过 **SSE 流式**输出 AI 结果。

---

## 功能概览

| 功能 | 说明 |
|------|------|
| 评价列表 | 支持全部 / 好评 / 差评 Tab 筛选，低于 3 分自动高亮 |
| 游标分页 | 按 `created_at + id` 双字段游标翻页，性能稳定 |
| 趋势图 | 近 7 日每日平均评分折线图，缺口日期自动补全 |
| AI 聚合分析 | 对当前 Tab 下评价批量分析，输出总结、正负向关键词、情绪评分、优化建议 |
| AI 差评回复 | 针对单条差评，流式生成商家专业回复话术 |
| 本地缓存 | AI 分析结果写入 LocalStorage，避免重复请求 |

---

## 技术栈

### 前端

- **Vue 2.7** + Element UI 2.x
- **ECharts 5** 趋势图
- **Axios** HTTP 请求
- **EventSource (SSE)** 流式消费 AI 结果

### 后端

- **Go 1.26** + **Gin** Web 框架
- **GORM** + **MySQL 8.0** 数据持久化
- **google.golang.org/adk**（adk-go）+ **Gemini Flash** 大模型调用
- **SSE** 结构化事件流推送
- **godotenv** `.env` 环境变量管理

---

## 项目结构

```
meituan-aicoding/
├── .env                        # 环境变量（API Key、数据库配置）
├── docker-compose.yml          # MySQL 一键启动
├── plan.md                     # 项目进度记录
├── backend/
│   ├── cmd/server/main.go      # 服务入口
│   ├── configs/config.yaml     # 配置文件
│   ├── review.sql              # 数据库建表 + 示例数据
│   └── internal/
│       ├── ai/                 # AI 客户端、Prompt 模板、SSE 流解析
│       ├── api/                # Handler、Router、DTO
│       ├── config/             # 配置结构体
│       ├── model/              # GORM 数据模型
│       ├── repository/         # 数据库查询层
│       ├── service/            # 业务逻辑层
│       └── pkg/                # 通用工具（db、sse、logger、response）
└── frontend/
    ├── src/
    │   ├── views/              # ReviewDashboard 主页面
    │   ├── components/         # AIAnalysisPanel、ReviewList、TrendChart 等
    │   ├── api/                # review.js、ai.js 接口封装
    │   └── utils/              # storage.js、sse.js、format.js 工具
    └── public/
```

---

## 本地运行

### 环境依赖

- Go >= 1.21
- Node.js >= 16
- Docker & Docker Compose

### 1. 克隆仓库

```bash
git clone https://github.com/Chends123/meituan-aicoding.git
cd meituan-aicoding
```

### 2. 配置环境变量

复制并编辑根目录 `.env`，填入你的 Google API Key：

```env
APP_NAME=meituan-review-ai
APP_PORT=8081

MYSQL_HOST=127.0.0.1
MYSQL_PORT=13306
MYSQL_DATABASE=meituan_review_ai
MYSQL_USERNAME=root
MYSQL_PASSWORD=123456

GOOGLE_API_KEY=<你的 Google AI Studio API Key>
AI_MODEL=gemini-2.0-flash
```

> 获取 API Key：[https://aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)

### 3. 启动 MySQL

```bash
docker-compose up -d
```

Docker Compose 会自动创建数据库、建表并导入示例评价数据（`backend/review.sql`）。

验证数据库连通性：

```bash
docker exec -it meituan-aicoding-mysql mysql -uroot -p123456 -e "SELECT COUNT(*) FROM meituan_review_ai.reviews;"
```

### 4. 启动后端

```bash
cd backend
go run ./cmd/server/main.go
```

服务默认监听 `http://localhost:8081`，可通过以下接口验证：

```
GET http://localhost:8081/ping
```

### 5. 启动前端

```bash
cd frontend
npm install
npm run serve
```

前端默认运行在 `http://localhost:8080`，访问即可看到评价看板。

---

## API 接口说明

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/reviews` | 评价列表，支持 `tab`、`cursor`、`limit` 参数 |
| `GET` | `/api/v1/reviews/trends` | 近 7 日评分趋势 |
| `GET` | `/api/v1/ai/reviews/analyze/stream` | AI 聚合分析（SSE 流式） |
| `GET` | `/api/v1/ai/reviews/:id/reply/stream` | 单条差评 AI 回复建议（SSE 流式） |

### SSE 事件类型（AI 分析）

| 事件名 | 数据说明 |
|--------|----------|
| `meta` | 分析元信息（评价数量、Tab 类型） |
| `positive_keywords` | 正向关键词列表 |
| `negative_keywords` | 负向关键词列表 |
| `sentiment_score` | 情绪评分（0–100） |
| `suggestions` | 商家优化建议列表 |
| `summary` | 一句话总结 |
| `reply_delta` | 差评回复内容（流式分片） |
| `done` | 流式结束标记 |

---

## 数据库说明

表名：`reviews`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT | 主键，自增 |
| `username` | VARCHAR(64) | 用户昵称 |
| `score` | TINYINT | 评分（1–5 分） |
| `content` | TEXT | 评价正文 |
| `created_at` | DATETIME | 评价时间 |

数据库建表及初始化脚本见 `backend/review.sql`，Docker Compose 启动时自动执行。

---

## 注意事项

- AI 功能依赖有效的 `GOOGLE_API_KEY`，未配置时 AI 接口将返回错误。
- 模型输出依赖 Prompt 格式约束，偶发输出不规范时解析可能失败（待补兜底处理）。
- 前端开发模式下已配置代理，所有 `/api` 请求转发至 `http://localhost:8081`。
