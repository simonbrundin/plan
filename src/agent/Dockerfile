# Agent execution service.
# Build context is the plan repository root.
# Go-modulen ligger i src/ och omfattar både api/ och agent/ - därför
# kopieras go.mod/go.sum från src/ och koden från src/agent/.
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates bash
WORKDIR /app/src

COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/agent/ ./agent/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agent ./agent
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agentworker ./agent/cmd/agentworker

FROM alpine:3.19
RUN apk add --no-cache ca-certificates bash git github-cli python3
WORKDIR /app

COPY --from=builder /app/agent ./agent
COPY --from=builder /app/agentworker ./agentworker
COPY src/agent/commands ./src/agent/commands
COPY README.md ./README.md

ENV AGENT_DOCS_DIR=/app
ENV PROJECT_ROOT=/app
USER 65532:65532
EXPOSE 8080
CMD ["./agent"]
