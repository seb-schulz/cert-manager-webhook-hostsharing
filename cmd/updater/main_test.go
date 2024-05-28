package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func newBindUpdater(keys []string) bindUpdater {
	r := bindUpdater{config: Config{}, keys: map[string]void{}}

	for _, v := range keys {
		r.keys[v] = void{}
	}

	return r
}

func TestParseZoneFile(t *testing.T) {
	tests := map[string]map[string]void{
		"Noop": {},
		`{DEFAULT_ZONEFILE}
_acme-challenge.{DOM_HOSTNAME}. IN TXT "1234" ; acme-updater
_acme-challenge.{DOM_HOSTNAME}. IN TXT "5678" ; acme-updater`: {"1234": {}, "5678": {}},
		`{DEFAULT_ZONEFILE}
_acme-challenge.{DOM_HOSTNAME}. IN TXT "1234" ; acme-updater
_acme-challenge.{DOM_HOSTNAME}. IN TXT "5678"`: {"1234": void{}},
	}

	for input, expected := range tests {
		obj := newBindUpdater([]string{})
		if err := obj.parseZoneFile(input); err != nil {
			t.Errorf("Failed due to %v", err)
		}

		if !reflect.DeepEqual(obj.keys, expected) {
			t.Errorf("Expected %v instead of %v", expected, obj.keys)
		}
	}
}

func TestWriteZoneFile(t *testing.T) {
	cfg := Config{"pri.example.com", "", ConfigTmpl{"{DEFAULT_ZONEFILE}"}}

	expected := []string{
		"{DEFAULT_ZONEFILE}\n",
		"{DEFAULT_ZONEFILE}\n_acme-challenge.{DOM_HOSTNAME}. IN TXT \"123\" ; acme-updater\n",
	}

	for idx, keys := range [][]string{{}, {"123"}} {
		b := new(bytes.Buffer)
		obj := newBindUpdater(keys)
		obj.config = cfg
		obj.writeZoneFile(b)
		// if err := got.parseZoneFile(strings.NewReader(input)); err != nil {
		// 	t.Errorf("Failed due to %v", err)
		// }

		got := b.String()

		if strings.Contains(expected[idx], got) {
			t.Errorf("Expected %#v instead of %#v", expected[idx], b.String())
		}
	}
}
