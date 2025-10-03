BIN_DIR := bin
TITLE := nas-updater

VERSION := $(shell jq -r '.Version' ./config.json)
KOMGA_SERVICE_FILE := $(shell jq -r '.KomgaServiceFile' ./config.json)
KOMGA_DIR := $(shell jq -r '.KomgaDir' ./config.json)
IMMICH_DIR := $(shell jq -r '.ImmichDir' ./config.json)

LD_FLAGS := -ldflags "\
	-s -w \
	-X main.Version=$(VERSION) \
	-X main.KomgaServiceFile=$(KOMGA_SERVICE_FILE) \
	-X main.KomgaDir=$(KOMGA_DIR) \
	-X main.ImmichDir=$(IMMICH_DIR) \
"

OS_ARCH := \
	linux_amd64 linux_arm linux_arm64 linux_ppc64 linux_ppc64le \
	linux_mips linux_mipsle linux_mips64 linux_mips64le linux_s390x \
	darwin_amd64 darwin_arm64 \
	freebsd_amd64 freebsd_386 \
	openbsd_amd64 openbsd_386 openbsd_arm64 \
	netbsd_amd64 netbsd_386 netbsd_arm \
	dragonfly_amd64 \
	solaris_amd64 \
	plan9_386 plan9_amd64

RED := \033[0;31m
GREEN := \033[0;32m
NC := \033[0m

.PHONY: all clean host_default cross

all: host_default cross

host_default:
	@mkdir -p $(BIN_DIR)
	@echo "Building host binary..."
	@CGO_ENABLED=0 go build $(LD_FLAGS) -trimpath -buildvcs=false -o $(BIN_DIR)/$(TITLE) ./ && \
		printf '$(GREEN)Build succeeded: host_default$(NC)\n' || \
		(printf '$(RED)Build failed: host_default$(NC)\n' && exit 1)

cross: $(OS_ARCH)

$(OS_ARCH):
	@mkdir -p $(BIN_DIR)
	@OS=$$(echo $@ | cut -d_ -f1); \
	ARCH=$$(echo $@ | cut -d_ -f2); \
	OUT=$(BIN_DIR)/$(TITLE)-$$OS-$$ARCH; \
	echo "Building $@..."; \
	CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build $(LD_FLAGS) -trimpath -buildvcs=false -o $$OUT ./ && \
	printf '$(GREEN)Build succeeded: $@$(NC)\n' || \
	(printf '$(RED)Build failed: $@$(NC)\n' && exit 1)

clean:
	@rm -rf $(BIN_DIR)
	@echo "Cleaned binaries."
