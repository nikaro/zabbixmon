PREFIX?=/usr/local
_INSTDIR=${DESTDIR}${PREFIX}
BINDIR?=${_INSTDIR}/bin
MANDIR?=${_INSTDIR}/share/man
APP=zabbixmon

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: build

.PHONY: build
## build: Build for the current target
build:
	@echo "Building..."
	env CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o build/${APP}-${GOOS}-${GOARCH} .

.PHONY: build-all
## build-all: Build for all targets
build-all:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) build
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(MAKE) build
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(MAKE) build
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(MAKE) build
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(MAKE) build
	env CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(MAKE) build

.PHONY: install
## install: Install the application
install:
	@echo "Installing..."
	install build/${APP}-${GOOS}-${GOARCH} ${BINDIR}/${APP}

.PHONY: uninstall
## uninstall: Uninstall the application
uninstall:
	@echo "Uninstalling..."
	rm -rf ${BINDIR}/${APP}

.PHONY: run
## run: Runs go run
run:
	go run -race .

.PHONY: clean
## clean: Cleans the binary
clean:
	@echo "Cleaning..."
	rm -rf build/
	rm -rf dist/

.PHONY: precommit
## precommit: Setup pre-commit hooks
precommit:
	pre-commit install
	pre-commit install --hook-type commit-msg

.PHONY: setup
## setup: Setup go modules
setup:
	go get -u all
	go mod tidy

.PHONY: lint
## lint: Runs linter on the project
lint:
	go vet ./...
	go fix ./...
	staticcheck ./...
	govulncheck ./...
	golangci-lint run

.PHONY: format
## format: Runs goimports on the project
format:
	go fmt ./...

.PHONY: test
## test: Runs go test
test:
	go test ./...

.PHONY: help
## help: Prints this help message
help:
	@echo -e "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
