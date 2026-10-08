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

# Kopiera källkod
COPY . .

# Uppdatera dependencies
RUN go mod tidy

# Bygg binären
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plan-api ./cmd/server/main.go

# ==============================
# Runtime Stage (minimal, säker)
# ==============================
FROM alpine:3.19 AS production

WORKDIR /app

# Installera ca-certificates för HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Kopiera binären från build-steget
COPY --from=build /app/plan-api .

# Exponera port
EXPOSE 8080

# Miljövariabler
ENV GIN_MODE=release
ENV AUTH_DISABLED=false

# Startkommando
CMD ["./plan-api"]
