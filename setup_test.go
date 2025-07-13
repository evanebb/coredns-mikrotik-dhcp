package coredns_mikrotik_dhcp

import (
	"github.com/coredns/caddy"
	"testing"
)

func TestSetupMikrotikDhcp(t *testing.T) {
	tests := []struct {
		desc          string
		body          string
		expectedError bool
	}{
		{
			`Invalid URL`,
			`mikrotik-dhcp {
	url
}`,
			true,
		},
		{
			`Invalid username`,
			`mikrotik-dhcp {
	username
}`,
			true,
		},
		{
			`Invalid password`,
			`mikrotik-dhcp {
	password
}`,
			true,
		},
		{
			`Unnecessary value for insecure option`,
			`mikrotik-dhcp {
	insecure foobar
}`,
			true,
		},
		{
			`Missing credentials`,
			`mikrotik-dhcp`,
			true,
		},
		{
			`Missing credentials: URL`,
			`mikrotik-dhcp {
	username api
	password secure-password
}`,
			true,
		},
		{
			`Missing credentials: username`,
			`mikrotik-dhcp {
	url http://localhost
	password secure-password
}`,
			true,
		},
		{
			`Missing credentials: password`,
			`mikrotik-dhcp {
	url http://localhost
	username api
}`,
			true,
		},
		{
			`Complete credentials`,
			`mikrotik-dhcp {
	url http://localhost
	username api
	password secure-password
}`,
			false,
		},
		{
			`Complete credentials with insecure option`,
			`mikrotik-dhcp {
	url http://localhost
	username api
	password secure-password
	insecure
}`,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			c := caddy.NewTestController("dns", test.body)
			err := setup(c)
			t.Logf("error: %v", err)
			if (err == nil) == test.expectedError {
				t.Errorf("unexpected errors: %v", err)
			}
		})
	}
}
