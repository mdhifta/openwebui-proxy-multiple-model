FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/proxy-gateway

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/proxy-gateway /app/proxy-gateway
COPY cerdential-gcp.json /app/cerdential-gcp.json


ENTRYPOINT ["/app/proxy-gateway"]
