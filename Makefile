cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/erasure/longhair/bin/liblonghair.a -lstdc++"

all: submodule-update install

submodule-update:
	git submodule update --init

fmt:
	$(cgo_ldflags) go fmt ./...

install: fmt
	$(cgo_ldflags) go install ./...

release: fmt
	$(cgo_ldflags) go install -ldflags '-extldflags "-static"' ./...
	cd $(GOPATH) && tar -cJvf release.xz bin/server bin/client-cli
	mv $(GOPATH)/release.xz .

test:
	$(cgo_ldflags) go test -short ./...

test-verbose:
	$(cgo_ldflags) go test -short -v ./...

test-race:
	$(cgo_ldflags) go test -short -race ./...

test-race-verbose:
	$(cgo_ldflags) go test -short -race -v ./...

test-long:
	$(cgo_ldflags) go test -race ./...

test-long-verbose:
	$(cgo_ldflags) go test -v -race ./...

test-consensus:
	$(cgo_ldflags) go test -v -race ./consensus

test-delta:
	$(cgo_ldflags) go test -v -race ./delta

test-state:
	$(cgo_ldflags) go test -v -race ./state

cover-set:
	mkdir -p cover
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/consensus-set.out ./consensus
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/delta-set.out ./delta
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/client-set.out ./client
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/erasure-set.out ./erasure
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/network-set.out ./network
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/siacrypto-set.out ./siacrypto
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/siaencoding-set.out ./siaencoding
	$(cgo_ldflags) go test -covermode=set -coverprofile=cover/state-set.out ./state
	$(cgo_ldflags) go tool cover -html=cover/consensus-set.out -o=cover/consensus-set.html
	$(cgo_ldflags) go tool cover -html=cover/delta-set.out -o=cover/delta-set.html
	$(cgo_ldflags) go tool cover -html=cover/client-set.out -o=cover/client-set.html
	$(cgo_ldflags) go tool cover -html=cover/erasure-set.out -o=cover/erasure-set.html
	$(cgo_ldflags) go tool cover -html=cover/network-set.out -o=cover/network-set.html
	$(cgo_ldflags) go tool cover -html=cover/siacrypto-set.out -o=cover/siacrypto-set.html
	$(cgo_ldflags) go tool cover -html=cover/siaencoding-set.out -o=cover/siaencoding-set.html
	$(cgo_ldflags) go tool cover -html=cover/state-set.out -o=cover/state-set.html

cover-count:
	mkdir -p cover
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/consensus-count.out ./consensus
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/delta-count.out ./delta
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/client-count.out ./client
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/erasure-count.out ./erasure
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/network-count.out ./network
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/siacrypto-count.out ./siacrypto
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/siaencoding-count.out ./siaencoding
	$(cgo_ldflags) go test -covermode=count -coverprofile=cover/state-count.out ./state
	$(cgo_ldflags) go tool cover -html=cover/consensus-count.out -o=cover/consensus-count.html
	$(cgo_ldflags) go tool cover -html=cover/delta-count.out -o=cover/delta-count.html
	$(cgo_ldflags) go tool cover -html=cover/client-count.out -o=cover/client-count.html
	$(cgo_ldflags) go tool cover -html=cover/erasure-count.out -o=cover/erasure-count.html
	$(cgo_ldflags) go tool cover -html=cover/network-count.out -o=cover/network-count.html
	$(cgo_ldflags) go tool cover -html=cover/siacrypto-count.out -o=cover/siacrypto-count.html
	$(cgo_ldflags) go tool cover -html=cover/siaencoding-count.out -o=cover/siaencoding-count.html
	$(cgo_ldflags) go tool cover -html=cover/state-count.out -o=cover/state-count.html

cover-atomic:
	mkdir -p cover
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/consensus-atomic.out ./consensus
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/delta-atomic.out ./delta
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/client-atomic.out ./client
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/erasure-atomic.out ./erasure
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/network-atomic.out ./network
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/siacrypto-atomic.out ./siacrypto
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/siaencoding-atomic.out ./siaencoding
	$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/state-atomic.out ./state
	$(cgo_ldflags) go tool cover -html=cover/consensus-atomic.out -o=cover/consensus-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/delta-atomic.out -o=cover/delta-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/client-atomic.out -o=cover/client-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/erasure-atomic.out -o=cover/erasure-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/network-atomic.out -o=cover/network-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/siacrypto-atomic.out -o=cover/siacrypto-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/siaencoding-atomic.out -o=cover/siaencoding-atomic.html
	$(cgo_ldflags) go tool cover -html=cover/state-atomic.out -o=cover/state-atomic.html

cover: cover-set

bench:
	$(cgo_ldflags) go test -run=XXX -bench=. ./...

dependencies: submodule-update race-libs
	cd siacrypto/libsodium && ./autogen.sh && ./configure && make check && sudo make install && sudo ldconfig

race-libs:
	go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt install test test-verbose test-race test-race-verbose test-long test-long-verbose test-consensus test-delta test-state dependencies race-libs docs
