FROM golang:1.20-alpine3.18 as build

ADD . /build
WORKDIR /build

RUN echo go version: `go version`
RUN CGO_ENABLED=0 go build -o webhooksvc -ldflags "-s -w" ./

FROM alpine:3.18

ENV \
    APP_USER=app \
    APP_UID=1001
RUN \
    apk add --no-cache --update curl ca-certificates && \
    mkdir -p /home/$APP_USER && \
    adduser -s /bin/sh -D -u $APP_UID $APP_USER && chown -R $APP_USER:$APP_USER /home/$APP_USER && \
    mkdir -p /srv && chown -R $APP_USER:$APP_USER /srv && \
    rm -rf /var/cache/apk/*

USER app
WORKDIR /srv
COPY --from=build --chown=$APP_USER:$APP_USER /build/webhooksvc /srv/webhooksvc

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1

ENTRYPOINT ["/srv/webhooksvc"]
