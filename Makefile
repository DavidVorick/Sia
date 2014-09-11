cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/erasure/longhair/bin/liblonghair.a -lstdc++"
packages = consensus delta erasure network server siacrypto siaencoding state

all: submodule-update install

submodule-update:
	git submodule update --init

fmt:
	go fmt ./...

install: fmt
	$(cgo_ldflags) go install ./...

release: fmt
	$(cgo_ldflags) go install -ldflags '-extldflags "-static"' ./...
	cp $(GOPATH)/bin/client-cli sia-cli
	tar -cJvf release.xz sia-cli Release.md && rm -f sia-cli

test:
	$(cgo_ldflags) go test -short ./...

test-verbose:
	$(cgo_ldflags) go test -short -v ./...

test-race:
	$(cgo_ldflags) go test -short -race ./...

test-race-verbose:
	$(cgo_ldflags) go test -short -race -v ./...

test-long:
	$(cgo_ldflags) go test -race -timeout 1h ./...

test-long-verbose:
	$(cgo_ldflags) go test -v -race -timeout 1h ./...

test-consensus:
	$(cgo_ldflags) go test -v -race -timeout 1h ./consensus

test-delta:
	$(cgo_ldflags) go test -v -race -timeout 1h ./delta

test-server:
	$(cgo_ldflags) go test -v -race -timeout 1h ./server

test-state:
	$(cgo_ldflags) go test -v -race -timeout 1h ./state

cover-set:
	@mkdir -p cover
	@for package in $(packages); do \
		$(cgo_ldflags) go test -covermode=set -coverprofile=cover/$$package-set.out ./$$package ; \
		$(cgo_ldflags) go tool cover -html=cover/$$package-set.out -o=cover/$$package-set.html ; \
		rm cover/$$package-set.out ; \
	done

cover-count:
	@mkdir -p cover
	@for package in $(packages); do \
		$(cgo_ldflags) go test -covermode=count -coverprofile=cover/$$package-count.out ./$$package ; \
		$(cgo_ldflags) go tool cover -html=cover/$$package-count.out -o=cover/$$package-count.html ; \
		rm cover/$$package-count.out ; \
	done

cover-atomic:
	@mkdir -p cover
	@for package in $(packages); do \
		$(cgo_ldflags) go test -covermode=atomic -coverprofile=cover/$$package-atomic.out ./$$package ; \
		$(cgo_ldflags) go tool cover -html=cover/$$package-atomic.out -o=cover/$$package-atomic.html ; \
		rm cover/$$package-atomic.out ; \
	done

cover: cover-set

bench:
	$(cgo_ldflags) go test -run=XXX -bench=. ./...

dependencies: submodule-update race-libs
	cd siacrypto/libsodium && sudo ./autogen.sh && sudo ./configure && sudo make check && sudo make install && sudo ldconfig
	go get -u code.google.com/p/gcfg
	go get -u github.com/spf13/cobra

race-libs:
	go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 
	pdflatex -output-directory=doc/ doc/addressing.tex

.PHONY: all submodule-update fmt install test test-verbose test-race test-race-verbose test-long test-long-verbose test-consensus test-delta test-state dependencies race-libs docs
