.PHONY: build
build:
	go build -v -o udpdemux .

.PHONY: test
test: build

.PHONY: lint
lint:
	golangci-lint run
