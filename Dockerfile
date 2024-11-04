FROM golang:1.22 AS builder
WORKDIR /tmp/msapi
COPY . ./

RUN ls && env GOOS=linux GOARCH=amd64 go build -o msapi .

FROM alpine:latest AS product

WORKDIR /data/msapi

COPY --from=builder /tmp/msapi/msapi ./

RUN ls && chmod a+x /data/msapi/msapi

EXPOSE 3000

CMD ["/data/msapi/msapi"]


