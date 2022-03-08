FROM busybox

WORKDIR /target/usr/local/bin/
COPY . .
RUN chmod -v +x *.sh

###

# yay build

FROM archlinux

WORKDIR /usr/src/ndi-to-srt/
RUN pacman --noconfirm -Sy git sudo make binutils fakeroot base-devel
RUN echo "" && echo "%wheel ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN useradd -UMr -d /usr/src/ndi-to-srt/ -G wheel app
RUN chown -R app .

USER app
RUN git clone --recursive https://aur.archlinux.org/yay.git yay/
RUN cd yay && makepkg --noconfirm -si && cd .. && rm -r yay
RUN yay --noconfirm -S pod2man && sudo rm -r ~/.cache /var/cache/pacman/*
RUN yay --noconfirm -S ndi-advanced-sdk && sudo rm -r ~/.cache /var/cache/pacman/*
RUN yay --noconfirm -S ffmpeg-ndi && sudo rm -r ~/.cache /var/cache/pacman/*

COPY --from=0 /target/ /
CMD ["ndi-to-srt.sh"]

STOPSIGNAL SIGTERM
