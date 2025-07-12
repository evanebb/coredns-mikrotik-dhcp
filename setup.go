package coredns_mikrotik_dhcp

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register("mikrotik-dhcp", setup)
}

func setup(c *caddy.Controller) error {
	var origins []string
	var baseURL string
	var username string
	var password string

	var opts []MikroTikAPILeaseGetterOption

	for c.Next() {
		origins = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)

		for c.NextBlock() {
			switch c.Val() {
			case "url":
				v := c.RemainingArgs()
				if len(v) != 1 {
					return plugin.Error(pluginName, c.Err("missing URL"))
				}
				baseURL = v[0]
			case "username":
				v := c.RemainingArgs()
				if len(v) != 1 {
					return plugin.Error(pluginName, c.Err("missing username"))
				}
				username = v[0]
			case "password":
				v := c.RemainingArgs()
				if len(v) != 1 {
					return plugin.Error(pluginName, c.Err("missing password"))
				}
				password = v[0]
			case "insecure":
				v := c.RemainingArgs()
				if len(v) != 0 {
					return plugin.Error(pluginName, c.Err("unnecessary value specified for insecure option"))
				}
				opts = append(opts, WithInsecureSkipVerify())
			}
		}
	}

	if baseURL == "" || username == "" || password == "" {
		return plugin.Error(pluginName, c.Err("missing credentials"))
	}

	leaseGetter := NewMikroTikAPILeaseGetter(baseURL, username, password, opts...)

	handler := New(leaseGetter, origins)

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		handler.Next = next
		return handler
	})

	return nil
}
