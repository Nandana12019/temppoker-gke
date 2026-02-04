# Stage 1: Build the binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Enable cross-compilation for amd64
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o poker-backend ./cmd/server

# Stage 2: Final lean image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/poker-backend .
EXPOSE 8080
CMD ["./poker-backend"]