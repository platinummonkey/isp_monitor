# ISP Monitor

Monitor your outbound connection and collect && report long term statistics to
observe if your ISP might be to blame for poor connections.

This is a simple utility and the scope is meant to solve a recurring problem...

## Configuration

You can specify a config file with `-config=<path...>` (defaults to `~/.isp_monitor`). 

It expects a YAML file like:

```yaml


reporters:
  - name: log
    type: log
  - name: datadog # assumes you have the datadog-agent locally running
    type: datadog

collectors:
  - name: device_to_local_gateway
    type: ping
    interval: 30s
    address: <ip>
  - name: device_to_isp_dns
    type: ping
    interval: 30s
    address: <ip>
  - name: device_to_external_stable_destination
    type: ping
    interval: 30s
    address: <ip/dns>
  - name: speedtest
    type: speedtest
    interval: 5m

```

