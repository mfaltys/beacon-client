FROM alpine

RUN apk add -U ca-certificates

COPY config.gcfg /
COPY beacon-client* /beacon-client
COPY auth /

ENTRYPOINT ["/beacon-client"]
