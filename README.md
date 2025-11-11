# 🚀 Wish Wall 后端服务 (25-HACK-2-xueluocangyuan_wish_wall-BE) 🚀
2025 Hackweek 第2组  代码别和我作队 雪落藏愿许愿墙  后端仓库

一个基于 Go + Gin + GORM 的许愿墙与社交互动平台后端服务。

## 目录
- [Wish Wall 后端服务 (25-HACK-2-xueluocangyuan\_wish\_wall-BE) 🚀](#wish-wall-后端服务-25-hack-2-xueluocangyuan_wish_wall-be-)
- [🚀 Wish Wall 后端服务 (25-HACK-2-xueluocangyuan\_wish\_wall-BE) 🚀](#-wish-wall-后端服务-25-hack-2-xueluocangyuan_wish_wall-be-)
  - [目录](#目录)
  - [📝 项目简介](#-项目简介)
  - [📁 文件目录说明](#-文件目录说明)
  - [🧭 上手指南](#-上手指南)
    - [开发前的配置要求](#开发前的配置要求)
    - [安装步骤](#安装步骤)
    - [开发的架构](#开发的架构)
    - [部署](#部署)
  - [📚 API 接口文档](#-api-接口文档)
    - [公开接口 (无需认证)](#公开接口-无需认证)
    - [认证接口 (需要 `Authorization: Bearer <token>`)](#认证接口-需要-authorization-bearer-token)
  - [🛠️ 使用到的框架](#️-使用到的框架)
  - [📦 版本控制](#-版本控制)
  - [🎉 鸣谢](#-鸣谢)

---

## 📝 项目简介
雪落藏愿 (Wish Wall) 是一个帮助用户发布愿望、进行公开/私密分享的应用。用户可以：

- 创建带标签的愿望，并选择公开或私密
- 发布公开愿望并获得他人的点赞与评论
- 查看自己的“个人星河”（私密愿望列表）
- 自定义个人昵称和系统头像
- 发布内容时自动进行 AI 内容审核

---

## 📁 文件目录说明

```
├── cmd/myapp/main.go     # Go 应用主入口
├── internal/
│   ├── app/
│   │   ├── handler/      # HTTP 处理器 (业务逻辑)
│   │   ├── model/        # GORM 数据库模型
│   │   └── service/      # 第三方服务 (如 AI 审核)
│   ├── middleware/       # Gin 中间件 (如 JWT 鉴权)
│   ├── pkg/              # 内部公共包 (数据库, 日志, JWT工具, Seeder)
│   └── router/           # 路由配置
├── nginx/
│   ├── Dockerfile        # Nginx 镜像配置文件
│   └── nginx.conf        # Nginx 反向代理配置
├── .env                  # (需手动创建) 环境变量，包含所有密钥
├── .gitignore            # Git 忽略配置
├── docker-compose.yml    # Docker 编排文件 (Nginx, Go App, MySQL)
├── Dockerfile            # Go 应用的 Docker 镜像配置文件
├── go.mod                # Go 模块依赖
└── README.md             # 你正在阅读的文档
```

---

## 🧭 上手指南
本项目为前后端分离的后端部分，所有 API 均通过 `docker-compose` 启动，并通过 Nginx 容器统一暴露在 `80` 端口。

### 开发前的配置要求
- **Go**: `1.25.1` 或更高版本
- **Docker**: 最新稳定版
- **Docker Compose**: 最新稳定版
- **AI 服务**: 需要一个 [Silicon Flow](https://siliconflow.cn/) 账号以获取 API Key。

### 安装步骤
1.  在 [Silicon Flow](https://siliconflow.cn/) 注册并获取一个免费的 API Key。

2.  克隆本项目
    ```bash
    git clone [https://github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE.git](https://github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE.git)
    cd 25-HACK-2-xueluocangyuan_wish_wall-BE
    ```
3.  创建并配置 `.env` 文件
    在项目根目录（`docker-compose.yml` 所在位置）创建一个 `.env` 文件。

    **注意：** `.gitignore` 已经配置为忽略此文件，你的密钥不会被上传。

    请填入以下内容 (并替换为你自己的密码和密钥)：
    ```env
    # 数据库连接字符串 (供 Go 应用在 Docker 内部连接)
    MYSQL_DSN="root:your_password@tcp(db:3306)/wish_wall?charset=utf8mb4&parseTime=True&loc=Local"

    # GIN 模式 (debug 或 release)
    GIN_MODE="debug"

    # JWT 密钥 (请修改为一个复杂的随机字符串)
    JWT_SECRET="my_strong_secret_key!"

    # ----------------------------------------------------
    # 以下配置供 docker-compose 启动 MySQL 容器使用
    MYSQL_ROOT_PASSWORD=your_password
    MYSQL_DATABASE=wish_wall
    # ----------------------------------------------------

    # 应用状态 (V1/V2 切换)
    ACTIVE_ACTIVITY="v1"

    # 你的 Silicon Flow API Key
    SILICONFLOW_API_KEY=sk-your-real-api-key-here

    # (可选) 用于本地测试的数据库 DSN
    MYSQL_TEST_DSN="root:your_password@tcp(127.0.0.1:3307)/wish_wall_test?charset=utf8mb4&parseTime=True&loc=Local"
    ```

4.  使用 Docker Compose 启动服务
    ```bash
    docker-compose up --build
    ```
    服务启动后，API 将在 `http://localhost:80/api` 可访问。

### 开发的架构
本项目采用**容器化三层架构**：

- **反向代理 (Nginx):** 作为应用的统一入口，监听 `80` 端口，并将所有 `/api` 请求反向代理到 Go 应用的 `8080` 端口。
- **后端应用 (Go/Gin):** 运行在 `qpp` 容器中。负责处理所有业务逻辑、JWT 鉴权、数据库操作和 AI 审核。
- **数据库 (MySQL):** 运行在 `db` 容器中，通过 Docker 的内部网络 (`wish-network`) 与 Go 应用通信，数据通过 `volumes` 持久化在宿主机。

Go 应用内部遵循：`router` -> `middleware` -> `handler` -> `model` 的标准分层。

### 部署
本项目为部署而生。部署到生产环境（如云服务器）的步骤如下：

1.  在服务器上安装 `docker` 和 `docker-compose`。
2.  `git clone` 你的仓库。
3.  **安全提示：** **请勿** 上传你本地的 `.env` 文件。请在服务器上手动创建一个全新的 `.env` 文件，并填入**生产环境专用**的数据库密码、`JWT_SECRET` 和 `SILICONFLOW_API_KEY`。
4.  **重要：** 在生产环境的 `.env` 文件中，将 `GIN_MODE` 设置为 `release`。这将关闭 `debug` 日志并停止数据填充。
5.  在后台构建并启动服务：
    ```bash
    docker-compose up -d --build
    ```

---

## 📚 API 接口文档

(以下路由基于 `ACTIVE_ACTIVITY="v1"` 模式)

### 公开接口 (无需认证)
| 路径 | 方法 | 描述 |
| :--- | :--- | :--- |
| `/api/register` | `POST` | 用户注册 |
| `/api/login` | `POST` | 用户登录 |
| `/api/app-state` | `GET` | 获取应用状态 (V1/V2) |
| `/api/wishes/:wishId/comments` | `GET` | 列出某个愿望的评论 |
| `/api/wishes/:wishId/likes` | `GET` | 列出某个愿望的点赞 |
| `/api/test-ai` | `POST` | (测试用) AI 内容审核 |

### 认证接口 (需要 `Authorization: Bearer <token>`)
| 路径 | 方法 | 描述 |
| :--- | :--- | :--- |
| `/api/user/me` | `GET` | 获取当前用户信息 |
| `/api/user` | `PUT` | 更新当前用户信息 |
| `/api/wishes/:id/like` | `POST` | (推荐) 点赞/取消点赞愿望 |
| `/api/comments` | `POST` | 创建新评论 |
| `/api/comments/:id` | `PUT` | 编辑自己的评论 |
| `/api/comments/:id` | `DELETE` | 删除自己的评论 (或管理员) |
| `/api/likes` | `POST` | (V1 兼容) 点赞 |
| `/api/likes/:id` | `DELETE` | (V1 兼容) 取消点赞 |
| `/api/wishes` | `POST` |  发布新愿望 |
| `/api/wishes/me` | `GET` |  获取个人愿望 |
| `/api/wishes/:id` | `DELETE` |  删除愿望 |

---

## 🛠️ 使用到的框架
| 框架 | 描述 |
| :--- | :--- |
| [Gin](https://github.com/gin-gonic/gin) | Web 框架 |
| [GORM](https://gorm.io/) | ORM 工具 |
| [go-openai](https://github.com/sashabaranov/go-openai) | OpenAI / Silicon Flow SDK |
| [jwt-go (v5)](https://github.com/golang-jwt/jwt) | JWT 鉴权 |
| [Zap](https://github.com/uber-go/zap) | 高性能日志 |
| [godotenv](https://github.com/joho/godotenv) | 环境变量加载 |
| [Docker](https://www.docker.com/) | 容器化 |

---

## 📦 版本控制
该项目使用 Git 进行版本管理。

---

## 🎉 鸣谢
- 2025 Hackweek 第2组所有成员
- 所有依赖库的作者
