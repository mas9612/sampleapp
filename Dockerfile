FROM golang:1.11

RUN mkdir -p /go/src/sampleapp
WORKDIR /go/src/sampleapp

COPY . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN CGO_ENABLED=0 go build .


FROM alpine:latest

RUN mkdir /app
WORKDIR /app

COPY --from=0 /go/src/sampleapp/sampleapp .
ENTRYPOINT ["./sampleapp"]
