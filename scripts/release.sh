#!/bin/bash

set -xeu

rootdir=$(realpath $(dirname $0)/..)
uploader_filename=updater-${VERSION}-amd64

pushd ${rootdir}

git_email=$(git config user.email)
git_name=$(git config user.name)

gh_page_path=$(mktemp -d gh-pages.XXXX)
upload_dir=$(mktemp -d upload.XXXX)

function finish {
  rm -rf ${gh_page_path} ${upload_dir}
}
trap finish EXIT ERR

git clone -b gh-pages ${GIT_REMOTE_URL} ${gh_page_path}

pushd ${gh_page_path}
git config user.email "${git_email}"
git config user.name "${git_name}"

helm package ${rootdir}/deploy/cert-manager-webhook-hostsharing
git add cert-manager-webhook-hostsharing-${VERSION}.tgz

helm repo index . --url ${WEBPAGE_URL}

git commit -a -m "Release version ${VERSION}"
git remote -v
git push origin gh-pages
popd

cp _out/updater ${upload_dir}/${uploader_filename}
echo ${VERSION} > ${upload_dir}/version.txt

(cd ${upload_dir} && sha256sum ${uploader_filename} > ${uploader_filename}.sha256sum.txt)

gh release create ${VERSION} \
    --draft \
    --generate-notes \
    --latest \
    ${upload_dir}/*
