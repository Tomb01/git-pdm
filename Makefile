WIX := wix
WXS_FILE := installer\Package.wxs
MSI := dist/git-pdm_win.msi
PROJECTDIR := $(shell pwd)
PROJECTDIR_WIN := $(shell powershell -NoProfile -Command "[System.Environment]::CurrentDirectory")

ICON_SIZES := 16 32 48 64 256
ICON_SRC := $(PROJECTDIR_WIN)\icon.svg
ICON_DIR := $(PROJECTDIR_WIN)\ignore

.PHONY: dist
.PHONY: build

dist:
	go build -o .\build\git-pdm.exe
	PROJECTDIR=$(PROJECTDIR) $(WIX) build $(PROJECTDIR)/installer/Package.wxs $(PROJECTDIR)/installer/UI.wxs -o $(MSI) -ext $(PROJECTDIR)/installer/WixToolset.UI.wixext.dll

icon:
	@mkdir -p $(ICON_DIR)
	@for size in $(ICON_SIZES); do \
		inkscape -o $(ICON_DIR)\icon_$${size}.png -w $$size -h $$size $(ICON_SRC); \
	done
	magick $(foreach size,$(ICON_SIZES),$(ICON_DIR)\icon_$(size).png) $(PROJECTDIR_WIN)\icon.ico

build:
	go build -o .\build\git-pdm.exe

clean:
	rm -f $(MSI)
