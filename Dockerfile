FROM golang:alpine as builder

RUN go version

COPY . /contract_feed/
WORKDIR /contract_feed/

RUN apk update && apk add --no-cache tzdata build-base

RUN go mod download
RUN GOOS=linux go build -o ./.bin/bot ./cmd/bot

FROM alpine:latest

WORKDIR /root/

RUN apk update && apk add --no-cache tzdata

COPY --from=builder /contract_feed/.bin/bot .

CMD ["./bot"]
