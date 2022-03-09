FROM busybox

WORKDIR /target/usr/local/bin/
COPY . .
RUN chmod -v +x *.sh

###

FROM golang:1 AS fakesilence

WORKDIR /usr/src/fakesilence
COPY fakesilence/ .
RUN go build -v .
RUN cp fakesilence /usr/local/bin/

###

# yay build

FROM archlinux

WORKDIR /usr/src/ndi-feeder/
RUN pacman --noconfirm -Sy git sudo make binutils fakeroot base-devel
RUN echo "" && echo "%wheel ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN useradd -UMr -d /usr/src/ndi-feeder/ -G wheel app
RUN chown -R app .

USER app
RUN git clone --recursive https://aur.archlinux.org/yay.git yay/
RUN cd yay && makepkg --noconfirm -si && cd .. && rm -r yay
RUN yay --noconfirm -S pod2man && sudo rm -r ~/.cache /var/cache/pacman/*
RUN yay --noconfirm -S ndi-advanced-sdk && sudo rm -r ~/.cache /var/cache/pacman/*
RUN yay --noconfirm -S ffmpeg-ndi && sudo rm -r ~/.cache /var/cache/pacman/*

COPY --from=fakesilence /usr/local/bin/fakesilence /usr/local/bin/

COPY --from=0 /target/ /
CMD ["ndi-feeder.sh"]

STOPSIGNAL SIGTERM
