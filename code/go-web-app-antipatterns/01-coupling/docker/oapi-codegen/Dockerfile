FROM golang:1.16
RUN go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]