# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Wiard van Rij <wiard@outlook.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod, sum and main files
COPY go.mod go.sum main.go ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

######## Start a new stage from scratch #######
FROM scratch

ENV COUNTER_RECHECKAMOUNT=20
ENV COUNTER_HEALTHCHECKTIME=500
ENV COUNTER_REDISHOST="redis:6379"

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"] 