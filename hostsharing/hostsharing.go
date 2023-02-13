package hostsharing

import (
	"log"
	"net/http"
)

type Updater interface {
	Add(key string)
	Remove(key string)
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

func UpdateHandler(u Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("Receive request.")
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
