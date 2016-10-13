#!/bin/sh

mkdir -p ipmitool && cd ipmitool
curl -L -s -o ipmitool.tar.bz2 http://downloads.sourceforge.net/project/ipmitool/ipmitool/1.8.17/ipmitool-1.8.17.tar.bz2
tar -xjf ipmitool.tar.bz2
cd ipmitool-1.8.17
./configure LDFLAGS=-static
make
make install
cd ../..
rm -rf ipmitool
