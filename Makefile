install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/abice/go-enum@latest
lint:
	golangci-config-generator
	golangci-lint run -v
vet:
	go vet  ./...

fmt:
	gofumpt -l -w .
	goimports -l -w .
