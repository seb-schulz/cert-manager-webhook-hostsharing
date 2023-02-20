package main

import (
	"encoding/json"
	"fmt"
	"os"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/seb-schulz/cert-manager-webhook-hostsharing/hostsharing"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName,
		&hostsharingDNSSolver{},
	)
}

type hostsharingDNSSolver struct {
}

type customConfig struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	BaseUrl string `json:"baseUrl"`
	ApiKey  string `json:"apiKey"`
}

func (c *hostsharingDNSSolver) Name() string {
	return "hostsharing"
}

func (c *hostsharingDNSSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	if err := hostsharing.AddTxtRecord(cfg.BaseUrl, cfg.ApiKey, ch.Key); err != nil {
		fmt.Println("Failed request change: ", err)
	}

	return nil
}

func (c *hostsharingDNSSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	if err := hostsharing.RemoveTxtRecord(cfg.BaseUrl, cfg.ApiKey, ch.Key); err != nil {
		fmt.Println("Failed request change: ", err)
	}

	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *hostsharingDNSSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	return nil
}

func loadConfig(cfgJSON *extapi.JSON) (customConfig, error) {
	cfg := customConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
