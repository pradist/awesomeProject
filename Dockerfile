FROM golang:latest as builder

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 go build -o /app main.go

FROM alpine:latest
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
