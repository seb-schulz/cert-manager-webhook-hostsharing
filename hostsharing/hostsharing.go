package hostsharing

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Updater interface {
	Add(key string) error
	Remove(key string) error
	ApiKey() string
}

func removeTxtRecord(updater Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}
		updater.Remove(req.Form.Get("key"))
		log.Printf("TXT Record %#v removed.", req.Form.Get("key"))
	})
}

func addTxtRecord(updater Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		updater.Add(req.Form.Get("key"))
		log.Printf("TXT Record %#v added.", req.Form.Get("key"))
	})
}

func verifyAuthHeader(apiKey, headerValue string) error {
	var token string

	apiKey = strings.TrimSpace(apiKey)

	if len(apiKey) == 0 {
		return fmt.Errorf("apiKey is empty and needs to be configured")
	}

	_, err := fmt.Sscanf(headerValue, "Bearer %s", &token)
	if err != nil {
		return fmt.Errorf("Cannot read token from http header: %v", err)
	}

	if token != apiKey {
		return fmt.Errorf("Tokens are not equal")
	}
	return nil

}

func UpdateHandler(u Updater) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("Receive request.")
		if err := verifyAuthHeader(u.ApiKey(), req.Header.Get("Authorization")); err != nil {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		}
		switch req.Method {
		case http.MethodPost:
			addTxtRecord(u).ServeHTTP(w, req)
		case http.MethodDelete:
			removeTxtRecord(u).ServeHTTP(w, req)
		default:
			http.NotFoundHandler()
		}
	})
}
