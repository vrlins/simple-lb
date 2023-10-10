FROM golang:1.21.1 AS builder
WORKDIR /app
COPY go.mod ./
COPY ./cmd ./cmd
COPY ./pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o lb ./cmd/lb

FROM alpine:3.18.4
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/lb .
ENTRYPOINT [ "/root/lb" ]