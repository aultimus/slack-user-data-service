FROM golang:1.18
RUN mkdir -p /go/src/github.com/aultimus/slack-user-data-service
WORKDIR /go/src/github.com/aultimus/slack-user-data-service
ADD . /go/src/github.com/aultimus/slack-user-data-service
EXPOSE 8081
RUN go test -c ./integrationtest/...
CMD ["./integrationtest.test", "-test.v", "./integrationtest/...", "-test.count", "1"]
