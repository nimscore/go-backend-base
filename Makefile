env:
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go get -tool google.golang.org/protobuf/cmd/protoc-gen-go
	go get -tool google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install tool

generate-proto:
	protoc \
		--proto_path=proto \
		--proto_path=$(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway) \
		--proto_path=$(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)/third_party/googleapis \
		--go_out=./internal/proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=./internal/proto \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./internal/proto \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=./api/swagger \
		--openapiv2_opt=allow_merge=true \
		--openapiv2_opt=merge_file_name=api \
		authorization.proto \
		community.proto \
		entity.proto \
		post.proto \
		user.proto

binary-build:
	go build -o backend ./cmd/backend
	strip backend

image-build: version ?= latest
image-build:
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile.microservice \
		--tag stormic/backend:${version} \
		.

image-push: version ?= latest
image-push:
	docker push stormic/backend:${version}

chart-build: version ?= 0.1.0
chart-build:
	helm package chart --version ${version}

chart-install: version ?= 0.1.0
chart-install:
	helm -n community install --create-namespace community community-${version}.tgz

database-create-migration: name ?= initial
database-create-migration:
	migrate create -ext sql -dir migration -seq ${name}

database-apply-migrations:
	migrate -database 'postgres://postgres:postgres@127.0.0.1:5432?sslmode=disable' -path migration up

database-delete-migrations:
	migrate -database 'postgres://postgres:postgres@127.0.0.1:5432?sslmode=disable' -path migration down
