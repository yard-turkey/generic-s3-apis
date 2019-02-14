FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/generic-s3-apis /usr/local/bin/generic-s3-apis
