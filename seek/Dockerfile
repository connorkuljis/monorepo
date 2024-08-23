FROM golang:1.23-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /bin/server ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/server /bin/server

CMD ["server"]

