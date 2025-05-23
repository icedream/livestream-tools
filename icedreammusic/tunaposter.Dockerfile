FROM golang:1.24-alpine

WORKDIR /usr/src/icedreammusic/
COPY . .

RUN cd tunaposter && go build -v .
RUN install -v -m0755 -d /target/usr/local/bin/
RUN install -v -m0755 tunaposter/tunaposter /target/usr/local/bin/tunaposter

###

FROM alpine:3.21

COPY --from=0 /target/ /

CMD ["tunaposter"]
