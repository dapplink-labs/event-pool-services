# 使用官方的Golang镜像作为基础镜像
FROM golang:1.25-alpine

# 将当前目录（包含你的Go应用程序）的内容复制到容器中的/build目录
COPY event-pod-services /opt/chooseme-services/event-pod-services

# 复制配置文件（重命名为程序期望的文件名）
COPY event-pod-services-config.local.yaml /opt/chooseme-services/phoenix-services-config.local.yaml

# 创建一个非root用户并使用它来运行应用程序，提高安全性
# RUN adduser -D event-services
# USER event-services

# 设置可执行权限
RUN chmod -R 755 /opt/chooseme-services

# 设置工作目录
WORKDIR /opt/chooseme-services

# 暴露端口8080，这是你的Go应用程序监听的端口
EXPOSE 8080

# 设置容器启动时执行的命令，即运行你的Go应用程序
CMD ["./event-pod-services","api"]