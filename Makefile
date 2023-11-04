GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

DOCKER ?= $(shell which docker)
SCP_BIN ?=$(shell which scp)
SSH_BIN ?=$(shell which ssh)
SSH_OPTS ?=

IMAGE_NAME ?= cert-manager-webhook-hostsharing
IMAGE_TAG ?= 2 #latest
-include Makefile.variables

OUT := $(shell pwd)/_out
KUBE_VERSION=1.25.0

$(shell mkdir -p "$(OUT)")
export TEST_ASSET_ETCD=$(CURDIR)/_test/kubebuilder/etcd
export TEST_ASSET_KUBE_APISERVER=$(CURDIR)/_test/kubebuilder/kube-apiserver
export TEST_ASSET_KUBECTL=$(CURDIR)/_test/kubebuilder/kubectl
export TEST_ZONE_NAME=example.com.

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
	$(DOCKER) build -v $(OUT):/workspace/_out:z -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push:
	$(DOCKER) push "$(IMAGE_NAME):$(IMAGE_TAG)" ${REMOTE_REGISTRY}

deploy-hostsharing:
	$(SSH_BIN) $(SSH_OPTS) $(SSH_HOST) killall updater || true
	$(SCP_BIN) $(SSH_OPTS) $(OUT)/updater $(SCP_DEST)

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
            --set image.repository=$(IMAGE_NAME) \
            --set image.tag=$(IMAGE_TAG) \
            cert-manager-webhook-hostsharing deploy/cert-manager-webhook-hostsharing > "$(OUT)/rendered-manifest.yaml"
