.DEFAULT_GOAL := help

GO ?= go
# Use an isolated, npm-cached toolchain when the shell's Node is too old.
# npx adds both packages to PATH for npm and all of its build/test subprocesses.
NODE_SUPPORTED := $(shell node -e 'const [a,b]=process.versions.node.split(".").map(Number);process.exit(a>22||(a===22&&b>=12)?0:1)' >/dev/null 2>&1 && echo yes)
ifeq ($(NODE_SUPPORTED),yes)
NPM ?= npm
else
NPM ?= npx --yes --package=node@22 --package=npm@10 -- npm
endif
HOST ?= 127.0.0.1
PORT ?= 8080
# Preserve the original default history; alternate ports get independent state.
SIM_STATE ?= $(if $(filter 8080,$(PORT)),.depthpilot-sim-home.json,.depthpilot-sim-$(PORT)-home.json)
URL = http://127.0.0.1:$(PORT)

.PHONY: help check-node setup build sim open test test-go test-ui

help:
	@echo "DepthPilot simulator"
	@echo "  make sim       Install dependencies, build, and start the simulator"
	@echo "  make open      Open the UI in your browser (simulator must be running)"
	@echo "  make setup     Install frontend and Go dependencies only"
	@echo "  make build     Build the UI and Go web service"
	@echo "  make test      Run Go checks and browser tests (installs Chromium)"
	@echo "  make test-go   Run Go race tests and vet"
	@echo "  make test-ui   Build and run browser tests against a dedicated simulator"
	@echo ""
	@echo "Requires npm/npx and Go. Node 22 is bootstrapped if your Node is older."
	@echo "Default UI: $(URL). Stop the simulator with Ctrl+C."
	@echo "Override PORT=8081, or HOST=0.0.0.0 for access from the local network."
	@echo "Each port has separate saved home/activity. Override SIM_STATE=path if needed."

check-node:
	@command -v npm >/dev/null 2>&1 || { echo "Install Node.js/npm first."; exit 1; }
ifneq ($(NODE_SUPPORTED),yes)
	@command -v npx >/dev/null 2>&1 || { echo "npx is required to bootstrap Node 22."; exit 1; }
	@echo "Using isolated Node 22 + npm 10 (cached by npx; global Node stays unchanged)."
endif

# Keep the dependency stamp inside the ignored node_modules directory so a
# removed install, or an updated package/lock file, causes a fresh npm ci.
web/node_modules/.make-ready: web/package.json web/package-lock.json | check-node
	$(NPM) --prefix web ci
	@touch $@

setup: web/node_modules/.make-ready
	cd controller && $(GO) mod download

build: web/node_modules/.make-ready
	$(NPM) --prefix web run build
	cd controller && $(GO) build -o depthpilot-web ./cmd/depthpilot-web

sim: build
	@echo "Open $(URL) in your browser. Simulation only; no serial devices are opened."
	@echo "For LAN access, open http://<this-machine-IP>:$(PORT). Press Ctrl+C to stop."
	cd controller && exec ./depthpilot-web -simulate -listen $(HOST):$(PORT) -state "$(SIM_STATE)" -web ../web/dist

open:
	@if command -v open >/dev/null 2>&1; then open "$(URL)"; \
	elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$(URL)"; \
	else echo "Open $(URL) in your browser."; fi

test: test-go test-ui

test-go:
	cd controller && $(GO) test -race ./...
	cd controller && $(GO) vet ./...

test-ui: build
	cd web && $(NPM) exec -- playwright install chromium
	$(NPM) --prefix web test
