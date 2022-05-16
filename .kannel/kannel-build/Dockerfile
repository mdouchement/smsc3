FROM alpine:3.12 AS base
LABEL maintainer="mdouchement"

ARG VERSION=1.4.5
ARG BUILD_DEPS="\
  ca-certificates bash curl \
  libxml2 pcre musl openssl \
  libxml2-dev pcre-dev build-base libtool musl-dev bison \
  hiredis-dev openssl-dev mariadb-dev automake"
ARG RUNTIME_DEPS="libxml2 pcre musl openssl ca-certificates hiredis-dev"

LABEL version=${VERSION} \
  os="Alpine" \
  os.version="3" \
  name="Kannel ${VERSION}" \
  description="Kannel image" \
  maintainer="mdouchement"

ENV VERSION=${VERSION} \
  BUILD_DEPS=${BUILD_DEPS} \
  RUNTIME_DEPS=${RUNTIME_DEPS}

#
#
FROM base AS build

RUN apk add --no-cache ${BUILD_DEPS}

RUN curl -Ls https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-alpine-linux-amd64-v0.6.1.tar.gz | tar zxvf - -C /usr/local/bin
RUN curl -Lks https://www.kannel.org/download/${VERSION}/gateway-${VERSION}.tar.gz | tar zxvf - -C /tmp

WORKDIR /tmp/gateway-${VERSION}

ADD patch/gateway-${VERSION}.patch .

RUN patch -p1 < gateway-${VERSION}.patch \
  && ln -sf /usr/include/poll.h /usr/include/sys/poll.h \
  && ln -s /usr/include/unistd.h /usr/include/sys/unistd.h \
  && ./configure \
  --with-redis --enable-docs=no --enable-start-stop-daemon=no --disable-assertions \
  --disable-wap \
  --without-sdb --without-oracle --without-sqlite2 --without-cassandra \
  --without-mysql \
  # --with-mysql --with-mysql-dir=/usr/include/mariadb/ \
  && touch .depend \
  && make \
  && make install \
  && cp test/fakesmsc /usr/local/sbin/

#
#
FROM base

COPY --from=build /usr/local/bin/dockerize       /usr/local/sbin/dockerize
COPY --from=build /usr/local/sbin/bearerbox      /usr/local/sbin/bearerbox
COPY --from=build /usr/local/sbin/smsbox         /usr/local/sbin/smsbox
COPY --from=build /usr/local/sbin/run_kannel_box /usr/local/sbin/run_kannel_box
COPY --from=build /usr/local/sbin/wapbox         /usr/local/sbin/wapbox
COPY --from=build /usr/local/sbin/fakesmsc       /usr/local/sbin/fakesmsc

RUN apk add --no-cache ${RUNTIME_DEPS}

CMD ["/usr/local/sbin/bearerbox", "/etc/kannel/kannel.conf" ]