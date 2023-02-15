build:
	go build cmd/cosmos-wallets-exporter.go

install:
	go install cmd/cosmos-wallets-exporter.go

lint:
	golangci-lint run --fix ./...
