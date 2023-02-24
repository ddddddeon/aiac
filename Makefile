GO = go
BINARY = aiac
BINDIR = /usr/bin

.PHONY: build

build:
	$(GO) build

install:
	mv $(BINARY) $(BINDIR)/$(BINARY)


