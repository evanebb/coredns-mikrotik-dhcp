package coredns_mikrotik_dhcp

import (
	"context"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	crequest "github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"net"
	"testing"
)

type fakeLeaseGetter struct{}

func (fakeLeaseGetter) GetBoundLeases(_ context.Context) ([]Lease, error) {
	leases := []Lease{
		{
			Address:  net.ParseIP("192.168.1.10"),
			Hostname: "host1",
		},
		{
			Address:  net.ParseIP("192.168.1.11"),
			Hostname: "host2",
		},
		{
			Address:  net.ParseIP("192.168.1.12"),
			Hostname: "host2",
		},
	}

	return leases, nil
}

func TestMikrotikDHCP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := New(fakeLeaseGetter{}, []string{"example.org."})

	p.Next = test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		// Just always return the same answer
		state := crequest.Request{W: w, Req: r}

		rr := new(dns.A)
		rr.Hdr = dns.RR_Header{
			Name:   state.Name(),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    86400,
		}
		rr.A = net.ParseIP("10.0.0.1")

		a := new(dns.Msg)
		a.SetReply(r)
		a.Authoritative = true
		a.Answer = []dns.RR{rr}

		_ = w.WriteMsg(a)
		return dns.RcodeSuccess, nil
	})

	tests := []struct {
		desc           string
		qname          string
		qtype          uint16
		expectedRcode  int
		expectedAnswer []string
	}{
		{
			desc:          "Error if non-A record is requested",
			qname:         "cname.example.org",
			qtype:         dns.TypeCNAME,
			expectedRcode: dns.RcodeServerFailure,
		},
		{
			desc:           "Chain to next handler if non-handled domain is requested",
			qname:          "host.unknown.local",
			qtype:          dns.TypeA,
			expectedRcode:  dns.RcodeSuccess,
			expectedAnswer: []string{"host.unknown.local.\t86400\tIN\tA\t10.0.0.1"},
		},
		{
			desc:           "Return A record if requested host is found in DHCP leases",
			qname:          "host1.example.org",
			qtype:          dns.TypeA,
			expectedRcode:  dns.RcodeSuccess,
			expectedAnswer: []string{"host1.example.org.\t86400\tIN\tA\t192.168.1.10"},
		},
		{
			desc:           "Return one A record if duplicate hostnames found in DHCP leases",
			qname:          "host2.example.org",
			qtype:          dns.TypeA,
			expectedRcode:  dns.RcodeSuccess,
			expectedAnswer: []string{"host2.example.org.\t86400\tIN\tA\t192.168.1.11"},
		},
		{
			desc:          "Error if requested host is not found in DHCP leases",
			qname:         "unknown.example.org",
			qtype:         dns.TypeA,
			expectedRcode: dns.RcodeServerFailure,
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.desc, func(t *testing.T) {
			req := new(dns.Msg)
			req.SetQuestion(dns.Fqdn(testcase.qname), testcase.qtype)

			rec := dnstest.NewRecorder(&test.ResponseWriter{})
			code, _ := p.ServeDNS(ctx, rec, req)

			if code != testcase.expectedRcode {
				t.Fatalf("expected status code %d, got %d", testcase.expectedRcode, code)
			}

			var answer []dns.RR
			// Avoid nil pointer dereference
			if rec.Msg != nil {
				answer = rec.Msg.Answer
			}

			if len(testcase.expectedAnswer) != len(answer) {
				t.Fatalf("expected answer length of %v, got %v", len(testcase.expectedAnswer), len(answer))
			}

			for i, actualAnswerRecord := range answer {
				if actualAnswerRecord.String() != testcase.expectedAnswer[i] {
					t.Errorf("unexpected answer, got %v, expected %v", actualAnswerRecord, testcase.expectedAnswer[i])
				}
			}
		})
	}
}
