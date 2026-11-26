# ==============================
# OpenAPI documentation stage
# ==============================
FROM golang:1.25-alpine AS swag
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
COPY . .
RUN swag init -g cmd/server/main.go -o docs --parseFuncBody --quiet

# ==============================
# Build Stage
# ==============================
FROM golang:1.25-alpine AS build
WORKDIR /app

# Installera build-beroenden
RUN apk add --no-cache git

# Kopiera go mod-filer
COPY go.mod go.sum ./

# Hämta dependencies
RUN go mod download

# Kopiera källkod och den genererade OpenAPI-specifikationen
COPY . .
COPY --from=swag /app/docs ./docs

# Bygg binären
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plan-api ./cmd/server/main.go

# ==============================
# Runtime Stage (minimal, säker)
# ==============================
FROM alpine:3.19 AS production

WORKDIR /app

# Installera ca-certificates för HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Kopiera binären och OpenAPI-specifikationen
COPY --from=build /app/plan-api .
COPY --from=build /app/docs ./docs

# Scalar läser specen från denna sökväg
ENV DOCS_PATH=/app/docs/swagger.json

# Exponera port
EXPOSE 8080

# Miljövariabler
ENV GIN_MODE=release
ENV AUTH_DISABLED=false

# Startkommando
CMD ["./plan-api"]
