FROM golang:1.14.12-stretch AS builder
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod tidy && go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags netgo -o bin/main cmd/server/main.go

FROM debian:stretch AS bin
WORKDIR /app
COPY --from=builder /src/bin/main .
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
CMD ["./main"]