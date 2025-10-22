FROM golang:1.25.2-alpine3.22 AS build
RUN apk add -U make binutils
WORKDIR /build
RUN --mount=type=bind,source=.,destination=/build,rw \
    make binary-build && \
    mv backend /usr/bin/backend

FROM alpine:3.22.1
RUN apk add -U make postgresql
RUN go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
WORKDIR /application
COPY --from=build /usr/bin/backend /usr/bin/backend
COPY migration migration
COPY template template
