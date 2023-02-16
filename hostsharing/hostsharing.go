package hostsharing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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

func generateAuthToken(apiKey string, acmeKey string, salt int) []byte {
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(acmeKey))
	mac.Write([]byte(strconv.Itoa(salt)))
	return mac.Sum(nil)
}

func UpdateHandler(u Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("Receive request.")

		if req.Method != http.MethodPost {
			http.NotFoundHandler()
			return
		}

		err := req.ParseForm()
		if err != nil {
			log.Fatalf("Cannot parse request: %v\n", err)
			return
		}

		acmeKey := req.Form.Get("key")
		if len(acmeKey) == 0 {
			http.Error(w, "Empty key is invalid", http.StatusBadRequest)
			return
		}

		salt, err := strconv.Atoi(req.Form.Get("salt"))
		if err != nil {
			http.Error(w, "Invalid salt", http.StatusBadRequest)
			return
		}

		auth, err := hex.DecodeString(req.Form.Get("auth"))
		if err != nil {
			http.Error(w, "Cannot decode auth param", http.StatusBadRequest)
			return
		}

		if len(auth) == 0 {
			http.Error(w, "Unauthorized access: missing auth param", http.StatusUnauthorized)
			return
		}

		if !hmac.Equal(auth, generateAuthToken(u.ApiKey(), acmeKey, salt)) {
			http.Error(w, fmt.Sprintf("Unauthorized access!"), http.StatusUnauthorized)
			return
		}

		switch req.Form.Get("action") {
		case "add":
			u.Add(acmeKey)
		case "remove":
			u.Remove(acmeKey)
		default:
			http.Error(w, fmt.Sprintf("Invalid or missing action field"), http.StatusBadRequest)
			return
		}
	})
}

func prepareHeader(data url.Values, apiKey, acmeKey string) {
	salt := rand.Int()
	data.Add("auth", hex.EncodeToString(generateAuthToken(apiKey, acmeKey, salt)))
	data.Add("salt", strconv.Itoa(salt))
	data.Add("key", acmeKey)
}

func AddTxtRecord(baseUrl, apiKey, acmeKey string) error {
	data := url.Values{}
	data.Add("action", "add")
	prepareHeader(data, apiKey, acmeKey)

	resp, err := http.PostForm(baseUrl, data)
	if err != nil {
		fmt.Println("Failed to create record: ", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Cannot create TXT record")
	}
	return nil
}

func RemoveTxtRecord(baseUrl, apiKey, acmeKey string) error {
	data := url.Values{}
	data.Add("action", "remove")
	prepareHeader(data, apiKey, acmeKey)

	resp, err := http.PostForm(baseUrl, data)
	if err != nil {
		return fmt.Errorf("Failed to remove record: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Cannot remove TXT record")
	}
	return nil
}
