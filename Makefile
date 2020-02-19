# Go parameters
GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 

darwin: PREFIX=/usr/local
darwin: test

test: 
	PKG_CONFIG_PATH="$(PREFIX)/lib/pkgconfig" $(GOTEST) -v ./...

clean: 
	$(GOCLEAN)