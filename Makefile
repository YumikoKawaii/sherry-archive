PROTOC_VERSION = 30.2
PROTOC_ZIP = protoc-$(PROTOC_VERSION)-osx-x86_64.zip
BUF_VERSION=1.53.0
BUF_BINARY_NAME=buf

.PHONY: install-protoc install-protoc-go install-buf

install-protoc:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)
	sudo unzip -o $(PROTOC_ZIP) -d /usr/local bin/protoc
	sudo unzip -o $(PROTOC_ZIP) -d /usr/local 'include/*'
	rm -f $(PROTOC_ZIP)

install-protoc-go:
	go install github.com/golang/protobuf/protoc-gen-go@v1.5.4
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/envoyproxy/protoc-gen-validate@v1.2.1
	go install github.com/gogo/protobuf/protoc-gen-gofast@v1.3.2
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0

install-buf:
	sudo curl -sSL "https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/$(BUF_BINARY_NAME)-$(shell uname -s)-$(shell uname -m)"  -o "/usr/local/bin/$(BUF_BINARY_NAME)" && sudo chmod +x "/usr/local/bin/$(BUF_BINARY_NAME)"

.PHONY: start-infra stop-infra

start-infra:
	docker compose --env-file .env.docker -f infrastructures/archive.docker-compose.yaml up -d

stop-infra:
	docker compose --env-file .env.docker -f infrastructures/archive.docker-compose.yaml down

.PHONY: generate-proto

generate-proto:
	buf build && buf generate ./protobuf