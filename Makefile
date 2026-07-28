NAME        = api
BUILD_DIR   = $(CURDIR)/build
PLATFORMS   = linux-amd64 darwin-amd64 darwin-arm64

# ./cmd/ 下有什么子目录，添加到这里即可
BINS = server

.PHONY: all clean $(BINS)

all: clean $(BINS)

clean:
	rm -rf $(BUILD_DIR)

define build_bin
$(1): $(addprefix $(1)-, $(PLATFORMS))

$(1)-%:
	@mkdir -p $(BUILD_DIR)
	GOOS=$$(word 1, $$(subst -, , $$*)) \
	GOARCH=$$(word 2, $$(subst -, , $$*)) \
	CGO_ENABLED=0 \
	go build -o $(BUILD_DIR)/$(NAME)-$(1)-$$* ./cmd/$(1)
endef

$(foreach bin,$(BINS),$(eval $(call build_bin, $(bin))))
