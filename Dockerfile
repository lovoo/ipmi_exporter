FROM alpine:latest

ENV GOPATH /go
ENV APPPATH $GOPATH/src/github.com/lovoo/ipmi_exporter

COPY . $APPPATH
RUN apk add --update -t build-deps go git mercurial libc-dev gcc libgcc make curl && \
    $APPPATH/build_ipmitool.sh && \
    cd $APPPATH && make build && mv build/ipmi_exporter / && \
    apk del --purge build-deps && \
    rm -rf $GOPATH

EXPOSE 9289

ENTRYPOINT [ "/ipmi_exporter" ]
