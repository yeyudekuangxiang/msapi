FROM golang:1.22 as builder
WORKDIR /tmp
COPY . ./

RUN ls && env GOOS=linux GOARCH=amd64 go build -o msapi .

FROM alpine:latest

WORKDIR /data/msapi

COPY --from=builder /tmp/msapi ./

RUN ls && chmod a+x ./msapi

EXPOSE 8080

CMD ["./msapi"]


