# ARG IMAGE=savonet/liquidsoap:master
ARG IMAGE=savonet/liquidsoap-ci-build:v2.1.4_alpine_amd64

# FROM $IMAGE

# USER root

# ENV DEBIAN_FRONTEND=noninteractive
# RUN sed -e 's,^deb\s\+,deb-src ,g' /etc/apt/sources.list > /etc/apt/sources.list.d/sources.list
# RUN apt-get update
# RUN sed -i -e '/crontab/d' -e '/Debian-exim/d' /var/lib/dpkg/statoverride
# RUN apt-get build-dep -y ffmpeg
# RUN apt-get install -y libfdk-aac-dev devscripts
# RUN apt-get source -y ffmpeg
# RUN sed -i -e 's,--enable-gnutls,--enable-gnutls \\\n\t--enable-libfdk-aac --enable-nonfree,g' ffmpeg-*/debian/rules
# RUN cd ffmpeg-*; dch --local '-icedream' "Build with --enable-libfdk-aac and --enable-nonfree." && cat debian/changelog
# RUN cd ffmpeg-*; dpkg-buildpackage -b -tc -j$(nproc)
# RUN mkdir -p /packages/ && mv *_*.deb /packages/

###

FROM $IMAGE

# COPY --from=0 /packages/ /packages/
# USER root
# RUN DEBIAN_FRONTEND=noninteractive apt install -y \
#     /packages/ffmpeg_*.deb \
#     /packages/libavcodec-dev*_*.deb \
#     /packages/libavcodec-extra*_*.deb \
#     /packages/libavdevice*_*.deb \
#     /packages/libavfilter-dev*_*.deb \
#     /packages/libavfilter-extra*_*.deb \
#     /packages/libavformat*_*.deb \
#     /packages/libavresample*_*.deb \
#     /packages/libavutil*_*.deb \
#     /packages/libpostproc*_*.deb \
#     /packages/libswresample*_*.deb \
#     /packages/libswscale*_*.deb \
#     && rm -r /packages/
# USER liquidsoap

WORKDIR /liq/
COPY . .
RUN liquidsoap -c stream.liq

EXPOSE 8050 8051 9000 9000/udp
STOPSIGNAL SIGTERM
ENTRYPOINT [ "liquidsoap" ]
CMD ["./stream.liq"]
