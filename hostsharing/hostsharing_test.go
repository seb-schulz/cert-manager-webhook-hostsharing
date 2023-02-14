package hostsharing

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func TestVerifyAuthHeader(t *testing.T) {

	if err := verifyAuthHeader("1234", "Bearer 1234"); err != nil {
		t.Error("Cannot verify token")
	}

	if err := verifyAuthHeader("123a", "Bearer 123b"); err == nil {
		t.Error("Cannot verify token")
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

		err := verifyAuthHeader(a, fmt.Sprintf("Bearer %v", a))
		if err != nil && len(a) == 0 {
			return
		}
		if err != nil {
			t.Fatalf("Cannot verify %#v due to %v", a, err)
		}
	})
}
