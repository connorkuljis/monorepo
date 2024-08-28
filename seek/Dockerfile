FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/server/main.go

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/server ./
COPY templates /app/templates
CMD ["./server"]

