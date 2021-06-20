export GO111MODULE=on

build:
	cd cmd/bookkeeper && go build -mod=vendor -o ../../bookkeeper
	cd cmd/collector && go build -mod=vendor -o ../../collector
	cd cmd/distributor && go build -mod=vendor -o ../../distributor

build-windows:
	cd cmd/bookkeeper && go build -mod=vendor -o ../../bookkeeper.exe
	cd cmd/collector && go build -mod=vendor -o ../../collector.exe
	cd cmd/distributor && go build -mod=vendor -o ../../distributor.exe

demo:
	./scripts/demo.sh

install:
	GO111MODULE=off go get github.com/goware/modvendor

vendor:
	go mod vendor
	modvendor -copy="**/*.c **/*.h **/*.proto" -v

.PHONY: build build-windows demo install vendor
