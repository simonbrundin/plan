# Agent execution service.
# Build context is the plan repository root.
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates bash
WORKDIR /app/src/agent

COPY src/agent/go.mod src/agent/go.sum ./
RUN go mod download
COPY src/agent/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agent .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agentworker ./cmd/agentworker

FROM alpine:3.19
RUN apk add --no-cache ca-certificates bash git python3
WORKDIR /app

COPY --from=builder /app/agent ./agent
COPY --from=builder /app/agentworker ./agentworker
COPY src/commands ./src/commands
COPY README.md ./README.md

ENV AGENT_DOCS_DIR=/app
ENV PROJECT_ROOT=/app
USER 65532:65532
EXPOSE 8080
CMD ["./agent"]
