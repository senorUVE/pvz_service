FROM golang:1.23.0-alpine3.20 as builder

WORKDIR /app

COPY ../go.mod .
COPY ../go.sum .
RUN go mod download

COPY .. .

RUN go build -o ./bin/pvz_service cmd/main.go

FROM debian:bookworm-slim

WORKDIR /root/

COPY --from=builder /app/bin/pvz_service .

EXPOSE 8080

CMD ["./pvz_service"]