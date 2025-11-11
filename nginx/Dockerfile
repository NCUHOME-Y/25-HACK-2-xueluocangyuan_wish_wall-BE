# 1. 使用官方的 Nginx (alpine) 镜像
FROM nginx:alpine

# 2. 删除 Nginx 自带的默认配置
RUN rm /etc/nginx/conf.d/default.conf

# 3. 把 nginx.conf 复制到 Nginx 的配置目录中
COPY nginx.conf /etc/nginx/conf.d/default.conf