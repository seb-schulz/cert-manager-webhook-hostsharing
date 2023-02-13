package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/seb-schulz/cert-manager-webhook-hostsharing/hostsharing"
	"gopkg.in/yaml.v3"
)

type ConfigTmpl struct {
	Head string `yaml:"head"`
}

type Config struct {
	Type     ServeType  `yaml:"type"`
	ZoneFile string     `yaml:"zone-file"`
	ApiKeys  []string   `yaml:"api-keys"`
	Template ConfigTmpl `yaml:"template"`
}

type ServeType string

const (
	FastCGIServeType    ServeType = "fastcgi"
	HttpServeType       ServeType = "http"
	DefaultTxTRegex     string    = `^_acme-challenge.+IN\s+TXT\s+\"(?P<key>.+)\"\s+;\s+acme-updater`
	DefaultTxTLine      string    = "_acme-challenge.{DOM_HOSTNAME}. IN TXT \"%v\" ; acme-updater"
	DefaultTemplateHead string    = "{DEFAULT_ZONEFILE}"
)

type void struct{}

type bindUpdater struct {
	config Config
	keys   map[string]void
}

func loadConfig() Config {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("%v/config.yaml", wd)

	if _, err := os.Stat(path); err != nil {
		panic(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	cfg := Config{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		panic(err)
	}

	if cfg.Template.Head == "" {
		cfg.Template.Head = DefaultTemplateHead
	}

	if cfg.ZoneFile == "" {
		panic("No zone file defined")
	}

	log.Println("Config loaded.")
	return cfg
}

func (updater bindUpdater) parseZoneFile(zoneFile string) error {
	r := regexp.MustCompile(DefaultTxTRegex)
	idx := r.SubexpIndex("key")

	for _, v := range strings.Split(zoneFile, "\n") {
		group := r.FindStringSubmatch(v)
		if len(group) > idx {
			updater.keys[group[idx]] = void{}
		}
	}
	return nil
}

func (updater bindUpdater) writeZoneFile(w io.Writer) {
	_, err := io.WriteString(w, updater.config.Template.Head)
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, "\n")
	if err != nil {
		panic(err)
	}

	for key := range updater.keys {
		_, err := io.WriteString(w, fmt.Sprintf(DefaultTxTLine, key))
		if err != nil {
			panic(err)
		}

		_, err = io.WriteString(w, "\n")
		if err != nil {
			panic(err)
		}
	}
}

func (updater bindUpdater) readZoneFile() (bool, string) {
	if _, err := os.Stat(updater.config.ZoneFile); err != nil {
		return false, ""
	}

	zone, err := os.Open(updater.config.ZoneFile)
	defer zone.Close()
	if err != nil {
		log.Println("Error while opening zone file", err)
		return false, ""
	}

	result, err := io.ReadAll(zone)
	if err != nil {
		log.Println("Error while reading zone file", err)
		return false, ""
	}
	return true, string(result)
}

func (updater bindUpdater) Remove(key string) error {
	if ok, zone := updater.readZoneFile(); ok {
		if err := updater.parseZoneFile(zone); err != nil {
			return err
		}
	}

	delete(updater.keys, key)

	zoneFile, err := os.OpenFile(updater.config.ZoneFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	updater.writeZoneFile(zoneFile)
	return nil
}

func (updater bindUpdater) Add(key string) error {
	if ok, zone := updater.readZoneFile(); ok {
		if err := updater.parseZoneFile(zone); err != nil {
			return err
		}
	}

	updater.keys[key] = void{}
	zoneFile, err := os.OpenFile(updater.config.ZoneFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	updater.writeZoneFile(zoneFile)
	return nil
}

func main() {
	cfg := loadConfig()

	http.Handle("/acme-txt", hostsharing.UpdateHandler(bindUpdater{config: cfg, keys: map[string]void{}}))

	switch cfg.Type {
	case HttpServeType:
		log.Fatal(http.ListenAndServe(":9090", nil))
	}
	log.Println("Hello World")
}
