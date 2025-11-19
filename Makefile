#!/bin/bash
export NOW=$(shell date +"%Y/%m/%d %T")
export REPO_NAME=reconciliation-app
# version based tags
IMG_TAG ?= ${shell git rev-parse --short HEAD}

swag:
	@swag init --parseDependency --parseInternal --parseDepth 2 -g cmd/http/main.go 2>&1 | grep -v "warning:" || true

build: swag
	@echo "${NOW} == Building HTTP Server"
	@go build -o ./bin/${REPO_NAME}-http cmd/http/main.go 

run-http: build 
	@./bin/${REPO_NAME}-http

clean-mod-cache:
	@go clean -cache -modcache -i -r

test:
	@go test github.com/elkoshar/reconciliation-app/...

test-debug:
	@go test ./... -v | grep FAIL
