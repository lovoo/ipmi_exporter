FROM alpine:latest

RUN apk -U add curl file gcc libgcc libc-dev make automake autoconf libtool
COPY build_ipmitool.sh .
RUN bash build_ipmitool.sh

ENV GOPATH /go
ENV APPPATH $GOPATH/src/github.com/lovoo/ipmi_exporter

COPY . $APPPATH

RUN apk -U add --update -t build-deps go git mercurial

RUN cd $APPPATH && go get -d && go build -o /ipmi_exporter \
    && apk del --purge build-deps git mercurial curl file gcc libgcc libc-dev make automake autoconf libtool && rm -rf $GOPATH

EXPOSE 9289

ENTRYPOINT ["/ipmi_exporter"]
