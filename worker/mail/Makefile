binary-build:
	go build -o mail ./cmd/mail
	strip mail

image-worker-build: version ?= latest
image-worker-build:
	docker buildx build \
		--platform linux/amd64 \
		--file Dockerfile \
		--tag gdemojkod/stormic-mail-worker:${version} \
		.

image-worker-push: version ?= latest
image-worker-push:
	docker push gdemojkod/stormic-mail-worker:${version}

chart-build: version ?= 0.1.0
chart-build:
	helm package chart --version ${version}

chart-install: version ?= 0.1.0
chart-install:
	helm -n mail install --create-namespace mail mail-${version}.tgz
