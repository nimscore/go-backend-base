FROM golang:1.25.2-alpine3.22 AS build
RUN apk add -U make binutils
WORKDIR /build
RUN --mount=type=bind,source=.,destination=/build,rw \
    make binary-build && \
    mv mail /usr/bin/mail

FROM alpine:3.22.1
WORKDIR /application
COPY --from=build /usr/bin/mail /application/mail
COPY template /application/template
ENTRYPOINT [ "/application/mail", "worker" ]
