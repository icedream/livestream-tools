ARG IMAGE=savonet/liquidsoap-ci-build:v2.2.2_amd64

FROM $IMAGE

WORKDIR /liq/
COPY . .
RUN liquidsoap -c metadata_api.liq
RUN liquidsoap -c stream.liq

EXPOSE 8050 8051 9000 9000/udp
STOPSIGNAL SIGTERM
ENTRYPOINT [ "liquidsoap" ]
CMD ["./stream.liq"]
