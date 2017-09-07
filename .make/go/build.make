# Infer GOOS and GOARCH
GOOS   ?= $(or $(word 1,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOOS`"))
GOARCH ?= $(or $(word 2,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOARCH`"))

ifeq ($(GOOS),darwin)
	CGO_LDFLAGS ?= -lmpsse
else
ifeq ($(CFG_SPI),ftdi)
	CGO_LDFLAGS ?= -lrt -lmpsse
else
	CGO_LDFLAGS ?= -lrt
endif
ifneq ($(SDKTARGETSYSROOT),)
	CGO_CFLAGS ?= -I$(SDKTARGETSYSROOT)/usr/include/libftdi1 -I$(SDKTARGETSYSROOT)/usr/include
endif
endif


# build
go.build: $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM)

# default main file
MAIN ?= ./main.go

# Time margin in milliseconds
ifeq ($(PLATFORM),multitech)
	SENDING_TIME_MARGIN = 100
else ifeq ($(PLATFORM),kerlink)
	SENDING_TIME_MARGIN = 60
endif

LD_FLAGS = -ldflags "-w -X main.version=${PKTFWD_VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE} -X github.com/TheThingsNetwork/packet_forwarder/pktfwd.platform=${PLATFORM} -X github.com/TheThingsNetwork/packet_forwarder/cmd.downlinksMargin=${SENDING_TIME_MARGIN}"

# Build the executable
$(RELEASE_DIR)/$(NAME)-%: $(shell $(GO_FILES)) vendor/vendor.json
	@$(log) "building" [$(GO_ENV) CC="$(CC)" GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) $(GO_ENV) $(GO) build -tags '$(HAL_CHOICE)' $(GO_FLAGS) $(LD_FLAGS) $(MAIN) ...]
	@$(GO_ENV) CC="$(CC)" GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) build -tags '$(HAL_CHOICE)' -o "$(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM)-$(CFG_SPI)" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# Build the executable in dev mode (much faster)
go.dev: GO_FLAGS =
go.dev: GO_ENV =
go.dev: BUILD_TYPE = dev
go.dev: $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM)

## link the executable to a simple name
$(RELEASE_DIR)/$(NAME): $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM)
	@$(log) "linking binary" [ln -sfr $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM) $(RELEASE_DIR)/$(NAME)]
	@ln -sfr $(RELEASE_DIR)/$(NAME)-$(GOOS)-$(GOARCH)-$(PLATFORM) $(RELEASE_DIR)/$(NAME)

go.link: $(RELEASE_DIR)/$(NAME)

go.link-dev: GO_FLAGS =
go.link-dev: GO_ENV =
go.link-dev: BUILD_TYPE = dev
go.link-dev: go.link

## initialize govendor
vendor/vendor.json:
	@$(log) initializing govendor
	@govendor init

# vim: ft=make
