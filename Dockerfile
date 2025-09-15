# STAGE 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tsr .

# STAGE 2: Run
FROM alpine:latest

RUN apk add --no-cache ca-certificates ffmpeg streamlink

WORKDIR /app

COPY --from=builder /app/tsr ./
COPY .env ./

CMD ["./tsr"]