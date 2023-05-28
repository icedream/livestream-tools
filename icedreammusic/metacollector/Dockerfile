FROM golang:1.20-alpine

WORKDIR /usr/src/icedreammusic/
COPY . .

RUN apk add alpine-sdk sqlite-dev

RUN go build -v ./cmd/metacollectord/
RUN install -v -m0755 -d /target/usr/local/bin/
RUN install -v -m0755 metacollectord /target/usr/local/bin/metacollectord

###

FROM alpine:3.18

RUN apk add sqlite-libs

COPY --from=0 /target/ /

WORKDIR /library

VOLUME ["/library"]
#RUN addgroup -S -g 950 app
#RUN adduser -S -k /dev/empty -g "App user" -h /library -u 950 -G app app
#USER 950
CMD ["metacollectord"]
