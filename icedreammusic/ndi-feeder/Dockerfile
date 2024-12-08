# syntax=docker/dockerfile:1.12.0

FROM alpine AS rootfs

RUN apk add --no-cache gnupg

RUN wget -O- https://raw.githubusercontent.com/archlinuxarm/archlinuxarm-keyring/master/archlinuxarm.gpg | gpg --import

WORKDIR /target/
ARG ALARM_ROOTFS_URL=http://os.archlinuxarm.org/os/ArchLinuxARM-rpi-2-latest.tar.gz
RUN wget "${ALARM_ROOTFS_URL}" -O/tmp/rootfs.tar.gz
RUN wget "${ALARM_ROOTFS_URL}.sig" -O/tmp/rootfs.tar.gz.sig
RUN gpg --verify /tmp/rootfs.tar.gz.sig
RUN tar -xvpzf /tmp/rootfs.tar.gz

###
# PREPARE LAYER FOR UPDATES AND GENERAL PACKAGE INSTALLATION

# FROM scratch AS image-base

# COPY --from=rootfs /target/ /

FROM archlinux AS image-base

# Make powerpill not act up later, placing this early for validation consistency
RUN sed -i 's,SigLevel\s\+=\s\+Required,SigLevel = PackageRequired,' /etc/pacman.conf

RUN pacman -Sy --noconfirm
RUN pacman-key --init

# Install core keyring (https://archlinuxarm.org/about/package-signing)
# RUN pacman -S --needed --noconfirm archlinuxarm-keyring
# RUN pacman-key --populate archlinuxarm
RUN pacman-key --populate archlinux
RUN pacman -S --needed --noconfirm archlinux-keyring
RUN pacman-key --populate archlinux

# ###
# # INSTALL FILESYSTEM PACKAGE UPDATES
# # We have to do this with an alternative root since /etc/{hosts,resolv.conf}
# # are mounted read-only by Docker.

# FROM image-base AS updated-filesystem-base

# COPY --from=image-base / /target/

FROM image-base AS updated-filesystem-base

# RUN \
# 	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
# 	--mount=type=cache,target=/tmp/build/.cache \
# 	pacman -r /target/ -S --noconfirm --needed filesystem

# ###
# # LAYER USED FOR INSTALLING UPDATES AND ADDITIONAL PACKAGES USED IN FINAL IMAGE

# FROM scratch AS base

# COPY --from=updated-filesystem-base /target/ /

FROM updated-filesystem-base AS base

# # Install updates
# # NOTE - we install fsck helpers for fat and ext4 in this stage to save on time spent on /boot updates
# RUN \
# 	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
# 	--mount=type=cache,target=/tmp/build/.cache \
# 	pacman -Suu --noconfirm --needed dosfstools e2fsprogs

###
# LAYER USED TO COMPILE STUFF

FROM image-base AS base-devel

RUN pacman -S --noconfirm base-devel git

#RUN pacman -S --noconfirm --needed sudo
RUN (echo "" && echo "%wheel ALL=(ALL) NOPASSWD: ALL") >> /etc/sudoers

RUN useradd -r -N -m -G wheel -d /tmp/build -k /var/empty build

RUN sed -i \
	-e 's,#MAKEFLAGS=.*,MAKEFLAGS="-j$(getconf _NPROCESSORS_ONLN)",g' \
	/etc/makepkg.conf

RUN \
	--mount=type=cache,target=/tmp/build/.cache \
	chown -Rv build /tmp/build /tmp/build/.cache

RUN echo "ParallelDownloads = 5" >>/etc/pacman.conf

USER build

# Needed for anything commits
RUN git config --global user.email "$(whoami)@localhost"
RUN git config --global user.name "Build"

###
# FAKESILENCE

FROM golang:1 AS fakesilence

WORKDIR /usr/src/fakesilence
ARG FAKESILENCE_VERSION=5d8d79af09df5c195908e004a4a3971942180901
RUN go install -v -ldflags "-s -w" github.com/icedream/fakesilence@"${FAKESILENCE_VERSION}"
RUN cp -v "$GOPATH"/bin/* /usr/local/bin

###
# YAY

FROM base-devel AS yay

# NOTE - for why we do `ulimit -n 1024` see https://github.com/moby/moby/issues/27195#issuecomment-1410745778 

WORKDIR /usr/src/yay
RUN git clone --recursive https://aur.archlinux.org/yay.git .
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	ulimit -n 1024 && makepkg -sr --noconfirm --nocheck

###
# BASE DEVEL (YAY)

FROM base-devel AS base-devel-yay

USER root
COPY --from=yay /usr/src/yay/*.pkg.* /tmp/
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	pacman --noconfirm -U /tmp/*.pkg.* && rm /tmp/*.pkg.*

USER build

###
# POD2MAN

FROM base-devel-yay AS pod2man

WORKDIR /usr/src/pod2man

RUN git clone --recursive https://aur.archlinux.org/pod2man.git .

RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${depends[@]}") && (mv -v ~/.cache/yay/*/*.pkg.* . || true))
RUN ulimit -n 1024 && makepkg -sr --noconfirm

###
# NDI-SDK-EMBEDDED

FROM base-devel-yay AS ndi-sdk-embedded

WORKDIR /usr/src/ndi-sdk-embedded

RUN git clone --recursive https://aur.archlinux.org/ndi-sdk-embedded.git .

RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${depends[@]}") && (mv -v ~/.cache/yay/*/*.pkg.* . || true))
RUN ulimit -n 1024 && makepkg -sr --noconfirm

###
# NDI-SDK

FROM base-devel-yay AS ndi-sdk

WORKDIR /usr/src/ndi-sdk
RUN git clone --recursive https://aur.archlinux.org/ndi-sdk.git .

RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${depends[@]}") && (mv -v ~/.cache/yay/*/*.pkg.* . || true))
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${makedepends[@]}"))
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	ulimit -n 1024 && makepkg -sr --noconfirm

###
# NDI-ADVANCED-SDK

FROM base-devel-yay AS ndi-advanced-sdk

WORKDIR /usr/src/ndi-advanced-sdk
RUN git clone --recursive https://aur.archlinux.org/ndi-advanced-sdk.git .

RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${depends[@]}") && (mv -v ~/.cache/yay/*/*.pkg.* . || true))
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${makedepends[@]}"))
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	ulimit -n 1024 && makepkg -sr --noconfirm

###
# FFMPEG-NDI

FROM base-devel-yay AS ffmpeg-ndi

WORKDIR /usr/src/ffmpeg-ndi

USER root
# COPY --from=ndi-sdk-embedded /usr/src/ndi-sdk-embedded/*.pkg.* /tmp/
COPY --from=ndi-sdk /usr/src/ndi-sdk/*.pkg.* /tmp/
COPY --from=pod2man /usr/src/pod2man/*.pkg.* /tmp/
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	yay --noconfirm -U /tmp/*.pkg.* && rm /tmp/*.pkg.*

USER build
# ffmpeg-ndi AUR package is no longer maintained, use our own modified copy
COPY ./packages/ffmpeg-ndi/ .
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(\
	. ./PKGBUILD &&\
	if [ "${#depends[@]}" -eq 0 ]; then exit; fi &&\
	packages=$(yay -T "${depends[@]}" 2>/dev/null|| true) &&\
	if [ -z "$packages" ]; then exit; fi &&\
	yay -S --noconfirm --asdeps --provides --needed $packages &&\
	find ~/.cache/yay/ -mindepth 2 -maxdepth 2 -name \*.pkg.\* -exec mv {} . \;\
	)
# RUN (. ./PKGBUILD && yay -S --noconfirm --asdeps --provides --needed $(yay -T "${optdepends[@]}") && (mv -v ~/.cache/yay/*/*.pkg.* . || true))
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	(\
	. ./PKGBUILD &&\
	if [ "${#makedepends[@]}" -eq 0 ]; then exit; fi &&\
	packages=$(yay -T "${makedepends[@]}" 2>/dev/null|| true) &&\
	if [ -z "$packages" ]; then exit; fi &&\
	yay -S --noconfirm --asdeps --provides --needed $packages \
	)
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	ulimit -n 1024 && makepkg -sr --noconfirm --nocheck

###
# PERMISSIONS FOR FINAL IMAGE FILES

FROM busybox AS files

WORKDIR /target/usr/local/bin/
COPY *.sh .
RUN dos2unix *.sh
RUN chmod -v +x *.sh

###
# PACKAGES

FROM scratch as packages

COPY --from=ndi-sdk /usr/src/ndi-sdk/*.pkg.* /packages/
COPY --from=ffmpeg-ndi /usr/src/ffmpeg-ndi/*.pkg.* /packages/
COPY --from=fakesilence /usr/local/bin/fakesilence /target/usr/local/bin/

###
# FINAL IMAGE

FROM base AS final-image

USER root
# COPY --from=powerpill /usr/src/powerpill/*.pkg.* /tmp/
# RUN \
# 	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
# 	--mount=type=cache,target=/tmp/build/.cache \
# 	pacman --noconfirm -U /tmp/*.pkg.*; rm /tmp/*.pkg.*

#COPY --from=yay /usr/src/yay/*.pkg.* /tmp/
COPY --from=ndi-sdk /usr/src/ndi-sdk/*.pkg.* /tmp/
COPY --from=ffmpeg-ndi /usr/src/ffmpeg-ndi/*.pkg.* /tmp/
RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	rm -f /var/cache/pacman/pkg/cache.lck; pacman --noconfirm -U /tmp/*.pkg.*; rm /tmp/*.pkg.*

RUN \
	--mount=type=cache,target=/var/cache/pacman/pkg,sharing=locked \
	--mount=type=cache,target=/tmp/build/.cache \
	rm -f /var/cache/pacman/pkg/cache.lck; pacman -S --noconfirm --needed sudo realtime-privileges

COPY --from=fakesilence /usr/local/bin/fakesilence /usr/local/bin/

COPY --from=files /target/ /

RUN useradd -m -u 1000 -G wheel,realtime,audio,video ndi-feeder
RUN echo "ndi-feeder:ndi-feeder" | chpasswd

USER ndi-feeder
CMD ["ndi-feeder.sh"]
STOPSIGNAL SIGTERM
