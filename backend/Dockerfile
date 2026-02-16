# ==================== Dev Stage ====================
FROM golang:1.24-alpine AS dev

RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

# ==================== Build Stage ====================
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/server

# ==================== Prod Stage ====================
FROM alpine:latest AS prod

WORKDIR /app

COPY --from=build /app/main .

EXPOSE 8080

CMD ["./main"]
