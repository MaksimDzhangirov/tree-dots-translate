FROM golang:1.16

WORKDIR /go/src/github.com/MaksimDzhangirov/three-dots/code/monolith-microservice-shop
COPY . .

RUN go mod download
RUN go get github.com/cespare/reflex

