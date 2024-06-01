#!/usr/bin/env bash
set -e

if [ -z "${VERSION}" ]; then
	VERSION=latest
fi

if [ "${VERSION}" == "latest" ]; then
	versionStr=$(curl https://api.github.com/repos/helm/chart-testing/releases/latest | jq -r '.tag_name')
else
	versionStr=${VERSION}
fi

echo "Install helm with version ${versionStr}"

curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh -v ${versionStr}

rm get_helm.sh
