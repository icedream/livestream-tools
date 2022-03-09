FROM alpine:3.15 AS share

RUN apk add --no-cache git
WORKDIR /icecast/share/
#COPY share/ .
RUN git clone --depth=1 --recursive https://github.com/logue/icecast2-bootstrap-theme . && rm -rf .git
RUN chown 9999:0 .
RUN chmod -R a-rwx,a+rX .

###

FROM alpine:3.15 AS icecast-download

RUN apk add --no-cache curl ca-certificates
WORKDIR /usr/src/
ARG ICECAST_VERSION=2.4.4
RUN curl -L http://downloads.xiph.org/releases/icecast/icecast-${ICECAST_VERSION}.tar.gz | tar xz -v

###

FROM alpine:3.15 AS icecast

RUN apk add --no-cache \
	build-base file openssl-dev libxslt-dev \
	libvorbis-dev opus-dev libogg-dev speex-dev \
	libtheora-dev curl-dev

ARG ICECAST_VERSION=2.4.4
WORKDIR /usr/src/icecast-${ICECAST_VERSION}
COPY --from=icecast-download /usr/src/icecast-${ICECAST_VERSION}/ .
RUN ./configure
RUN make
RUN make install

###

FROM alpine:3.15

# add runtime deps
RUN \
	apk add --no-cache file libssl1.1 libxslt libvorbis \
		opus libogg speex libtheora \
		libtheora curl && \
	rm -rf /tmp/* /var/cache/apk/*

# add icecast user
RUN \
	addgroup -g 950 icecast &&\
	adduser -S -D -H -u 9999 -G icecast -s /bin/false icecast

# add mime.types file
RUN apk add --no-cache mailcap && cp /etc/mime.types /etc/mime.types.keep && apk del --no-cache mailcap && mv /etc/mime.types.keep /etc/mime.types

# add dumb-init
ARG DUMB_INIT_VERSION=1.2.5
ADD https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_x86_64 /usr/local/bin/dumb-init
RUN chmod +x /usr/local/bin/dumb-init

# install icecast bins
COPY --from=icecast /usr/local/ /usr/local/

# install share files
COPY --from=share /icecast/share/ /icecast/share/

USER 9999
#VOLUME [ "/data" ]
ENTRYPOINT [ "dumb-init" ]
CMD [ "icecast", "-c", "/icecast.xml" ]
EXPOSE 8000
