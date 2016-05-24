FROM alpine:latest

ENV GOPATH /go
ENV APPPATH $GOPATH/src/github.com/lovoo/ipmi_exporter

COPY . $APPPATH

RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN apk -U add --update ipmitool@testing
RUN apk -U add --update -t build-deps go git mercurial

RUN cd $APPPATH && go get -d && go build -o /ipmi_exporter \
    && apk del --purge build-deps git mercurial && rm -rf $GOPATH

EXPOSE 9189

ENTRYPOINT ["/ipmi_exporter"]
