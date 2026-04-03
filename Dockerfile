FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auction cmd/auction/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/auction .
COPY --from=builder /app/cmd/auction/.env ./cmd/auction/.env

EXPOSE 8080

ENTRYPOINT ["./auction"]
