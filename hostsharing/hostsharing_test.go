package hostsharing

import (
	"strings"
	"testing"
	"unicode"
)

func TestVerifyAuthHeader(t *testing.T) {

	if err := verifyAuthHeader("1234", "1234"); err != nil {
		t.Error("Cannot verify token")
	}

	if err := verifyAuthHeader("123a", "123b"); err == nil {
		t.Error("Cannot verify token")
	}

	if err := verifyAuthHeader("123a", ""); err == nil {
		t.Error("Empty token is accepted")
	}
}

func FuzzVerifyAuthHeader(f *testing.F) {
	f.Add("123")
	f.Add("abc")

	f.Fuzz(func(t *testing.T, a string) {
		a = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, a)

		err := verifyAuthHeader(a, a)
		if err != nil && len(a) == 0 {
			return
		}
		if err != nil {
			t.Fatalf("Cannot verify %#v due to %v", a, err)
		}
	})
}
