FROM golang:1.8

COPY . /go

CMD ["/go/boot.sh"]