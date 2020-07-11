PKG := "github.com/janbina/swm"
VERSION := $(shell git describe --tags)

ifneq ($(strip $(VERSION)),)
    GOLDFLAGS += -X $(PKG)/internal/buildconfig.Version=$(VERSION)
endif

GOFLAGS = -ldflags "$(GOLDFLAGS)"

build:
	go build -o bin/swm $(GOFLAGS) $(PKG)/cmd/swm
	go build -o bin/swmctl $(GOFLAGS) $(PKG)/cmd/swmctl

clean:
	rm -rf bin

.PHONY = build clean
