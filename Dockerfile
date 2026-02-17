FROM golang:1.25.1-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./   

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build ./cmd/main/main.go 

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build .

EXPOSE 8080

CMD ["./build"]