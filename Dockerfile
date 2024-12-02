# Use Golang base image (you can keep it alpine to minimize size)
FROM golang:alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Install dependencies (if you have a go.mod, go.sum, etc.)
RUN go mod tidy

# Build the Go application (adjust the build command as necessary)
RUN go build -o server ./cmd/main.go

# Expose the port your Go app will listen on
EXPOSE 8080

# Start the Go server when the container runs
CMD ["./server"]
