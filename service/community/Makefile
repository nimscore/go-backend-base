env-setup:
	curl -sSf https://atlasgo.sh | sh

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

ent-add-orm: name ?= User
ent-add-orm:
	go run -mod=mod entgo.io/ent/cmd/ent new ${name} --target ./internal/ent/schema

ent-generate:
	go generate ./internal/ent

database-create-migration: name ?= initial
database-create-migration:
	atlas migrate diff ${name} \
		--dir "file://migrations" \
		--to "ent://internal/ent/schema" \
		--dev-url "postgres://postgres:@127.0.0.1:5432/postgres?sslmode=disable"

database-apply-migrations:
	atlas migrate apply \
		--dir "file://migrations" \
		--url "postgres://postgres:@127.0.0.1:5432/postgres?sslmode=disable" --dry-run

graphql-generate:
	go get github.com/99designs/gqlgen@latest
	go generate ./...
