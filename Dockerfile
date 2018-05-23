# Build hypereth in a stock Go builder container
FROM golang:1-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . $GOPATH/src/github.com/getamis/hypereth
RUN cd $GOPATH/src/github.com/getamis/hypereth && make hypereth && mv build/bin/hypereth /hypereth

# Pull hypereth into a second stage deploy alpine container
FROM alpine:3.7

RUN apk add --no-cache ca-certificates
COPY --from=builder /hypereth /usr/local/bin/

ENTRYPOINT [ "hypereth" ]
