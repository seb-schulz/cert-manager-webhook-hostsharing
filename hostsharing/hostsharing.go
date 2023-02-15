package hostsharing

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
)

type Updater interface {
	Add(key string) error
	Remove(key string) error
	ApiKey() string
}

func removeTxtRecord(updater Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		updater.Remove(req.Form.Get("key"))
		log.Printf("TXT Record %#v removed.", req.Form.Get("key"))
	})
}

func addTxtRecord(updater Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		updater.Add(req.Form.Get("key"))
		log.Printf("TXT Record %#v added.", req.Form.Get("key"))
	})
}

func verifyAuthHeader(apiKey, token string) error {
	if len(apiKey) == 0 {
		return fmt.Errorf("apiKey is empty and needs to be configured")
	}

	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(token)) != 1 {
		return fmt.Errorf("Tokens are not equal")
	}
	return nil

}

func UpdateHandler(u Updater) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("Receive request.")

		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
		}

		if err := verifyAuthHeader(u.ApiKey(), req.Form.Get("auth")); err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized access: %v", err), http.StatusUnauthorized)
			return
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
