FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN doc=0 GOOS=linux go build -o /go-drive ./cmd/app/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /go-drive /app/go-drive
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["/app/go-drive"]