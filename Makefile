# Bootstap Requirements
BOOTREQS = \
	github.com/golang/lint/golint \
	github.com/alecthomas/gometalinter

.PHONY: generated run clean static bootstrap $(BOOTREQS) install test lint

$(BOOTREQS):
	go get -u $@

bootstrap: $(BOOTREQS)
	@which glide || curl http://glide.sh/get | sh

protoc:
	@which protoc || (echo 'Protocol Buffers is required. Install protoc' && exit 1)

vendor: bootstrap glide.lock
	glide install

bin:
	mkdir -p bin

bin/waffyd: vendor bin generated
	go build ./cmd/waffyd/
	mv waffyd bin/

bin/waffy: vendor bin generated
	go build ./cmd/waffy/
	mv waffy bin/

generated: protoc bootstrap vendor
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
		./pkg/...

# Tests

test:
	go test ./pkg/...