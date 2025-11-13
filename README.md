# Wish Wall 后端服务 (25-HACK-2-xueluocangyuan_wish_wall-BE) 🚀
2025 Hackweek 第2组 代码别和我作队 雪落藏愿许愿墙 后端仓库


## 目录
- [Wish Wall 后端服务 (25-HACK-2-xueluocangyuan\_wish\_wall-BE) 🚀](#wish-wall-后端服务-25-hack-2-xueluocangyuan_wish_wall-be-)
  - [目录](#目录)
  - [📝 项目简介](#-项目简介)
    - [✨ 核心功能](#-核心功能)
  - [📁 文件目录说明](#-文件目录说明)
  - [🧭 上手指南](#-上手指南)
    - [开发前的配置要求](#开发前的配置要求)
    - [安装与启动](#安装与启动)
    - [开发的架构](#开发的架构)
    - [部署](#部署)
  - [🔑 环境变量配置 (.env)](#-环境变量配置-env)
    - [🧪 运行测试](#-运行测试)
    - [📚 API 接口文档](#-api-接口文档)
      - [📦 API 响应格式](#-api-响应格式)
      - [公开接口 (无需认证)](#公开接口-无需认证)
      - [认证接口 (需要 `Authorization: Bearer <token>`)](#认证接口-需要-authorization-bearer-token)
      - [API 详情示例](#api-详情示例)
    - [🛠️ 使用到的框架](#️-使用到的框架)
    - [📦 版本控制](#-版本控制)
    - [🎉 鸣谢](#-鸣谢)


---

## 📝 项目简介
雪落藏愿 (Wish Wall) 是一个帮助用户发布愿望、进行公开/私密分享的应用。本项目是其后端服务，负责处理所有业务逻辑、数据存储和第三方服务集成。

### ✨ 核心功能
- **用户认证**: 基于 JWT (HS256) 的注册和登录流程。
- **愿望管理**: 用户可以创建、删除、查看自己的私密愿望和公共愿望列表。
- **社交互动**: 支持对公共愿望进行点赞、取消点赞和发表评论。
- **AI 内容审核**: 集成 [Silicon Flow](https://siliconflow.cn/) API，在用户注册（昵称）、发布愿望、发表评论时自动进行内容安全审核。
- **动态功能路由**: 通过环境变量 `ACTIVE_ACTIVITY` 控制 API 模式（例如 `v1` 为读写模式，`v2` 为只读模式），参见 `internal/router/router.go`。
- **容器化部署**: 提供完整的 `Dockerfile` 和 `docker-compose.yml`，实现 Nginx、Go 应用、MySQL 数据库的一键启动。
- **数据库填充 (Seeding)**: 在非 `release` 模式下启动时，自动填充机器人用户和愿望数据，便于开发和测试，参见 `internal/pkg/seeder/seeder.go`。
- **角色系统**: 基础的用户角色定义 (如 `user`, `admin`)，用于权限控制（如删除评论），参见 `internal/app/model/user.go`。

---

## 📁 文件目录说明
```
.
├── .env                 # 环境变量 (包含数据库密码、JWT密钥、AI Key)
├── .gitignore           # Git 忽略配置，防止 .env 等敏感文件上传
├── docker-compose.yml   # Docker 编排文件 (一键启动 Nginx, Go App, MySQL)
├── Dockerfile           # Go 应用的 Docker 镜像配置文件 (多阶段构建)
├── go.mod               # Go 模块依赖
├── go.sum               # 依赖校验和
├── README.md            # 项目说明文档
│
├── cmd/
│   └── myapp/
│       └── main.go      # Go 应用主入口 (初始化日志、数据库、路由)
│
├── internal/
│   ├── app/
│   │   ├── handler/     # HTTP 处理器 (Gin 的 Ctx 在这里，负责业务逻辑)
│   │   │   ├── app.go             # (GetAppState, TestAI)
│   │   │   ├── app_test.go
│   │   │   ├── comment.go         # (CreateComment, DeleteComment, ListCommentsByWish)
│   │   │   ├── comment_test.go   
│   │   │   ├── CreatWish.go       # (CreateWish)
│   │   │   ├── DeleteWish.go      # (DeleteWish)
│   │   │   ├── GetMyWish.go       # (GetMyWishes)
│   │   │   ├── GetPublicWish.go   # (GetPublicWishes)
│   │   │   ├── interactions.go    # (CreateCommentAI, CreateReplyAI, GetInteractions)
│   │   │   ├── like.go            # (LikeWish)
│   │   │   ├── like_test.go
│   │   │   ├── main_test.go       # 测试主入口 (Setup/Cleanup)
│   │   │   ├── user.go            # (Register, Login, GetUserMe, UpdateUser)
│   │   │   ├── user_test.go
│   │   │   ├── wishes.go          # (已被拆分到 CreatWish 等文件)
│   │   │   └── wishes_test.go
│   │   │
│   │   ├── model/           # GORM 数据库模型 (Struct 定义)
│   │   │   ├── comment.go    
│   │   │   ├── like.go       
│   │   │   ├── user.go       
│   │   │   └── wish.go       
│   │   │
│   │   ├── repository/      # 数据库操作 
│   │   │   └── user_repo.go
│   │   │
│   │   └── service/         # 第三方服务
│   │       ├── ai_service.go      # AI 内容审核 (CheckContent)
│   │       └── ai_service_test.go
│   │
│   ├── middleware/        # Gin 中间件
│   │   └── auth.go          # CORS, Logger, Recovery, JWT 鉴权
│   │
│   ├── pkg/               # 内部公共包 (与业务逻辑无关的工具)
│   │   ├── database/
│   │   │   └── database.go  # 数据库初始化 (InitDB)
│   │   ├── err/
│   │   │   ├── msg.go     # 统一业务错误码
│   │   │   └── msg_test.go
│   │   ├── logger/
│   │   │   └── logger.go    # Zap 日志初始化
│   │   ├── seeder/
│   │   │   └── seeder.go    # 数据库初始数据填充
│   │   └── util/
│   │       ├── jwt.go     # JWT Token 生成与解析
│   │       └── jwt_test.go
│   │
│   └── router/
│       └── router.go        # 路由配置 (SetupRouter, 挂载所有 API 路由)
│
└── nginx/                   # Nginx 配置 (用于反向代理)
    ├── Dockerfile       # Nginx 镜像的 Dockerfile
    └── nginx.conf       # Nginx 反向代理配置 (将 80 端口转发到 8080)
    
```
---

## 🧭 上手指南
本项目为前后端分离的后端部分，所有 API 均通过 `docker-compose` 启动，并通过 Nginx 容器统一暴露在 `80` 端口。


### 开发前的配置要求
- **Go**: `1.25.1` 或更高版本 (参见 `go.mod` 和 `Dockerfile`)
- **Docker**: 最新稳定版
- **Docker Compose**: 最新稳定版
- **AI 服务**: 需要一个 [Silicon Flow](https://siliconflow.cn/) 账号以获取 API Key。


### 安装与启动
1.  在 [Silicon Flow](https://siliconflow.cn/) 注册并获取一个免费的 API Key。

2.  克隆本项目
    ```bash
    git clone [https://github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE.git](https://github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE.git)
    cd 25-HACK-2-xueluocangyuan_wish_wall-BE
    ```
3.  创建并配置 `.env` 文件
    在项目根目录（`docker-compose.yml` 所在位置）创建一个 `.env` 文件。**请参考下一章节** `🔑 环境变量配置 (.env)` 了解所有必填和可选字段。

4.  使用 Docker Compose 启动服务
    ```bash
    # --build 会强制重新构建镜像
    docker-compose up --build
    ```
    服务启动后，API 将在 `http://localhost:80/api` (由 Nginx 代理) 可访问。MySQL 数据库将暴露在 `http://localhost:3307`。

### 开发的架构
本项目采用**容器化三层架构**：

- **反向代理 (Nginx):** 作为应用的统一入口 (Nginx 容器)，监听 `80` 端口，并将所有请求反向代理到 Go 应用的 `8080` 端口 (参见 `nginx/nginx.conf`)。
- **后端应用 (Go/Gin):** 运行在 `qpp` 容器中。负责处理所有业务逻辑、JWT 鉴权、数据库操作和 AI 审核。
- **数据库 (MySQL):** 运行在 `db` 容器中，通过 Docker 的内部网络 (`wish-network`) 与 Go 应用通信，数据通过 `volumes` (db_data) 持久化在宿主机。

Go 应用内部遵循：`router` -> `middleware` -> `handler` -> `service` / `model` 的标准分层。

### 部署
部署到生产环境（如云服务器）的步骤如下：

1.  在服务器上安装 `docker` 和 `docker-compose`。
2.  `git clone` 你的仓库。
3.  **安全提示：** **请勿** 上传你本地的 `.env` 文件。请在服务器上手动创建一个全新的 `.env` 文件，并填入**生产环境专用**的数据库密码、`JWT_SECRET` 和 `SILICONFLOW_API_KEY`。
4.  **重要：** 在生产环境的 `.env` 文件中，将 `GIN_MODE` 设置为 `release`。这将关闭 `debug` 日志并停止数据填充 (Seeder)。
5.  在后台构建并启动服务：
    ```bash
    docker-compose up -d --build
    ```

---

## 🔑 环境变量配置 (.env)
在项目根目录创建 `.env` 文件。`docker-compose` 和 Go 应用（通过 `godotenv`）都会读取此文件。

```env
# --- Go 应用 (qpp) 和 MySQL (db) 容器共用 ---
# 数据库连接字符串 (供 Go 应用在 Docker 内部连接 db 容器)
MYSQL_DSN="root:your_password@tcp(db:3306)/wish_wall?charset=utf8mb4&parseTime=True&loc=Local"

# 供 docker-compose 启动 MySQL 容器使用
MYSQL_ROOT_PASSWORD=your_password
MYSQL_DATABASE=wish_wall

# --- Go 应用 (qpp) 专用 ---
# GIN 模式 (debug 或 release)。"release" 模式会关闭 seeder、关闭 zap 的 debug 日志
GIN_MODE="debug"

# JWT 密钥 (请修改为一个复杂的随机字符串)
JWT_SECRET="my_strong_secret_key!"

# 应用状态 (V1/V2 切换)
# "v1": 启用所有 API (读写模式)
# "v2" (或其他): 禁用 POST/PUT/DELETE API，变为"只读"模式 (参见 router.go)
ACTIVE_ACTIVITY="v1"

# Silicon Flow API Key (用于 AI 内容审核)
SILICONFLOW_API_KEY=sk-your-real-api-key-here

# Silicon Flow API 基础 URL (可选, 默认为 [https://api.siliconflow.cn/v1](https://api.siliconflow.cn/v1))
# SILICONFLOW_BASE_URL="[https://api.siliconflow.cn/v1](https://api.siliconflow.cn/v1)"

# (可选) 用于本地测试的数据库 DSN (运行 go test 时使用)
MYSQL_TEST_DSN="root:your_password@tcp(127.0.0.1:3307)/wish_wall_test?charset=utf8mb4&parseTime=True&loc=Local"
```


---

### 🧪 运行测试
本项目包含丰富的单元和集成测试 (参见 *_test.go 文件)。确保测试数据库已运行：测试依赖于 MYSQL_TEST_DSN 环境变量。请确保 docker-compose.yml 中 db 服务的 3307:3306 端口映射已开启，并且服务在运行中。确保 .env 文件存在：测试会加载根目录的 .env 文件来获取 MYSQL_TEST_DSN 和 SILICONFLOW_API_KEY。运行测试：在项目根目录运行：Bash
```
go test ./... -v
```
( -v 参数会显示详细的测试输出。)

### 📚 API 接口文档
(以下路由基于 ACTIVE_ACTIVITY="v1" 模式)

#### 📦 API 响应格式

成功响应 (HTTP 200):
```json
{
  "code": 200,
  "message": "成功",
  "data": { ... } // 业务数据
}
```

失败响应 (HTTP 4xx / 5xx):
```json
{
  "code": <int>, // 业务错误码 (参见 internal/pkg/err/msg.go)
  "message": "<string>", // 错误信息
  "data": {
    "error": "<string>" // (可选) 更详细的错误描述
  }
}
```

#### 公开接口 (无需认证)

| 路径                     | 方法 | 描述                         |
| ------------------------ | ---- | ---------------------------- |
| /api/register            | POST | 用户注册 (含 AI 昵称审核)    |
| /api/login               | POST | 用户登录                     |
| /api/app-state           | GET  | 获取应用状态 (V1/V2)         |
| /api/wishes/public       | GET  | 获取公共愿望列表 (可选鉴权)  |
| /api/wishes/:id/comments | GET  | 列出某个愿望的评论           |
| /api/test-ai             | POST | (测试用) AI 内容审核认证接口 |

#### 认证接口 (需要 `Authorization: Bearer <token>`)

| 路径                         | 方法      | 描述                               |
| ---------------------------- | --------- | ---------------------------------- |
| /api/user/me                 | GET       | 获取当前用户信息                   |
| /api/user                    | PUT       | 更新当前用户信息 (含 AI 昵称审核)  |
| /api/wishes                  | POST      | 发布新愿望 (含 AI 内容审核)        |
| /api/wishes/me               | GET       | 获取个人愿望                       |
| /api/wishes/:id              | DELETE    | 删除愿望 (仅限作者)                |
| /api/wishes/:id/like         | POST      | 点赞/取消点赞愿望                  |
| /api/wishes/:id/interactions | GET (V1)  | 获取愿望互动详情                   |
| /api/comments                | POST      | 创建新评论 (含 AI 内容审核)        |
| /api/comments/:id            | DELETE    | 删除自己的评论 (或管理员/愿望作者) |
| /api/comments/reply          | POST (V1) | 回复评论 (含 AI 内容审核)          |

#### API 详情示例

1. 用户注册 (POST /api/register)

请求体:
```json
{
  "username": "1234567890",
  "password": "mypassword123",
  "nickname": "我的昵称"
}
```

成功响应 (200):
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "token": "eyJh...<JWT_TOKEN>...wA",
    "user": {
      "id": 1,
      "username": "1234567890",
      "nickname": "我的昵称",
  "avatar_id": null,
      "role": "user",
      "createdAt": "2025-11-12T07:14:00Z"
    }
  }
}
```

错误响应 (400 - 昵称违规):
```json
{
  "code": 4,
  "message": "参数验证失败",
  "data": {
    "error": "昵称包含不当内容，请修改"
  }
}
```

2. 发布新愿望 (POST /api/wishes)

请求头: Authorization: Bearer <JWT_TOKEN>

请求体:
```json
{
  "content": "我希望期末考试顺利通过！",
  "isPublic": true,
  "background": "default"
}
```

成功响应 (200):
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "wishID": 10,
    "createdAt": "2025-11-12T07:15:00Z"
  }
}
```

错误响应 (400 - 内容违规):
```json
{
  "code": 4,
  "message": "参数验证失败",
  "data": {
    "error": "内容未通过审核"
  }
}
```

3. 点赞愿望 (POST /api/wishes/:id/like)

请求头: Authorization: Bearer <JWT_TOKEN>

请求路径: /api/wishes/10/like

成功响应 (200 - 点赞成功):
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "wishID": 10,
    "likeCount": 1,
    "liked": true
  }
}
```

成功响应 (200 - 取消点赞):
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "wishID": 10,
    "likeCount": 0,
    "liked": false
  }
}
```

---

### 🛠️ 使用到的框架
- Gin Web 框架
- GORM ORM 工具
- go-openai OpenAI / Silicon Flow SDK
- jwt-go (v5) JWT 鉴权
- Zap 高性能日志
- godotenv 环境变量加载
- Docker 容器化
- Testify Go 测试断言库

### 📦 版本控制
该项目使用 Git 进行版本管理。

### 🎉 鸣谢
2025 Hackweek 第2组所有成员  
所有依赖库的作者
