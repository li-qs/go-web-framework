NAME      = myframework
BUILD_DIR = $(CURDIR)/build

.PHONY: all clean
all: clean api

clean:
	rm -rf $(BUILD_DIR)

api: api-linux-amd64 api-darwin-amd64 api-darwin-arm64

api-linux-amd64 api-darwin-amd64 api-darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$$(echo $@ | sed 's/^api-//;s/-.*//') GOARCH=$$(echo $@ | sed 's/^api-//;s/.*-//') \
		go build -o $(BUILD_DIR)/$(NAME)-$@ ./cmd/api
