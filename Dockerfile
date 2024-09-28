# Build stage
FROM golang:1.16-alpine3.13 AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum 文件，以便下载依赖项
#COPY go.mod go.sum ./

# 下载依赖项并缓存
#RUN go mod download

COPY . .
RUN go build -o main main.go

#Run stage
#FROM alpine:3.13
#WORKDIR /app
#COPY --from=builder /app/main .
EXPOSE 8080
CMD [ "/app/main" ]

