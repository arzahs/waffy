# Bootstap Requirements
BOOTREQS = \
	github.com/golang/lint/golint \
	github.com/alecthomas/gometalinter \

.PHONY: $(BOOTREQS) generated run clean static bootstrap $(BOOTREQS) install test lint

$(BOOTREQS):
	go get -u $@

bootstrap: $(BOOTREQS)
	go get github.com/smartystreets/goconvey
	@which glide || curl http://glide.sh/get | sh

protoc:
	@which protoc || (echo 'Protocol Buffers is required. Install protoc' && exit 1)

glide.lock:
	touch $@

vendor: glide.lock glide.yaml
	glide install

bin:
	mkdir -p bin

bin/waffyd: vendor bin
	CGO_ENABLED=0 go build -race -i -v -o bin/waffyd ./cmd/waffyd/

bin/waffy: vendor bin
	CGO_ENABLED=0 go build -race -i -v -o bin/waffy ./cmd/waffy/

services:
	mkdir -p services

generated: services protoc bootstrap vendor
	go generate

clean:
	rm bin/*

# Linting

install-linters:
	gometalinter --install

lint: install
	golint ./pkg/...
	go vet ./pkg/...

lint-next: install
	gometalinter \
		--concurrency=2 --deadline=1m --sort=path \
		--disable=dupl --disable=vetshadow --enable=misspell \
		--disable=unconvert \
		--exclude=".*\.pb\.go" --exclude=".*_test.go" \
		--cyclo-over=12 ./pkg/...

# Tests

test:
	go test -race ./pkg/... -v

test-web:
	goconvey -excludedDirs=protos,vendor

