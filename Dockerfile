# Stage 1: Build Stage
FROM golang:1.22-alpine AS builder

ENV CGO_ENABLED=1 

# Install necessary packages
RUN apk add --no-cache gcc musl-dev git nodejs npm sqlite

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy the source code into the container
COPY . .

# Install Tailwind CSS
RUN npm install -g tailwindcss

# Build the Tailwind CSS
RUN tailwindcss -i ./assets/main.css -o ./assets/output.css --minify

# Generate templ files
RUN templ generate

# Build the Go app
RUN go build -o main ./cmd/main.go

# Stage 2: Run Stage
FROM alpine:latest

# Install sqlite
RUN apk add --no-cache sqlite

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets

# Copy any other necessary files (e.g., templates, static files, etc.)
COPY --from=builder /app/components ./components

COPY --from=builder /app/.env ./.env

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./main"]

