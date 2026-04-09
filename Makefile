BINARY    := xampp-tui
MAIN      := ./cmd/lampp-tui
INSTALL   := /usr/local/bin
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION)"
ARCHIVE   := $(BINARY)-linux-amd64.tar.gz

.PHONY: build install uninstall release clean run

## build — compile a development binary in the current directory
build:
	go build $(LDFLAGS) -o $(BINARY) $(MAIN)

## run — build and launch immediately (requires sudo for lampp commands)
run: build
	./$(BINARY)

## install — build an optimised binary and copy it to $(INSTALL)
install: build
	sudo install -m 755 $(BINARY) $(INSTALL)/$(BINARY)
	@echo "Installed $(INSTALL)/$(BINARY)"

## uninstall — remove the installed binary
uninstall:
	sudo rm -f $(INSTALL)/$(BINARY)
	@echo "Removed $(INSTALL)/$(BINARY)"

## release — build a stripped binary and create a distributable tarball
release: build
	tar czf $(ARCHIVE) $(BINARY) install.sh README.md
	sha256sum $(ARCHIVE) > $(ARCHIVE).sha256
	@echo ""
	@echo "Release artefacts:"
	@ls -lh $(ARCHIVE) $(ARCHIVE).sha256

## clean — remove build artefacts
clean:
	rm -f $(BINARY) $(ARCHIVE) $(ARCHIVE).sha256
