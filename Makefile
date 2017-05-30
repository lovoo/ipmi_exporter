BUILD_DIR    = $(CURDIR)/build
PROJECT_NAME = ipmi_exporter
VERSION      = $(shell git describe --tags || echo 0.0.0-dev)
GO           = go
GOX          = gox
MEGACHECK    ?= $(GOPATH)/bin/megacheck
GOX_ARGS     = -output="$(BUILD_DIR)/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="linux/amd64 linux/386 linux/arm linux/arm64 darwin/amd64 freebsd/amd64 freebsd/386 windows/386 windows/amd64"
pkgs         = $(shell $(GO) list ./... | grep -v /vendor/)

all: format vet megacheck build test

build:
	@echo ">> building binaries"
	GOBIN=$(BUILD_DIR) $(GO) install -v -ldflags '-X main.Version=$(VERSION)'

clean:
	rm -R $(BUILD_DIR)/* || true

test:
	@echo ">> running tests"
	@$(GO) test $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

megacheck:
	@echo ">> megacheck code"
	@$(GO) get -u honnef.co/go/tools/cmd/megacheck
	@$(MEGACHECK) $(pkgs)

release-build:
	@$(GO) get -u github.com/mitchellh/gox
	@$(GOX) $(GOX_ARGS) github.com/lovoo/$(PROJECT_NAME)

deb:
	make build-deb ARCH=amd64 GOARCH=amd64
	make build-deb ARCH=i386 GOARCH=386
	make build-deb ARCH=arm64 GOARCH=arm64
	make build-deb ARCH=armhf GOARCH=arm

build-deb:
	fpm -s dir -t deb \
		--name $(PROJECT_NAME) \
		--version $(VERSION) \
		--package $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(ARCH).deb \
		--depends ipmitool \
		--maintainer "LOVOO IT Operations <it.operations@lovoo.com>" \
		--deb-priority optional \
		--category monitoring \
		--force \
		--deb-compression bzip2 \
		--license "BSD-3-Clause" \
		--vendor "LOVOO GmbH" \
		--deb-no-default-config-files \
		--after-install packaging/postinst.deb \
		--before-remove packaging/prerm.deb \
		--url https://github.com/lovoo/ipmi_exporter \
		--description "Exports statistics from IPMI and publishes them for scraping by Prometheus." \
		--architecture $(ARCH) \
		$(BUILD_DIR)/$(PROJECT_NAME)_linux_$(GOARCH)=/usr/bin/ipmi_exporter \
		packaging/ipmi-exporter.service=/lib/systemd/system/ipmi-exporter.service

release-package:
	package_cloud push lovooOS/prometheus-exporters/debian/jessie build/*.deb
	package_cloud push lovooOS/prometheus-exporters/debian/stretch build/*.deb

.PHONY: all build build-deb clean deb format megacheck release-build release-package test vet
