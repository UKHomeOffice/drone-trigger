FROM alpine:3.4

RUN apk upgrade --no-cache
RUN apk add --no-cache ca-certificates

COPY bin/drone-trigger_linux_amd64 /bin/drone-trigger

ENTRYPOINT ["/bin/drone-trigger"]
