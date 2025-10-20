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
		community.proto

binary-build:
	go build -o community ./cmd/community
	strip community

image-microservice-build: version ?= latest
image-microservice-build:
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile.microservice \
		--tag gdemojkod/stormic-community-microservice:${version} \
		.

image-microservice-push: version ?= latest
image-microservice-push:
	docker push gdemojkod/stormic-community-microservice:${version}

image-migration-build: version ?= latest
image-migration-build:
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile.migration \
		--tag gdemojkod/stormic-community-migration:${version} \
		.

image-migration-push: version ?= latest
image-migration-push:
	docker push gdemojkod/stormic-community-migration:${version}

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
