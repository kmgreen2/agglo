FROM golang:1.15-alpine3.12 as builder
WORKDIR /go/src/github.com/kmgreen2/agglo
RUN apk update && apk add --no-cache make
COPY pkg ./pkg
COPY internal ./internal
COPY generated ./generated
COPY old ./old
COPY test ./test
COPY cmd ./cmd
COPY Makefile .
COPY go.mod .
RUN GOSUMDB=off go mod tidy
RUN make build

FROM alpine:3.12
WORKDIR /root
COPY --from=builder /go/src/github.com/kmgreen2/agglo/bin/ /usr/local/bin
