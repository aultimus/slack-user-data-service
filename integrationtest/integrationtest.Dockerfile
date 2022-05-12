FROM golang:1.18
RUN mkdir -p /go/src/github.com/workos-code-challenge/matthew-ault
WORKDIR /go/src/github.com/workos-code-challenge/matthew-ault
ADD . /go/src/github.com/workos-code-challenge/matthew-ault
EXPOSE 8081
RUN go test -c ./integrationtest/...
CMD ["./integrationtest.test", "-test.v", "./integrationtest/...", "-test.count", "1"]
