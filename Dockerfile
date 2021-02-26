FROM golang:1.15-alpine3.12 as main-builder
WORKDIR /go/src/github.com/kmgreen2/agglo
RUN apk update && apk add --no-cache make bash
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

FROM public.ecr.aws/lambda/provided:al2 as lambda-builder
# install compiler
WORKDIR /go/src/github.com/kmgreen2/agglo
RUN yum install -y golang make bash
RUN go env -w GOPROXY=direct
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


FROM public.ecr.aws/lambda/provided:al2 as lambda-local
COPY --from=lambda-builder /go/src/github.com/kmgreen2/agglo/bin/binge /binge
ARG CONFIGFILE
RUN bash -c 'if [ -z ${CONFIGFILE} ]; then echo "Environment variable CONFIGFILE must be specified. Exiting."; exit 1; fi'
COPY ${CONFIGFILE} /config.json
RUN cat /config.json
ADD https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie /usr/bin/aws-lambda-rie
RUN chmod 755 /usr/bin/aws-lambda-rie
RUN echo -e "exec /usr/local/bin/aws-lambda-rie /binge -config /config.json -runType lambda" > /entry.sh
RUN cat /entry.sh
RUN chmod 755 /entry.sh
ENTRYPOINT [ "sh", "/entry.sh" ]

FROM public.ecr.aws/lambda/provided:al2 as lambda-production
COPY --from=lambda-builder /go/src/github.com/kmgreen2/agglo/bin/binge /binge
ARG CONFIGFILE
RUN bash -c 'if [ -z ${CONFIGFILE} ]; then echo "Environment variable CONFIGFILE must be specified. Exiting."; exit 1; fi'
COPY ${CONFIGFILE} /config.json
ENTRYPOINT [ "/binge" ]
CMD ["-config", "/config.json", "-runType", "lambda"]

FROM alpine:3.12 as main
WORKDIR /root
COPY --from=builder /go/src/github.com/kmgreen2/agglo/bin/ /usr/local/bin
