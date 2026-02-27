ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine as base

WORKDIR /usr/src/reddit_tracker

COPY go.mod go.sum ./
RUN go mod download

FROM base as build
COPY src ./
RUN go build -o ./bin/reddit_tracker .

FROM base as final
COPY --from=build /usr/src/reddit_tracker/bin/reddit_tracker /usr/local/bin/reddit_tracker
ENTRYPOINT ["reddit_tracker"]