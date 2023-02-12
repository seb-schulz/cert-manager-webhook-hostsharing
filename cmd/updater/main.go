package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ConfigTmpl struct {
	Head string `yaml:"head"`
}

type Config struct {
	Type     ServeType  `yaml:"type"`
	ApiKeys  []string   `yaml:"api-keys"`
	Template ConfigTmpl `yaml:"template"`
}

type ServeType string

const (
	FastCGIServeType ServeType = "fastcgi"
	HttpServeType    ServeType = "http"
	DefaultTxTRegex  string    = `^_acme-challenge.+IN\s+TXT\s+\"(?P<key>.+)\"\s+;\s+acme-updater`
	DefaultTxTLine   string    = "_acme-challenge.{DOM_HOSTNAME}. IN TXT %v ; acme-updater"
)

const Dummmy string = `{DEFAULT_ZONEFILE}
_acme-challenge.{DOM_HOSTNAME}. IN TXT "1234" ; acme-updater
_acme-challenge.{DOM_HOSTNAME}. IN TXT "5678" ; acme-updater
`

type void struct{}

type AcmeUpdater map[string]void

func loadConfig() Config {
	return Config{HttpServeType, []string{"123"}, ConfigTmpl{"{DEFAULT_ZONEFILE}"}}
}

func (updater AcmeUpdater) parseZoneFile(reader io.Reader) error {
	r := regexp.MustCompile(DefaultTxTRegex)
	idx := r.SubexpIndex("key")

	zoneFile, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	for _, v := range strings.Split(string(zoneFile), "\n") {
		group := r.FindStringSubmatch(v)
		if len(group) > idx {
			updater[group[idx]] = void{}
		}
	}
	return nil
}

func (updater AcmeUpdater) writeZoneFile(cfg Config, w io.Writer) {
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

func removeTxtRecord(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		updater := AcmeUpdater{}
		err = updater.parseZoneFile(strings.NewReader(Dummmy))
		if err != nil {
			log.Fatal("Broken zone file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		delete(updater, req.Form.Get("key"))
		io.WriteString(w, "Remove recored!\n")

		log.Println("Start new zone file")
		updater.writeZoneFile(cfg, os.Stdout)
		log.Println("End new zone file")
	})
}

func addTxtRecord(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		updater := AcmeUpdater{}
		err = updater.parseZoneFile(strings.NewReader(Dummmy))
		if err != nil {
			log.Fatal("Broken zone file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		updater[req.Form.Get("key")] = void{}
		io.WriteString(w, "Add recored!\n")

		log.Println("Start new zone file")
		updater.writeZoneFile(cfg, os.Stdout)
		log.Println("End new zone file")
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
