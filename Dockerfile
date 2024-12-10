FROM golang:1.21 as builder

WORKDIR /tmp/msapi

COPY . .
RUN ls
RUN go mod download && \
    CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o msapi .

FROM node:20.14.0 as producer

WORKDIR /data/msapi

COPY ./neteasecloudmusicapi ./
RUN ls \
    && cd neteasecloudmusicapi \
    && npm install \

COPY --from=builder /tmp/msapi/msapi ./
RUN chmod a+x msapi

CMD ["./msapi"]
