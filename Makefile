BINARY    := xampp-tui
MAIN      := ./cmd/lampp-tui
INSTALL   := /usr/local/bin
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION)"
ARCHIVE_LINUX   := $(BINARY)-linux-amd64.tar.gz
ARCHIVE_WINDOWS := $(BINARY)-windows-amd64.zip

.PHONY: build build-windows install uninstall release release-windows clean run

## build — compile a development binary for Linux
build:
	go build $(LDFLAGS) -o $(BINARY) $(MAIN)

## build-windows — cross-compile for Windows (produces xampp-tui.exe)
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY).exe $(MAIN)

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

## release — build stripped Linux binary and create distributable tarball
release: build
	tar czf $(ARCHIVE_LINUX) $(BINARY) install.sh README.md
	sha256sum $(ARCHIVE_LINUX) > $(ARCHIVE_LINUX).sha256
	@echo ""
	@echo "Release artefacts:"
	@ls -lh $(ARCHIVE_LINUX) $(ARCHIVE_LINUX).sha256

## release-windows — cross-compile for Windows and create distributable zip
release-windows: build-windows
	zip $(ARCHIVE_WINDOWS) $(BINARY).exe README.md
	sha256sum $(ARCHIVE_WINDOWS) > $(ARCHIVE_WINDOWS).sha256
	@echo ""
	@echo "Release artefacts:"
	@ls -lh $(ARCHIVE_WINDOWS) $(ARCHIVE_WINDOWS).sha256

## clean — remove build artefacts
clean:
	rm -f $(BINARY) $(BINARY).exe $(ARCHIVE_LINUX) $(ARCHIVE_LINUX).sha256 $(ARCHIVE_WINDOWS) $(ARCHIVE_WINDOWS).sha256
