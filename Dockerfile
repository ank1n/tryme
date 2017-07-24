FROM golang:latest

RUN mkdir -p /go/src/wapp

WORKDIR /go/src/wapp

COPY . /go/src/wapp

RUN go-wrapper download

RUN go-wrapper install

CMD ["go-wrapper", "run", ""]

EXPOSE 8080