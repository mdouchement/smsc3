FROM alpine:3.6
LABEL maintainer="mdouchement"

RUN apk add --no-cache curl ca-certificates bash \
  libxml2 pcre musl hiredis openssl \
  libxml2-dev pcre-dev build-base libtool musl-dev bison \
  mariadb-dev hiredis-dev libressl-dev

RUN curl -Ls https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-alpine-linux-amd64-v0.6.1.tar.gz | tar zxvf - -C /usr/local/bin
# TODO: remove insecure flag and support lets-encrypt (alpine:3.6 have old ca-certificates)
RUN curl -Lks https://www.kannel.org/download/1.4.4/gateway-1.4.4.tar.gz | tar zxvf - -C /tmp

WORKDIR /tmp/gateway-1.4.4

# See: https://redmine.kannel.org/issues/699#note-5
# Kannel 1.4.4 can't compile with current bison version on alpine repositories
ADD patch/bison3.patch .

RUN patch -p1 < bison3.patch \
  && ln -sf /usr/include/poll.h /usr/include/sys/poll.h \
  && ln -s /usr/include/unistd.h /usr/include/sys/unistd.h \
  && ln -s /usr/share/libtool/build-aux/ltmain.sh /tmp/gateway-1.4.4/ltmain.sh \
  && ./configure \
  --with-redis --enable-docs=no --enable-start-stop-daemon=no \
  --without-sdb --without-oracle --without-sqlite2 --with-mysql --with-mysql-dir=/usr/include/mariadb/ \
  && make \
  && make install \
  && cp /tmp/gateway-1.4.4/test/fakesmsc /usr/local/bin/ \
  && apk del libxml2-dev pcre-dev build-base libtool musl-dev bison hiredis-dev libressl-dev \
  && rm -rf /tmp/gateway-1.4.4 && rm -rf /tmp/*

CMD ["/usr/local/sbin/bearerbox", "/etc/kannel/kannel.conf" ]