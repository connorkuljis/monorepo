FROM golang:1.23

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app with optimizations for size
RUN go build -v -o /usr/bin/server ./cmd/server/main.go

CMD ["server"]

