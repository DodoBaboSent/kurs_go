# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory in the container to /go/src/app
WORKDIR /go/src/app

# Copy the local package files to the container's workspace
COPY . .

# Expose port 3000 for Node.js application
EXPOSE 8080


RUN go get
RUN go build main.go

# Run npm run dev when the container launches
CMD ["./main"]
