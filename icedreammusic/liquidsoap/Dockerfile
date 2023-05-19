ARG IMAGE=savonet/liquidsoap-ci-build:rolling-release-v2.2.x@sha256:4cee43e436f6dff60e77273a5608686991f0703d2763fae1c6131e7fdb0886ad

FROM $IMAGE

WORKDIR /liq/
COPY . .
RUN liquidsoap -c metadata_api.liq
RUN liquidsoap -c stream.liq

EXPOSE 8050 8051 9000 9000/udp
STOPSIGNAL SIGTERM
ENTRYPOINT [ "liquidsoap" ]
CMD ["./stream.liq"]
