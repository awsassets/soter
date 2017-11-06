# Build
FROM golang:alpine AS build

ARG TAG
ARG BUILD

ENV APP soter
ENV REPO prologic/$APP

RUN apk add --update git make build-base && \
    rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/$REPO
COPY . /go/src/github.com/$REPO
RUN make TAG=$TAG BUILD=$BUILD build

# Runtime
FROM alpine

ENV APP soter
ENV REPO prologic/$APP

LABEL soter.app main

COPY --from=build /go/src/github.com/${REPO}/${APP} /${APP}

EXPOSE 6667/tcp 6697/tcp

ENTRYPOINT ["/soter"]
CMD [""]
