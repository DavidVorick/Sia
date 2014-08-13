cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/erasure/longhair/bin/liblonghair.a -lstdc++"

all: submodule-update install

submodule-update:
	git submodule update --init

fmt:
	go fmt ./...

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

bench:
	$(cgo_ldflags) go test -run=XXX -bench=. ./...

dependencies: submodule-update race-libs
	cd siacrypto/libsodium && ./autogen.sh && ./configure && make check && sudo make install && sudo ldconfig

race-libs:
	go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt install test test-verbose test-race test-race-verbose test-long test-long-verbose test-consensus test-delta test-state dependencies race-libs docs
