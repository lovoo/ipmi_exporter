#!/bin/bash

set -e
set -x

PN=ipmitool
PV=1.8.17
P="${PN}-${PV}"

SRC_URI="http://downloads.sourceforge.net/project/${PN}/${PN}/${PV}/${P}.tar.bz2"

S="${PWD}/${P}"

A="${SRC_URI##*/}"

curl -L -o "${A}" "${SRC_URI}"

tar -xjf "${A}"

cd "${S}"
./configure LDFLAGS=-static
make
make install

cd /
rm -rf "${S}"
