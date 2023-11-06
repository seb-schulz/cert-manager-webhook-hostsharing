GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

BUILDAH ?= $(shell which buildah)
SCP_BIN ?=$(shell which scp)
SSH_BIN ?=$(shell which ssh)
SSH_OPTS ?=

VERSION ?= $(shell git describe --tags --abbrev --always)

IMAGE_NAME ?= ghcr.io/seb-schulz/cert-manager-webhook-hostsharing
IMAGE_TAG ?= $(patsubst v%,%,$(VERSION))

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
# export IMAGE_TAG

debug:
	echo $(VERSION) $(IMAGE_TAG)

test: test-webhook test-all

test-webhook: _test/kubebuilder
	$(GO) test -v ./cmd/webhook/

test-all:
	$(GO) test -v ./cmd/updater/
	$(GO) test -v ./hostsharing/...

_test/kubebuilder:
	curl -fsSL https://go.kubebuilder.io/test-tools/$(KUBE_VERSION)/$(OS)/$(ARCH) -o kubebuilder-tools.tar.gz
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

deploy-hostsharing:
	$(SSH_BIN) $(SSH_OPTS) $(SSH_HOST) killall updater || true
	$(SCP_BIN) $(SSH_OPTS) $(OUT)/updater $(SCP_DEST)

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
            --set image.repository=$(IMAGE_NAME) \
            --set image.tag=$(IMAGE_TAG) \
            cert-manager-webhook-hostsharing deploy/cert-manager-webhook-hostsharing > "$(OUT)/rendered-manifest.yaml"
