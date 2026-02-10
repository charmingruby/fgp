.PHONY: fmt vet test race lint ci tag

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run ./...

ci: fmt vet test race lint

tag:
	@if [ -z "$(V)" ]; then echo "usage: make tag V=vX.Y.Z"; exit 1; fi
	git tag -a $(V) -m "Release $(V)"
