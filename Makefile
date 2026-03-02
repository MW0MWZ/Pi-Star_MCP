BINARY   := pistar-dashboard
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
BUILDDIR := build

.PHONY: build linux-arm linux-arm64 linux-amd64 all clean

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/$(BINARY)-linux-arm .

linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/$(BINARY)-linux-arm64 .

linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/$(BINARY)-linux-amd64 .

all: linux-arm linux-arm64 linux-amd64

clean:
	rm -f $(BINARY)
	rm -rf $(BUILDDIR)
