.PHONY: build integration-test unit-test lint imports obs-stack-up obs-stack-down run-example run-test

build:
	go build -v ./...

integration-test:
	go test -race -v ./tests --tags=integration

unit-test:
	go test -race -v ./... -coverprofile=unit-test.cov -covermode=atomic

lint:
	@golangci-lint run --config=.golangci.yml

imports:
	goimports -w .

obs-stack-up:
	docker-compose -f ./tests/testdata/docker-compose.yaml up -d

obs-stack-down:
	docker-compose -f ./tests/testdata/docker-compose.yaml down

run-example:
	make obs-stack-up
	cd ./example && go build -v ./... && ./example

run-test:
	make unit-test
	make integration-test

