gopath = GOPATH=$(CURDIR)
cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/src/erasure/longhair/bin/liblonghair.a -lstdc++"
govars = $(gopath) $(cgo_ldflags)
packages = logger network siacrypto siaencoding erasure quorum quorum/script participant client server

all: submodule-update libraries

submodule-update:
	git submodule update --init

directories:
	mkdir -p participantStorage

fmt:
	$(govars) go fmt $(packages)

libraries: fmt
	$(govars) go install $(packages)

test: libraries
	$(govars) go test -short $(packages)

test-verbose: libraries
	$(govars) go test -short -v $(packages)

test-race: libraries
	$(govars) go test -short -race $(packages)

test-race-verbose: libraries
	$(govars) go test -short -race -v $(packages)

test-long: libraries
	$(govars) go test -race $(packages)

test-long-verbose: libraries
	$(govars) go test -v -race $(packages)

test-participant: libraries
	$(govars) go test -v -race participant

test-quorum: libraries
	$(govars) go test -v -race quorum

dependencies: submodule-update race-libs directories

race-libs:
	$(govars) go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt libraries test test-verbose test-race test-race-verbose test-long test-long-verbose test-participant test-quorum dependencies race-libs docs directories
