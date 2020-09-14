FROM golang:1.14 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -mod vendor -o /auto-secret-replication

FROM alpine:3.12
COPY --from=builder /auto-secret-replication /auto-secret-replication
CMD ["/auto-secret-replication"]