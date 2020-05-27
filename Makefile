PKG := "github.com/janbina/swm"
VERSION := $(shell git describe --tags)

ifneq ($(strip $(VERSION)),)
GOLDFLAGS += -X $(PKG)/internal/buildconfig.Version=$(VERSION)
endif

GOFLAGS = -ldflags "$(GOLDFLAGS)"

run: build
	./swm

build:
	go build -o swm $(GOFLAGS) $(PKG)/cmd/swm
	go build -o swmctl $(GOFLAGS) $(PKG)/cmd/swmctl
