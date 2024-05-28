GO ?= $(shell which go)

BUILDAH ?= $(shell which buildah)
DOCKER_BIN ?= $(shell which podman 2> /dev/null)
HELM ?= $(shell which helm)
SCP_BIN ?=$(shell which scp)
SSH_BIN ?=$(shell which ssh)
SSH_OPTS ?=

VERSION ?= $(shell git describe --tags --abbrev --always)

IMAGE_NAME ?= ghcr.io/seb-schulz/cert-manager-webhook-hostsharing
IMAGE_TAG ?= $(patsubst v%,%,$(VERSION))

RENOVATE_REPOSITORIES ?= seb-schulz/cert-manager-webhook-hostsharing
WEBPAGE_URL=https://seb-schulz.github.io/cert-manager-webhook-hostsharing/
GIT_REMOTE_URL ?= $(shell git remote get-url origin)

-include Makefile.variables

OUT := $(shell pwd)/_out
KUBE_VERSION?=1.28.3

$(shell mkdir -p "$(OUT)")
TEST_ASSET_ETCD=$(CURDIR)/_test/kubebuilder/etcd
TEST_ASSET_KUBE_APISERVER=$(CURDIR)/_test/kubebuilder/kube-apiserver
TEST_ASSET_KUBECTL=$(CURDIR)/_test/kubebuilder/kubectl
TEST_ZONE_NAME=example.com.

export

test: test-webhook test-all

test-webhook: _test/kubebuilder
	$(GO) test -v ./cmd/webhook/

test-all:
	$(GO) test -v ./cmd/updater/
	$(GO) test -v ./hostsharing/...

_test/kubebuilder:
	curl -fsSL https://go.kubebuilder.io/test-tools/$(KUBE_VERSION)/$(shell $(GO) env GOOS)/$(shell $(GO) env GOARCH) -o kubebuilder-tools.tar.gz
	mkdir -p _test/kubebuilder
	tar -xvf kubebuilder-tools.tar.gz
	mv kubebuilder/bin/* _test/kubebuilder/
	rm kubebuilder-tools.tar.gz
	rm -R kubebuilder

clean: clean-kubebuilder clean-out

clean-kubebuilder:
	rm -Rf _test/kubebuilder

clean-out:
	rm -Rf _out && mkdir -p _out

build:
	$(BUILDAH) bud -v $(OUT):/workspace/_out:z --squash -t $(IMAGE_NAME):$(IMAGE_TAG)

push:
	$(BUILDAH) push $(IMAGE_NAME):$(IMAGE_TAG)

release:
	./scripts/$@.sh

lint:
	$(HELM) lint deploy/cert-manager-webhook-hostsharing/

deploy-hostsharing:
	$(SSH_BIN) $(SSH_OPTS) $(SSH_HOST) killall updater || true
	$(SCP_BIN) $(SSH_OPTS) $(OUT)/updater $(SCP_DEST)

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	$(HELM) template \
            --set image.repository=$(IMAGE_NAME) \
            --set image.tag=$(IMAGE_TAG) \
            cert-manager-webhook-hostsharing deploy/cert-manager-webhook-hostsharing > "$(OUT)/rendered-manifest.yaml"

.PHONY: check-updates
check-updates:
	./scripts/$@.sh
