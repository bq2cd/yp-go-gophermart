# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o /bin/gophermart ./cmd/gophermart/

#################################################################################

FROM alpine:3.23 AS runtime

COPY --from=builder /bin/gophermart /usr/local/bin/gophermart

EXPOSE 8080
ENV RUN_ADDRESS=":8080"
ENTRYPOINT ["gophermart"]
