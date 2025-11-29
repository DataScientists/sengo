# Use the official Golang image as the base image
FROM golang:1.25.4

# Install required files
RUN go install github.com/air-verse/air@latest
RUN go install gotest.tools/gotestsum@latest

ENV PATH="${PATH}:/go/bin"
# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Expose the port the app runs on
EXPOSE 8080

# Command to run Air for hot-reloading
CMD ["air", "-c", ".air.toml"]
