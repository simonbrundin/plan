# ==============================
# Byggcontexten är src/ (modulroten) - se src/.dockerignore.
# Go-modulen ligger i src/ och omfattar både api/ och agent/.
# ==============================
# ==============================
# OpenAPI documentation stage
# ==============================
FROM golang:1.26-alpine AS swag
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
COPY go.mod go.sum ./
RUN go mod download
COPY api/ ./api/
RUN swag init -g api/cmd/server/main.go -o api/docs --parseFuncBody --quiet

# ==============================
# Build Stage
# ==============================
FROM golang:1.26-alpine AS build
WORKDIR /app

# Installera build-beroenden
RUN apk add --no-cache git

# Kopiera go mod-filer
COPY go.mod go.sum ./

# Hämta dependencies
RUN go mod download

# Kopiera källkod och den genererade OpenAPI-specifikationen
COPY api/ ./api/
COPY --from=swag /app/api/docs ./api/docs

# Bygg binären
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plan-api ./api/cmd/server/main.go

# ==============================
# Runtime Stage (minimal, säker)
# ==============================
FROM alpine:3.19 AS production

WORKDIR /app

# Installera ca-certificates för HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Kopiera binären och OpenAPI-specifikationen
COPY --from=build /app/plan-api .
COPY --from=build /app/api/docs ./api/docs

# Scalar läser specen från denna sökväg
ENV DOCS_PATH=/app/api/docs/swagger.json

# Exponera port
EXPOSE 8080

# Miljövariabler
ENV GIN_MODE=release
ENV AUTH_DISABLED=false

# Startkommando
CMD ["./plan-api"]
