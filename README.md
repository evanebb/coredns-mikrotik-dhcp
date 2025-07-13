# mikrotik-dhcp

## Name

*mikrotik-dhcp* - enables automatically resolving hostnames based on DHCP leases from a MikroTik device.

## Description
This plugin will serve A records based on the hostnames and IP addresses found in DHCP leases on a MikroTik device.
It uses the [MikroTik REST API](https://help.mikrotik.com/docs/spaces/ROS/pages/47579162/REST+API) to retrieve the DHCP leases.

Since these DHCP leases only contain a hostname and not a FQDN, requests are handled by stripping the zone from the
requested name, and checking if a DHCP lease with a hostname matching the stripped name exists.  
For example, if this plugin is configured to handle the zone `dhcp.example.org` and a request for
`host1.dhcp.example.org` comes in, it will look for a DHCP lease with the hostname `host1`. If found, it will return an
A record with the corresponding IP address.

## Syntax

~~~ txt
mikrotik-dhcp [ZONES...] {
    url URL
    username USERNAME
    password PASSWORD
    insecure
}
~~~

* **ZONES** the zones this plugin should be authoritative for. Multiple zones can be specified, however they will all
    be served using the same DHCP leases.
* `url` the full **URL** at which the MikroTik REST API of your device can be reached, for example `https://router.example.org`.
* `username` is used to set the **USERNAME** to use to access the MikroTik REST API.
* `password` is used to set the **PASSWORD** corresponding to the username used to access the MikroTik REST API.
* `insecure` whether to verify the TLS certificate chain and host name of the MikroTik device, for example if you are
   using a self-signed certificate.

## Examples

The default recommended configuration, including caching.

~~~ corefile
dhcp.example.org {
    mikrotik-dhcp {
        url https://router.example.org
        username admin
        password foobar
    }
    cache
}

. {
    forward . 8.8.8.8:53 8.8.4.4:53
    cache
}
~~~

Or if you are using a self-signed certificate on your device, and want to disable TLS verification using the `insecure`
option (not recommended):

~~~ corefile
dhcp.example.org {
    mikrotik-dhcp {
        url https://router.example.org
        username admin
        password foobar
        insecure
    }
    cache
}
~~~

## Notes
It is recommended to use this plugin together with the [`cache`](https://coredns.io/plugins/cache/) plugin.
Calls to the MikroTik REST API can add unwanted latency to your DNS queries, and this plugin does not cache responses
internally.
