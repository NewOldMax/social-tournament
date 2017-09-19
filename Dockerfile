FROM golang:1.8

COPY . /go

RUN go get ./...

CMD ["/go/boot.sh"]