package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseZoneFile(t *testing.T) {
	tests := map[string]acmeUpdater{
		"Noop": acmeUpdater{},
		`{DEFAULT_ZONEFILE}
_acme-challenge.{DOM_HOSTNAME}. IN TXT "1234" ; acme-updater
_acme-challenge.{DOM_HOSTNAME}. IN TXT "5678" ; acme-updater`: acmeUpdater{"1234": void{}, "5678": void{}},
		`{DEFAULT_ZONEFILE}
_acme-challenge.{DOM_HOSTNAME}. IN TXT "1234" ; acme-updater
_acme-challenge.{DOM_HOSTNAME}. IN TXT "5678"`: acmeUpdater{"1234": void{}},
	}

	for input, expected := range tests {
		got := acmeUpdater{}
		if err := got.parseZoneFile(input); err != nil {
			t.Errorf("Failed due to %v", err)
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Expected %v instead of %v", expected, got)
		}
	}
}

func TestWriteZoneFile(t *testing.T) {
	cfg := Config{HttpServeType, "pri.example.com", []string{}, ConfigTmpl{"{DEFAULT_ZONEFILE}"}}

	expected := []string{
		"{DEFAULT_ZONEFILE}\n",
		"{DEFAULT_ZONEFILE}\n_acme-challenge.{DOM_HOSTNAME}. IN TXT \"123\" ; acme-updater\n",
	}

	for idx, u := range []acmeUpdater{acmeUpdater{}, acmeUpdater{"123": void{}}} {
		b := new(bytes.Buffer)
		u.writeZoneFile(cfg, b)
		// if err := got.parseZoneFile(strings.NewReader(input)); err != nil {
		// 	t.Errorf("Failed due to %v", err)
		// }

		if b.String() != expected[idx] {
			t.Errorf("Expected %#v instead of %#v", expected, b.String())
		}
	}
}
