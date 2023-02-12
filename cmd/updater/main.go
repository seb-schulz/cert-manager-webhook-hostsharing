package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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

type acmeUpdater map[string]void

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

func (updater acmeUpdater) parseZoneFile(zoneFile string) error {
	r := regexp.MustCompile(DefaultTxTRegex)
	idx := r.SubexpIndex("key")

	for _, v := range strings.Split(zoneFile, "\n") {
		group := r.FindStringSubmatch(v)
		if len(group) > idx {
			updater[group[idx]] = void{}
		}
	}
	return nil
}

func (updater acmeUpdater) writeZoneFile(cfg Config, w io.Writer) {
	_, err := io.WriteString(w, cfg.Template.Head)
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, "\n")
	if err != nil {
		panic(err)
	}

	for key := range updater {
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

func readZoneFile(cfg Config) (bool, string) {
	if _, err := os.Stat(cfg.ZoneFile); err != nil {
		return false, ""
	}

	zone, err := os.Open(cfg.ZoneFile)
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

func removeTxtRecord(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		updater := acmeUpdater{}

		if ok, zone := readZoneFile(cfg); ok {
			err = updater.parseZoneFile(zone)
			if err != nil {
				log.Fatal("Broken zone file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		delete(updater, req.Form.Get("key"))
		io.WriteString(w, "Remove recored!\n")

		zoneFile, err := os.OpenFile(cfg.ZoneFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln("Cannot write zone file: ", zoneFile)
		}
		updater.writeZoneFile(cfg, zoneFile)
	})
}

func addTxtRecord(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		updater := acmeUpdater{}

		if ok, zone := readZoneFile(cfg); ok {
			err = updater.parseZoneFile(zone)
			if err != nil {
				log.Fatal("Broken zone file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		updater[req.Form.Get("key")] = void{}
		io.WriteString(w, "Add recored!\n")

		zoneFile, err := os.OpenFile(cfg.ZoneFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln("Cannot write zone file: ", zoneFile)
		}
		updater.writeZoneFile(cfg, zoneFile)
	})
}

func dispatchRequest(cfg Config, add, remove http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			add.ServeHTTP(w, req)
		case http.MethodDelete:
			remove.ServeHTTP(w, req)
		default:
			http.NotFoundHandler()
		}
	})
}

func main() {
	config := loadConfig()

	http.Handle("/acme-txt", dispatchRequest(config, addTxtRecord(config), removeTxtRecord(config)))

	switch config.Type {
	case HttpServeType:
		log.Fatal(http.ListenAndServe(":9090", nil))
	}
	log.Println("Hello World")
}
