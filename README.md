<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>

# ACME webhook for Hostsharing

This solver can be used when you want to use cert-manager with [Hostsharing e.G.](https://www.hostsharing.net/).

## Requirements

* [buildah](https://buildah.io/) for building container and binaries
* [helm](https://helm.sh/)
* [kubernetes](https://kubernetes.io/) or [k0s](https://k0sproject.io/) which is more lightweight
* [cert-manager](https://cert-manager.io/)

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

#### Using public helm chart

```bash
helm repo add cert-manager-webhook-hostsharing https://seb-schulz.github.io/cert-manager-webhook-hostsharing
# Replace the groupName value with your desired domain
helm install --namespace cert-manager cert-manager-webhook-hostsharing cert-manager-webhook-hostsharing/cert-manager-webhook-hostsharing --set groupName=acme.yourdomain.tld
```

#### From local checkout

```bash
helm install --namespace cert-manager cert-manager-webhook-hostsharing deploy/cert-manager-webhook-hostsharing
```
**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run

```bash
helm uninstall --namespace cert-manager cert-manager-webhook-hostsharing
```

TODO: How to generate api token

### On hostsharing

Setup a domain with [HSAdmin](https://admin.hostsharing.net/). It is recommeded to setup a user as well. Please consider the [documentation](https://www.hostsharing.net/doc/) for more information. In this README we are going to use the user `xyz00-acme` and the domain `acme.example.com` as an example.

1. Download **updater** component from [latest release page](https://github.com/seb-schulz/cert-manager-webhook-hostsharing/releases/latest)
2. Move **updater** component to `~/doms/acme.example.com/fastcgi-ssl/`
3. Make **updater** executable
4. Run `updater -config > config.yaml` to generate config file
5. Generate an API key (e.x. `openssl rand -hex 32`) and update config file accordingly

The following shell script does all steps except generating an API key.

```shell
domain=acme.example.com
url=https://github.com/seb-schulz/cert-manager-webhook-hostsharing/releases/latest/download
ver=$(curl -L $url/version.txt)
curl -LO "$url/updater-$ver-amd64"
curl -LO "$url/updater-$ver-amd64.sha256sum.txt"
sha256sum -c updater-$ver-amd64.sha256sum.txt && rm updater-$ver-amd64.sha256sum.txt
chmod +x updater-$ver-amd64
echo mv updater-$ver-amd64 ~/doms/$domain/fastcgi-ssl/updater
~/doms/$domain/fastcgi-ssl/updater -config > ~/doms/$domain/fastcgi-ssl/config.yaml
```

The config file should look similar like

```yaml
zone-file: "/home/pacs/xyz00/users/acme/doms/acme.example.com/etc/pri.acme.example.com"
api-key: "random string"
template:
    head: '{DEFAULT_ZONEFILE}'
```

### Cluster Issuer

You are going to need an *Issuer* or *ClusterIssuer* on your kubernetes cluster to get all those pieces running. This readme can only provide an example. For more details, please consider the [documentation about webhooks](https://cert-manager.io/docs/configuration/acme/dns01/webhook/) of the cert-manager project.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-staging
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    solvers:
    - dns01:
        cnameStrategy: Follow
        webhook:
          config:
            apiKey: "random string"
            baseUrl: https://acme.example.com/fastcgi-bin/updater
          groupName: acme.example.com
          solverName: hostsharing
```

## How to...

### Use *let's encrypt* certificates within an intranet

TBD

## Development

You can build your own binaries with `make build` and push the container to your private registry with `make push IMAGE_NAME=registry.example.com/cert-manager-webhook-hostsharing`.

All variables of the makefile, you can overwrite by creating a `Makefile.variables` file.

### Running the test suite

You can run the test suite with:

```bash
$ make test
```
