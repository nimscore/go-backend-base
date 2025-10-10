image-build: version ?= latest
image-build:
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile \
		--tag gdemojkod/stormic-postgres:${version} \
		.

image-push: version ?= latest
image-push:
	docker push gdemojkod/stormic-postgres:${version}

chart-build: version ?= 0.1.0
chart-build:
	helm package chart --version ${version}

chart-install: version ?= 0.1.0
chart-install:
	helm -n postgres install --create-namespace postgres postgres-${version}.tgz
