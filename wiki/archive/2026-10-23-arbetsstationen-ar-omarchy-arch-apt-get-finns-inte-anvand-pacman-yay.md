---
title: Arbetsstationen ar Omarchy/Arch - apt-get finns inte, anvand pacman/yay
type: note
tags: [verktyg, omarchy, bygg, projekt:infrastructure]
---

Arbetstationen (datorn jag sitter pa) ar Omarchy, dvs Arch-baserat. Det finns INTE apt-get, bara pacman och yay. U-Boot-dokumentationen och bygginstruktioner pa webben (t.ex. docs.u-boot.org) anger Debian/Ubuntu-paketnamn - de maste oversattas.

Mapping som anvands: bc->bc, bison->bison, build-essential->base-devel, coccinelle->coccinelle-bin (AUR, fardigbyggd, ger /usr/bin/spatch), device-tree-compiler->dtc, dfu-util->dfu-util, efitools->efitools, flex->flex, gdisk->gptfdisk (gparted/sgdisk), graphviz->graphviz, imagemagick->imagemagick, liblz4-tool->lz4, libgnutls28-dev->gnutls, libguestfs-tools->guestfs-tools+libguestfs, libncurses-dev->ncurses, libpython3-dev->python, libsdl2-dev->sdl2, libssl-dev->openssl, lzma/lzma-alone->xz, pkg-config->pkgconf, python3-*->python-* (asteval, coverage, filelock, pycryptodome, pyelftools, pytest, pytest-xdist, sphinxcontrib-apidoc, sphinx_rtd_theme, subunit, testtools, virtualenv), uuid-dev->util-linux (uuidgen), gcc-arm-none-eabi->arm-none-eabi-gcc + arm-none-eabi-binutils.

Viktigt: namnen har UNDERSTRECK i Arch for python-sphinx_rtd_theme (inte bindestreck) - bindestreck-versionen finns inte i AUR. Paketet spatch finns inte i AUR, det heter coccinelle/coccinelle-bin.

Redan installerat for U-Boot/TF-A-byggen: base-devel, bison, flex, dtc, swig, python-pyelftools, aarch64-linux-gnu-gcc/binutils 16.2.0, arm-none-eabi-gcc/binutils 16.2.0.
