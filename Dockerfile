FROM golang:1.10.3-alpine3.7

COPY . /go/src/github.com/razorpay/nats-streaming-consumer
WORKDIR /go/src/github.com/razorpay/nats-streaming-consumer
RUN apk add git

RUN go get ; CGO_ENABLED=0 GOOS=linux go build -o nats-streaming-consumer main.go
FROM alpine:3.8
RUN apk update; apk add ca-certificates
COPY --from=0 /go/src/github.com/razorpay/nats-streaming-consumer/nats-streaming-consumer /nats-streaming-consumer
CMD [ "/nats-streaming-consumer" ]