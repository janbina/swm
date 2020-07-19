PKG := "github.com/janbina/swm"
VERSION := $(shell git describe --tags)

ifneq ($(strip $(VERSION)),)
    GOLDFLAGS += -X $(PKG)/internal/buildconfig.Version=$(VERSION)
endif

GOFLAGS = -ldflags "$(GOLDFLAGS)"

PREFIX    ?= /usr/local
BINPREFIX ?= $(PREFIX)/bin
MANPREFIX ?= $(PREFIX)/share/man
DOCPREFIX ?= $(PREFIX)/share/doc/swm
XSESSIONS ?= $(PREFIX)/share/xsessions

MD_DOCS    = README.md

build:
	go build -o bin/swm $(GOFLAGS) $(PKG)/cmd/swm
	go build -o bin/swmctl $(GOFLAGS) $(PKG)/cmd/swmctl

install: build
	mkdir -p "$(DESTDIR)$(BINPREFIX)"
	cp -pf bin/swm "$(DESTDIR)$(BINPREFIX)"
	cp -pf bin/swmctl "$(DESTDIR)$(BINPREFIX)"
	mkdir -p "$(DESTDIR)$(MANPREFIX)"/man1
	cp -p doc/swm.1 "$(DESTDIR)$(MANPREFIX)"/man1
	cp -Pp doc/swmctl.1 "$(DESTDIR)$(MANPREFIX)"/man1
	mkdir -p "$(DESTDIR)$(DOCPREFIX)"
	cp -p $(MD_DOCS) "$(DESTDIR)$(DOCPREFIX)"
	mkdir -p "$(DESTDIR)$(DOCPREFIX)"/examples
	cp -pr examples/* "$(DESTDIR)$(DOCPREFIX)"/examples
	mkdir -p "$(DESTDIR)$(XSESSIONS)"
	cp -p swm.desktop "$(DESTDIR)$(XSESSIONS)"

uninstall:
	rm -f "$(DESTDIR)$(BINPREFIX)"/swm
	rm -f "$(DESTDIR)$(BINPREFIX)"/swmctl
	rm -f "$(DESTDIR)$(MANPREFIX)"/man1/swm.1
	rm -f "$(DESTDIR)$(MANPREFIX)"/man1/swmctl.1
	rm -rf "$(DESTDIR)$(DOCPREFIX)"
	rm -f "$(DESTDIR)$(XSESSIONS)"/swm.desktop

clean:
	rm -rf bin

.PHONY = build install uninstall clean
