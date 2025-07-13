package coredns_mikrotik_dhcp

import (
	"context"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"strings"
)

const pluginName = "mikrotik-dhcp"

type MikroTikDHCPPlugin struct {
	leaseGetter LeaseGetter
	zones       []string
	Next        plugin.Handler
}

func New(g LeaseGetter, zones []string) *MikroTikDHCPPlugin {
	return &MikroTikDHCPPlugin{leaseGetter: g, zones: zones}
}

func (p *MikroTikDHCPPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var err error

	state := request.Request{W: w, Req: r}

	zoneName := plugin.Zones(p.zones).Matches(state.Name())
	if zoneName == "" {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	if state.Type() != dns.Type(dns.TypeA).String() {
		// only A records are supported for now
		return dns.RcodeServerFailure, nil
	}

	answers, err := p.getRecords(ctx, zoneName, state.Name())
	if err != nil || len(answers) == 0 {
		// We should return an NXDOMAIN, but we don't have a SOA record available here
		return dns.RcodeServerFailure, err
	}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true

	a.Answer = answers
	_ = w.WriteMsg(a)

	return dns.RcodeSuccess, nil
}

func (p *MikroTikDHCPPlugin) Name() string {
	return pluginName
}

func (p *MikroTikDHCPPlugin) getRecords(ctx context.Context, zone, name string) ([]dns.RR, error) {
	var rrset []dns.RR

	var q string
	if name != zone {
		q = strings.TrimSuffix(name, "."+zone)
	}

	leases, err := p.leaseGetter.GetBoundLeases(ctx)
	if err != nil {
		return rrset, err
	}

	for _, lease := range leases {
		if !strings.EqualFold(lease.Hostname, q) {
			continue
		}

		rr := new(dns.A)
		rr.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    86400,
		}
		rr.A = lease.Address

		rrset = append(rrset, rr)
	}

	return rrset, nil
}
