FROM golang:alpine
LABEL maintainer="mdouchement"

WORKDIR /root
ADD . .
RUN go mod init httpcat
RUN go get
RUN go build -o httpcat .

CMD ./httpcat