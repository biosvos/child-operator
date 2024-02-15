.PHONY: build
build:
	go build

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor

.PHONY: validate
validate:
	# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# sudo cp ~/go/bin/golangci-lint /usr/local/bin/
	golangci-lint run
	go test ./...
	go build