# Building stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /app/main ./cmd/main.go

# Execution stage
FROM alpine:latest

WORKDIR /root/
# Copy the built application from the build stage
COPY --from=builder /app/main .
# Expose port
EXPOSE 9080
# Command to run the application
CMD ["./main"]