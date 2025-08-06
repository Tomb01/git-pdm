WIX := wix
WXS_FILE := installer/Package.wxs
MSI := dist/git-pdm.msi
PROJECTDIR := $(shell pwd)

all: $(MSI)

$(MSI): $(WXS_FILE)
	PROJECTDIR=$(PROJECTDIR) $(WIX) build $(PROJECTDIR)/installer/Package.wxs $(PROJECTDIR)/installer/UI.wxs -o $(MSI) -ext $(PROJECTDIR)/installer/WixToolset.UI.wixext.dll

clean:
	rm -f $(MSI)
