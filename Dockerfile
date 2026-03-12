FROM golang:1.22 AS builder
WORKDIR /app
COPY code/ /app/code/
RUN cd /app/code && go mod download && go build -o /app/platform ./cmd/platform

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/platform /app/platform
COPY configs /app/configs
ENV JUYU_CONFIG=/app/configs/agents.yaml
ENV JUYU_STORE=postgres
ENV JUYU_PG_DSN="host=postgres port=5432 user=postgres password=postgres dbname=juyu sslmode=disable timezone=Asia/Shanghai"
EXPOSE 8080
CMD ["/app/platform"]
