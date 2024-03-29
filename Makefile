SHELL := /bin/bash
.DEFAULT_GOAL := default
.PHONY: all build

BINARY_NAME=govnocloud
IMAGE_TAG=$(shell git describe --tags --always)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
ORG_PREFIX := loqutus

export TEST_MASTER_HOST := master.govno.cloud
export TEST_MASTER_PORT := 6969
export TEST_NODE_NAME := node0
export TEST_NODE_HOST := node0.govno.cloud
export TEST_NODE_PORT := 6969
export TEST_NODES := node0.govno.cloud:6969,node1.govno.cloud:6969

tidy:
	go mod tidy

get:
	go get -v ./...

build:
	GOARCH=arm64 GOOS=darwin go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-deploy-darwin-arm64 cmd/deploy/main.go
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-deploy-linux-amd64 cmd/deploy/main.go
	GOARCH=arm64 GOOS=darwin go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-client-darwin-arm64 cmd/client/main.go
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-client-linux-amd64 cmd/client/main.go
	GOARCH=arm64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-client-linux-arm64 cmd/client/main.go
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-master-linux-amd64 cmd/master/main.go
	GOARCH=arm64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-master-linux-arm64 cmd/master/main.go
	GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-node-linux-amd64 cmd/node/main.go
	chmod +x bin/*
	docker build -t ${ORG_PREFIX}/${BINARY_NAME}-front:${IMAGE_TAG} -f build/Dockerfile-front .
	docker tag ${ORG_PREFIX}/${BINARY_NAME}-front:${IMAGE_TAG} ${ORG_PREFIX}/${BINARY_NAME}-front:latest
	docker push ${ORG_PREFIX}/${BINARY_NAME}-front:${IMAGE_TAG}
	docker push ${ORG_PREFIX}/${BINARY_NAME}-front:latest

buildclient:
	GOARCH=arm64 GOOS=darwin go build -ldflags "-X main.version=$(GIT_COMMIT)" -o bin/${BINARY_NAME}-client-darwin-arm64 cmd/client/main.go

test:
	go test -timeout 30m -v ./...

deploy:
	bin/govnocloud-deploy-linux-amd64 --master master.govno.cloud --nodes node0.govno.cloud,node1.govno.cloud

logs:
	journalctl _SYSTEMD_INVOCATION_ID=`systemctl show -p InvocationID --value govnocloud-master.service`
	ssh node0.govno.cloud "get_logs.sh"
	ssh node1.govno.cloud "get_logs.sh"

remotetest:
	rsync -avz . master.govno.cloud:~/govnocloud
	ssh master.govno.cloud "cd govnocloud; make ansible get build deploy test logs"

rsync:
	rsync -avz . master.govno.cloud:~/govnocloud

default: get build

