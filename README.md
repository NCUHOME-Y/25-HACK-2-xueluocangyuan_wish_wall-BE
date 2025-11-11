
# Wish Wall 后端服务 (25-HACK-2-xueluocangyuan_wish_wall-BE) 🚀
2025 Hackweek 第2组  代码别和我作队 雪落藏愿许愿墙  后端仓库


🚀 Wish Wall 后端 (xueluocangyuan_wish_wall-BE)
欢迎使用“雪落藏愿”许愿墙后端服务。这是一个基于 Go (Gin + GORM) 构建的轻量级、容器化的后端应用，旨在提供一个发布、浏览、评论和点赞愿望的平台，并集成了 AI 内容审核功能。

✨ 主要功能
用户认证: 基于 JWT 的注册、登录和鉴权。

愿望管理: 发布、删除、分页获取公共愿望和个人愿望。

互动系统: 对愿望进行点赞/取消点赞，发布/删除/编辑评论。

AI 内容审核: 在发布内容时，自动调用 Silicon Flow API 对违规内容（色情、暴力、仇恨言论等）进行审查。

动态路由: 支持通过环境变量 ACTIVE_ACTIVITY 切换 V1 / V2 版本的 API 路由。

数据填充: 在非生产模式下启动时，自动填充机器人用户和示例愿望，便于测试。

🛠️ 技术栈
后端: Go, Gin (Web 框架)

数据库: MySQL, GORM (ORM)

容器化: Docker, Docker Compose

反向代理: Nginx

日志: Zap Logger

AI 服务: go-openai (Silicon Flow)

🚀 快速开始 (推荐)
本项目已完全容器化，使用 Docker Compose 是最简单、最推荐的启动方式。

克隆仓库

Bash

git clone <your-repo-url>
cd 25-HACK-2-xueluocangyuan_wish_wall-BE
创建配置文件 在项目根目录（与 docker-compose.yml 同级）创建一个名为 .env 的文件。

复制 hack-2/25-HACK-2-xueluocangyuan_wish_wall-BE/.env 的内容到你新创建的 .env 文件中。

重要： .env 文件已被 .gitignore 忽略，不会提交到 Git 仓库，请妥善保管你的密钥。

启动服务 在项目根目录运行以下命令：

Bash

docker-compose up --build
Docker Compose 将会自动完成以下所有工作：

构建 Go 应用镜像（基于 Dockerfile）。

构建 Nginx 镜像（基于 nginx/Dockerfile）。

启动 Go 应用、Nginx 和 MySQL 数据库三个容器。

Go 应用会自动连接数据库并执行数据库迁移 (AutoMigrate)。

Go 应用会执行数据填充 (Seeder)。

访问服务 启动成功后，API 服务将通过 Nginx 暴露在 http://localhost:80。

API 根地址: http://localhost:80/api

数据库 (本地): mysql://root:Missyousomuch0@127.0.0.1:3307/wish_wall




