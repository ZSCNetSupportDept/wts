.PHONY:start-server dev server tool build-all start-tool

PWD= $(shell pwd)
SRC= $(PWD)/src
BUILD= $(PWD)/build
DEV_CONFIG= $(BUILD)/dev.yaml
SERVER_BINARY_NAME=wts
TOOL_BINARY_NAME=wtstool
VERSION ?= dev

build-all: $(SERVER_BINARY_NAME) $(TOOL_BINARY_NAME)

start-server: $(SERVER_BINARY_NAME)
	$(BUILD)/$(SERVER_BINARY_NAME) -c $(DEV_CONFIG)

dev: $(SERVER_BINARY_NAME) start-server

server: $(SERVER_BINARY_NAME)
$(SERVER_BINARY_NAME): 
	cd $(SRC)/cmd/wts-server && go build -o $(BUILD)/$(SERVER_BINARY_NAME)

tool: $(TOOL_BINARY_NAME)
$(TOOL_BINARY_NAME): 
	mkdir -p $(BUILD)
	cd $(SRC)/cmd/wtstool && go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD)/$(TOOL_BINARY_NAME)

start-tool: $(TOOL_BINARY_NAME)
	$(BUILD)/$(TOOL_BINARY_NAME) -c $(DEV_CONFIG) $(ARGS)
