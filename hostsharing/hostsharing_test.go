package hostsharing

import (
	"net/http/httptest"
	"testing"
)

type mock struct {
	key *string
}

func (u mock) Add(key string) error {
	u.key = &key
	return nil
}
func (u mock) Remove(key string) error {
	u.key = &key
	return nil
}
func (u mock) ApiKey() string { return "abc" }

func TestAddTxtRecord(t *testing.T) {
	mock := mock{}
	svr := httptest.NewServer(UpdateHandler(mock))

	if err := AddTxtRecord(svr.URL, "abc", ""); err == nil {
		t.Errorf("Accepted empty acme key")
	}

	err := AddTxtRecord(svr.URL, "invalid", "1234")
	if err == nil {
		t.Errorf("Accepted invalid api key")
	}

	err = AddTxtRecord(svr.URL, "abc", "1234")
	if err != nil {
		t.Errorf("Faild to add txt record: %v", err)
	}

	if mock.key != nil && *mock.key != "1234" {
		t.Errorf("Key was not set correctly")
	}

	defer svr.Close()
}
