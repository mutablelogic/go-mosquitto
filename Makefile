# Paths to packages
GO=$(shell which go)
SED=$(shell which sed)

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))
PLUGIN_DIR := $(wildcard plugin/*)

# Build flags
BUILD_MODULE = "github.com/djthorpe/go-mosquitto"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
BUILD_ARCH = $(shell $(GO) env GOARCH)
BUILD_PLATFORM = $(shell $(GO) env GOOS)

all: clean test server plugins cmd

cmd: dependencies mkdir $(CMD_DIR)

server: dependencies mkdir plugins
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/mutablelogic/go-server/cmd/server

plugins: $(PLUGIN_DIR)
	@echo Build httpserver
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/httpserver
	@echo Build env
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/env.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/env
	@echo Build log
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/log
	@echo Build sqlite3
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/sqlite3.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-sqlite/plugin/sqlite3


$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

deb: nfpm go-server-mqtt-deb

go-server-mqtt-deb: plugin/mqtt
	@echo Package go-server-mqtt deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-mqtt/nfpm.yaml > $(BUILD_DIR)/go-server-mqtt-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-mqtt-nfpm.yaml --packager deb --target $(BUILD_DIR)

nfpm:
	@echo Installing nfpm
	@${GO} mod tidy
	@${GO} install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.3.1

test:
	@echo Test sys/mosquitto
	@${GO} test ./sys/mosquitto
	@echo Test pkg/mosquitto
	@${GO} test ./pkg/mosquitto

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
