package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/cert-manager/cert-manager/test/acme/dns"
	dns_mock "github.com/miekg/dns"
	"github.com/seb-schulz/cert-manager-webhook-hostsharing/hostsharing"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

type void struct{}

type dnsServer struct {
	server     *dns_mock.Server
	txtRecords map[string]void
	apiKey     string
	sync.RWMutex
}

func (e *dnsServer) Add(key string) error {
	e.Lock()
	e.txtRecords[key] = void{}
	e.Unlock()
	return nil
}

func (e *dnsServer) Remove(key string) error {
	e.Lock()
	delete(e.txtRecords, key)
	e.Unlock()
	return nil
}

func (e *dnsServer) ApiKey() string {
	return e.apiKey
}

func (e *dnsServer) handleDNSRequest(w dns_mock.ResponseWriter, req *dns_mock.Msg) {
	msg := new(dns_mock.Msg)
	msg.SetReply(req)
	switch req.Opcode {
	case dns_mock.OpcodeQuery:
		for _, q := range msg.Question {
			if err := e.addDNSAnswer(q, msg, req); err != nil {
				msg.SetRcode(req, dns_mock.RcodeServerFailure)
				break
			}
		}
	}
	w.WriteMsg(msg)
}

func (e *dnsServer) addDNSAnswer(q dns_mock.Question, msg *dns_mock.Msg, req *dns_mock.Msg) error {
	switch q.Qtype {
	case dns_mock.TypeA:
		rr, err := dns_mock.NewRR(fmt.Sprintf("%s 5 IN A 127.0.0.1", q.Name))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil

		// TXT records are the only important record for ACME dns-01 challenges
	case dns_mock.TypeTXT:
		e.RLock()
		// record, found := e.txtRecords[q.Name]
		for record := range e.txtRecords {
			rr, err := dns_mock.NewRR(fmt.Sprintf("%s 5 IN TXT %s", q.Name, record))
			if err != nil {
				return err
			}
			msg.Answer = append(msg.Answer, rr)
		}
		e.RUnlock()
		return nil

		// NS and SOA are for authoritative lookups, return obviously invalid data
	case dns_mock.TypeNS:
		rr, err := dns_mock.NewRR(fmt.Sprintf("%s 5 IN NS ns.example-acme-webook.invalid.", q.Name))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil
	case dns_mock.TypeSOA:
		rr, err := dns_mock.NewRR(fmt.Sprintf("%s 5 IN SOA %s 20 5 5 5 5", "ns.example-acme-webook.invalid.", "ns.example-acme-webook.invalid."))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil
	default:
		return fmt.Errorf("unimplemented record type %v", q.Qtype)
	}
}

func newDnsServer(port, apiKey string) *dnsServer {
	e := &dnsServer{txtRecords: map[string]void{}}
	e.server = &dns_mock.Server{
		Addr:    ":" + port,
		Net:     "udp",
		Handler: dns_mock.HandlerFunc(e.handleDNSRequest),
	}

	go func() {
		if err := e.server.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	}()

	return e
}

func TestRunsSuite(t *testing.T) {
	apiKey := "12345"
	dnsPort, _ := rand.Int(rand.Reader, big.NewInt(50000))
	dnsSvr := newDnsServer(dnsPort.String(), apiKey)
	svr := httptest.NewServer(hostsharing.UpdateHandler(dnsSvr))
	defer svr.Close()
	defer dnsSvr.server.Shutdown()

	fixture := dns.NewFixture(&hostsharingDNSSolver{},
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetUseAuthoritative(false),
		dns.SetDNSServer(fmt.Sprintf("127.0.0.1:%v", dnsPort)),
		dns.SetConfig(customConfig{svr.URL, "", apiKey}),
	)
	fixture.RunConformance(t)
}
