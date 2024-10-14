# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Set GOPROXY to direct to bypass Go module proxy
ENV GOPROXY=direct

# Install git to fetch modules that require git
RUN apk add --no-cache git

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Run go mod tidy and download dependencies
RUN go mod tidy
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz


# Run stage
FROM alpine:3.13
WORKDIR /app

# Copy only the built binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration


# Expose the application port
EXPOSE 8080

# Run the application
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
# ENTRYPOINT 指定容器启动时运行的主程序，这里设置为 `/app/start.sh`。
# 它会在容器启动时自动执行，用于进行一些初始化任务或启动相关服务。
# 容器在启动时会首先执行 ENTRYPOINT 指定的命令，任何传递给容器的命令行参数都会作为该命令的参数。