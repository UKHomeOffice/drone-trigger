FROM alpine:3.4

RUN apk upgrade --no-cache
RUN apk add --no-cache ca-certificates openssl && wget -q https://github.com/UKHomeOffice/drone-trigger/releases/download/v0.0.4/drone-trigger_linux_amd64 -O /bin/drone-trigger && chmod +x /bin/drone-trigger
ENTRYPOINT ["/bin/drone-trigger"]
