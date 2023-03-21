PWD := $(shell pwd)
all: build

build:
	@echo "Building snapidff binary to './snapdiff'"
	@CGO_ENABLED=0 go build -o $(PWD)/snapdiff 1>/dev/null
clean:
	@rm -rvf snapdiff
