# 项目计划

## 当前状态

当前项目已经完成“方案设计 + 前后端初始化 + 核心后端接口开发 + adk-go 流式 AI 接入”的阶段性工作，已经具备继续联调和演示的基础。

当前结论：

- 前端项目骨架已完成，且已通过构建验证。
- 后端项目骨架已完成，且已通过编译验证。
- AI 分析能力已从本地启发式逻辑切换为真实 `adk-go + Gemini Flash` 流式调用。
- 根目录 `.env` 已建立，AI API Key 和数据库配置均通过环境变量加载。
- 当前主要待办是：本地 MySQL 环境启动、前后端联调、OpenAPI 文档补充、流式体验收口。

## 已完成进度

### Milestone 1：方案确认与文档沉淀

- [x] 输出整体实现思路
- [x] 设计整体功能架构图
- [x] 设计评价列表与 AI 分析流程图
- [x] 设计前后端树形目录结构
- [x] 设计数据库表结构
- [x] 设计 RESTful API 草案
- [x] 设计 SSE 事件协议
- [x] 编写并落地 `AGENT.md`

### Milestone 2：项目初始化

- [x] 初始化 `frontend/` Vue2 工程基础结构
- [x] 安装前端基础依赖：`Vue2`、`ElementUI`、`Axios`、`ECharts`
- [x] 初始化 `backend/` Go 工程与 `go.mod`
- [x] 初始化 Gin / GORM / 配置目录结构
- [x] 新增根目录 `.env`
- [x] 新增 `docker-compose.yml`
- [x] 新增 `backend/review.sql`

### Milestone 3：前端页面骨架开发

- [x] 完成 `ReviewDashboard` 双栏页面骨架
- [x] 完成评价列表 Tab 组件
- [x] 完成评价卡片组件与低分高亮样式
- [x] 完成 AI 分析面板骨架
- [x] 完成关键词展示组件
- [x] 完成近 7 日趋势图组件
- [x] 完成前端 API 封装
- [x] 完成 LocalStorage 缓存工具
- [x] 完成 EventSource SSE 工具封装
- [x] 完成前端生产构建验证 `npm run build`

### Milestone 4：后端基础接口开发

- [x] 完成 `.env` 环境变量加载
- [x] 完成 MySQL 配置读取
- [x] 完成 GORM 数据库连接初始化
- [x] 完成 `reviews` 数据模型定义
- [x] 完成评价列表查询仓储层
- [x] 完成好评 / 差评 / 全部 Tab 筛选
- [x] 完成按 `created_at + id` 的游标分页
- [x] 完成近 7 日评分趋势查询
- [x] 完成趋势图缺口日期补全逻辑
- [x] 完成后端路由注册与 Handler 层
- [x] 完成后端编译验证 `go build ./...`

### Milestone 5：AI 能力接入

- [x] 接入 `adk-go`
- [x] 接入 `Gemini Flash` 模型配置
- [x] 完成 AI 聚合分析流式接口
- [x] 完成单条差评回复建议流式接口
- [x] 完成 AI Prompt 模板拆分
- [x] 完成模型 JSON 输出解析器
- [x] 完成结构化 SSE 事件输出：
- [x] `meta`
- [x] `positive_keywords`
- [x] `negative_keywords`
- [x] `sentiment_score`
- [x] `suggestions`
- [x] `summary`
- [x] `reply_delta`
- [x] `done`

## 当前已实现的核心能力

### 前端

- 评价看板双栏布局
- 评价列表展示
- `全部 / 好评 / 差评` Tab 切换
- 低于 3 分自动高亮
- AI 分析面板结构化展示
- 近 7 日评分趋势图
- AI 分析结果本地缓存
- 已接好后端 SSE 事件消费逻辑

### 后端

- `GET /api/v1/reviews`
- `GET /api/v1/reviews/trends`
- `GET /api/v1/ai/reviews/analyze/stream`
- `GET /api/v1/ai/reviews/:id/reply/stream`
- `.env` 方式加载 `GOOGLE_API_KEY`
- 使用 `adk-go runner + llmagent + gemini` 进行真实模型流式分析
- 将模型结果转换成前端可直接消费的结构化 SSE 事件

## 当前待完成事项

### Milestone 6：联调与运行验证

- [ ] 启动 MySQL 并导入 `backend/review.sql`
- [ ] 本地启动后端服务并验证数据库连接
- [ ] 本地启动前端服务并完成前后端联调
- [ ] 验证评价列表、分页、Tab 筛选是否与前端表现一致
- [ ] 验证趋势图真实数据展示
- [ ] 验证 AI 聚合分析 SSE 流式输出
- [ ] 验证单条差评回复建议 SSE 输出

### Milestone 7：AI 体验收口

- [ ] 前端消费 `model_delta` 事件，做更细粒度的流式展示
- [ ] 提升 AI 结果异常处理
- [ ] 补充模型输出 JSON 失败时的兜底策略
- [ ] 优化 AI 分析完成前后的加载态与错误态

### Milestone 8：文档与交付

- [ ] 补充后端 OpenAPI 文档
- [ ] 补充本地运行说明
- [ ] 补充联调说明
- [ ] 补充 `.env` 字段说明
- [ ] 补充 Demo 演示步骤

## 当前风险与注意事项

- 当前 AI 分析依赖真实 `GOOGLE_API_KEY`，未配置时 AI 接口会直接报错。
- 当前模型输出依赖 Prompt 约束为 JSON，如果模型偶发输出不规范，解析会失败，需要补充兜底处理。
- 前端当前已能接结构化 SSE 结果，但对 `model_delta` 的细粒度展示还没有完全用起来。
- OpenAPI 文档目前尚未生成。
- MySQL 环境尚未由当前会话实际启动和验证。

## 建议下一步顺序

1. 先由你完成 MySQL 环境搭建与数据导入。
2. 我继续负责启动后端、检查接口返回、完成前后端联调。
3. 联调通过后，我补 AI 流式展示优化与异常兜底。
4. 最后补 OpenAPI 文档和运行说明，收口成完整 Demo。

## 关键文件

- `AGENT.md`
- `plan.md`
- `.env`
- `docker-compose.yml`
- `frontend/src/views/ReviewDashboard.vue`
- `frontend/src/components/AIAnalysisPanel.vue`
- `backend/cmd/server/main.go`
- `backend/internal/service/ai_service.go`
- `backend/internal/service/review_service.go`
- `backend/internal/repository/review_repository.go`
- `backend/internal/ai/client.go`
- `backend/internal/ai/prompt.go`
- `backend/internal/ai/stream_parser.go`
- `backend/review.sql`
