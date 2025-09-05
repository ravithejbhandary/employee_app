# Stage 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum from root
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./backend/

# Build server binary
RUN go build -o server ./backend/main.go ./backend/server.go

# Stage 2: Final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]
