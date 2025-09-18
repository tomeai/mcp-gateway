FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM node:lts-bookworm-slim AS node

FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx && \
    ln -s /usr/local/bin/node /usr/local/bin/nodejs

COPY --from=builder /app/build/mcp-gateway /mcp-gateway
ENTRYPOINT ["/mcp-gateway"]
CMD ["--port", "8000"]