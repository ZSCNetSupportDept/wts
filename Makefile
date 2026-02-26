
SHELL := /bin/bash

.DEFAULT_GOAL := help

BACK_DIR := back
FRONT_DIR := front

.PHONY: help
help:
	@printf '%s\n' \
		'Usage:' \
		'  make <target>' \
		'' \
		'Common targets:' \
		'  dev            Run front+back dev servers' \
		'  build          Build front+back' \
		'  test           Run front+back tests' \
		'  clean          Remove build artifacts' \
		'' \
		'Backend targets:' \
		'  back-build     Build server+tool binaries' \
		'  back-dev       Build and run server (dev config)' \
		'  back-run       Run server (dev config)' \
		'  back-tool      Build tool binary' \
		'  back-run-tool  Run tool (dev config)' \
		'  back-test      go test ./... (in back/src)' \
		'  back-fmt       gofmt ./... (in back/src)' \
		'' \
		'Frontend targets:' \
		'  front-install  npm ci (in front)' \
		'  front-dev      npm run dev (in front)' \
		'  front-build    npm run build (in front)' \
		'  front-preview  npm run preview (in front)' \
		'  front-check    npm run check (in front)' \
		'  front-lint     npm run lint (in front)' \
		'  front-format   npm run format (in front)' \
		'  front-test     npm test (in front)'

.PHONY: dev build test clean install doctor

dev:
	@bash -c 'set -euo pipefail; \
		$(MAKE) dev-front & pf=$$!; \
		$(MAKE) dev-back & pb=$$!; \
		trap "kill $$pf $$pb 2>/dev/null || true" INT TERM EXIT; \
		wait $$pf $$pb'

build: back-build front-build

test: back-test front-test

clean: back-clean front-clean

install: back-install front-install

doctor:
	@command -v go >/dev/null 2>&1 && go version || echo 'go: not found'
	@command -v node >/dev/null 2>&1 && node --version || echo 'node: not found'
	@command -v npm >/dev/null 2>&1 && npm --version || echo 'npm: not found'

## Backend
.PHONY: dev-back back-dev back-build back-server back-tool back-run back-run-tool back-test back-fmt back-clean back-install

dev-back: back-dev

back-dev:
	@$(MAKE) -C $(BACK_DIR) dev

back-build:
	@$(MAKE) -C $(BACK_DIR) build-all

back-server:
	@$(MAKE) -C $(BACK_DIR) server

back-tool:
	@$(MAKE) -C $(BACK_DIR) tool

back-run:
	@$(MAKE) -C $(BACK_DIR) start-server

back-run-tool:
	@$(MAKE) -C $(BACK_DIR) start-tool

back-test:
	@cd $(BACK_DIR)/src && go test ./...

back-fmt:
	@cd $(BACK_DIR)/src && gofmt -w ./

back-clean:
	@rm -f $(BACK_DIR)/build/wts $(BACK_DIR)/build/wtstool

back-install:
	@cd $(BACK_DIR)/src && go mod download

## Frontend
.PHONY: dev-front front-install front-dev front-build front-preview front-check front-lint front-format front-test front-clean

dev-front: front-dev

front-install:
	@npm --prefix $(FRONT_DIR) ci

front-dev:
	@npm --prefix $(FRONT_DIR) run dev

front-build:
	@npm --prefix $(FRONT_DIR) run build

front-preview:
	@npm --prefix $(FRONT_DIR) run preview

front-check:
	@npm --prefix $(FRONT_DIR) run check

front-lint:
	@npm --prefix $(FRONT_DIR) run lint

front-format:
	@npm --prefix $(FRONT_DIR) run format

front-test:
	@npm --prefix $(FRONT_DIR) test

front-clean:
	@rm -rf $(FRONT_DIR)/build $(FRONT_DIR)/.svelte-kit $(FRONT_DIR)/.vite
