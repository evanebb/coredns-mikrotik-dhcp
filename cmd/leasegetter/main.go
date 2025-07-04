package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	coredns_mikrotik_dhcp "github.com/evanebb/coredns-mikrotik-dhcp"
	"log"
	"os"
	"text/tabwriter"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

var (
	baseURL  string
	username string
	password string
	insecure bool
)

func run(ctx context.Context) error {
	flag.StringVar(&baseURL, "base-url", "", "base URL of MikroTik API")
	flag.StringVar(&username, "username", "", "username for MikroTik API")
	flag.StringVar(&password, "password", "", "password for MikroTik API")
	flag.BoolVar(&insecure, "insecure", false, "whether to verify TLS certificates for the MikroTik API")

	flag.Parse()

	if baseURL == "" || username == "" || password == "" {
		return errors.New("specify base URL, username and password")
	}

	var opts []coredns_mikrotik_dhcp.MikroTikAPILeaseGetterOption
	if insecure {
		opts = append(opts, coredns_mikrotik_dhcp.WithInsecureSkipVerify())
	}

	g := coredns_mikrotik_dhcp.NewMikroTikAPILeaseGetter(baseURL, username, password, opts...)

	var leases []coredns_mikrotik_dhcp.Lease
	var err error

	leases, err = g.GetBoundLeases(ctx)

	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
	_, _ = fmt.Fprintln(w, "ADDRESS\tHOSTNAME")
	for _, lease := range leases {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", lease.Address, lease.Hostname)
	}
	_ = w.Flush()

	return nil
}
