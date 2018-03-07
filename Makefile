ifndef GOPATH
$(error GOPATH is not set)
endif

TESTS=$(shell go list ./... | grep -v /vendor/)

install: depends
	@go install .

depends:
	@git submodule update --init --recursive
	@go get -d .

# Automatially update code formatting
tidy:
	@find . \( -path ./vendor -o -path ./.git \) -prune -o -type f -print | grep -E '.*\.go$$' | xargs goimports -w
	@find . \( -path ./vendor -o -path ./.git \) -prune -o -type f -print | grep -E '.*\.go$$' | xargs gofmt -w -s

# Run code style and other checks
lint:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --enable=goimports --enable=misspell \
		--vendor \
		./...

test:
	@go test $(TESTS)

test-verbose:
	@go test -v $(TESTS)

# Build binaries for all OS/Architectures
cross-compile: depends
	@echo '=== Cross compiling... ==='
	@for GOOS in darwin linux; do \
		for GOARCH in amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -a \
				-ldflags '$(VERSION_LDFLAGS)' \
				-o build/bin/ebsmount-$$GOOS-$$GOARCH .; \
		done; \
	done

# Build docker image.
docker: cross-compile
	mkdir -p build/docker
	cp build/bin/ebsmount-linux-amd64 build/docker/ebsmount
	cp Dockerfile build/docker/
	cd build/docker/ && docker build -t adamstruck/ebsmount .
