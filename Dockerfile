FROM golang:1.22 AS builder
WORKDIR /app
COPY code/ /app/code/
RUN cd /app/code && go mod download && go build -o /app/platform ./cmd/platform

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/platform /app/platform
COPY configs /app/configs
ENV JUYU_CONFIG=/app/configs/agents.yaml
EXPOSE 8080
CMD ["/app/platform"]
