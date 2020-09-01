
# build stage
FROM golang:alpine as build-env
LABEL maintainer="mdouchement>"

RUN apk upgrade

ENV CGO_ENABLED 0
ENV GO111MODULE on

WORKDIR /smsc3
COPY . .

RUN go mod download
RUN go build -ldflags "-s -w" -o smsc3 .

# final stage
FROM scratch
LABEL maintainer="mdouchement>"

COPY --from=build-env /smsc3/smsc3 /usr/local/bin/

EXPOSE 6000
EXPOSE 20001
CMD ["smsc3"]
