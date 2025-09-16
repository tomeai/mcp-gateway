# =========================
# Build Stage
# =========================
FROM golang:1.24-alpine AS builder

# 安装依赖
RUN apk add --no-cache git ca-certificates upx

WORKDIR /app

# 拷贝 go.mod 和 go.sum 先下载依赖（缓存优化）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /wemcp-gateway .

# （可选）压缩二进制，减少体积
RUN upx --lzma /wemcp-gateway || true

# =========================
# Runtime Stage
# =========================
FROM gcr.io/distroless/base

WORKDIR /

# 复制构建产物
COPY --from=builder /wemcp-gateway /wemcp-gateway

EXPOSE 8080

ENTRYPOINT ["/wemcp-gateway"]
CMD ["start"]
