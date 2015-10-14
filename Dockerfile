FROM golang:1.5

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ENV GO15VENDOREXPERIMENT=1
CMD ["go-wrapper", "run"]

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
