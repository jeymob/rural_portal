FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rural-api ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/rural-api .
COPY .env* .
COPY internal/delivery/http/templates ./internal/delivery/http/templates
EXPOSE 8080
CMD ["./rural-api"]