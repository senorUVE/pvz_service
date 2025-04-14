FROM golang:1.23.0-alpine3.20 as builder

COPY ../go.mod .
COPY ../go.sum .
RUN go mod download

COPY .. .

RUN go build -o /app/bin/cmd .cmd/main.go

FROM debian:bookworm-slim

WORKDIR /root/

COPY --from=builder /app/bin/cmd ./cmd

EXPOSE 8080

CMD ["./cmd"]