FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod files first
COPY go.mod go.sum* ./

# Download dependencies (if any)
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o argocd-monitor ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/argocd-monitor .
COPY web/templates ./web/templates

EXPOSE 8080
CMD ["./argocd-monitor"]