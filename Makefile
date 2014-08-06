gopath = GOPATH=$(CURDIR)
cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/src/erasure/longhair/bin/liblonghair.a -lstdc++"
govars = $(gopath) $(cgo_ldflags)
packages = logger network siacrypto siaencoding siafiles erasure state delta consensus client client-cli

all: submodule-update install

submodule-update:
	git submodule update --init

directories:
	mkdir -p filesCreatedDuringTesting

fmt:
	$(govars) go fmt $(packages)

install: fmt
	$(govars) go install $(packages)

release: fmt
	$(govars) go install -ldflags '-extldflags "-static"' $(packages)
	tar -cJvf release.xz bin

test:
	$(govars) go test -short $(packages)

test-verbose:
	$(govars) go test -short -v $(packages)

test-race:
	$(govars) go test -short -race $(packages)

test-race-verbose:
	$(govars) go test -short -race -v $(packages)

test-long:
	$(govars) go test -race $(packages)

test-long-verbose:
	$(govars) go test -v -race $(packages)

test-consensus:
	$(govars) go test -v -race consensus

test-delta:
	$(govars) go test -v -race delta

test-state:
	$(govars) go test -v -race state

dependencies: submodule-update race-libs directories
	cd src/siacrypto/libsodium && ./autogen.sh && ./configure && make check && sudo make install && sudo ldconfig

race-libs:
	$(govars) go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt install test test-verbose test-race test-race-verbose test-long test-long-verbose test-consensus test-delta test-state dependencies race-libs docs directories
