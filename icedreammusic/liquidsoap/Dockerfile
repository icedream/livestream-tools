ARG IMAGE=savonet/liquidsoap-ci-build:v2.2.2_amd64

FROM $IMAGE

USER root

ENV DEBIAN_FRONTEND=noninteractive
# RUN sed -e 's,^deb\s\+,deb-src ,g' /etc/apt/sources.list > /etc/apt/sources.list.d/sources.list
RUN apt-get update

RUN mkdir -p /packages/
RUN chown _apt /packages/
WORKDIR /packages/

RUN apt-get download \
    -o Dir::Cache="/packages/" \
    -o Dir::Cache::archives="/packages/" \
    $(apt-cache depends --recurse --no-recommends --no-suggests \
    --no-conflicts --no-breaks --no-replaces --no-enhances \
    --no-pre-depends lame | grep "^\w")

###

FROM $IMAGE

COPY --from=0 /packages/ /packages/
USER root
RUN DEBIAN_FRONTEND=noninteractive dpkg -i /packages/*.deb \
    && rm -r /packages/
USER liquidsoap

WORKDIR /liq/
COPY . .
RUN liquidsoap -c metadata_api.liq
RUN liquidsoap -c stream.liq

EXPOSE 8050 8051 9000 9000/udp
STOPSIGNAL SIGTERM
ENTRYPOINT [ "liquidsoap" ]
CMD ["./stream.liq"]
