# Sending DNS logs to Splunk

Ties this platform back to the [Splunk lab](https://github.com/echeekene/splunk-security-automation).

CoreDNS writes one log line per query to stdout. On the node those land in `/var/log/containers/coredns-*.log`.

## Forward them

On the node, install the Splunk Universal Forwarder and add an input:

```ini
# $SPLUNK_HOME/etc/system/local/inputs.conf
[monitor:///var/log/containers/coredns-*.log]
index = dns
sourcetype = coredns
```

## Detection: possible DNS tunnelling

Very long query names, or many unique subdomains under one domain from one client, are a common sign of data exfiltration over DNS. See `queries/dns-tunnelling.spl`.
