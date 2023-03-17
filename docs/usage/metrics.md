# Metrics
Fuddle generates time series metrics which are exported with Prometheus at the
`/metrics` endpoint.

## Registry Metrics
* `fuddle_registry_node_count`: Gauge of the number of nodes registered with
Fuddle,
* `fuddle_registry_update_count`: Counter of the number of updates made to the
registry, with a `type` label containing the update type (`register`,
`unregister` or `metadata`)
* `fuddle_registry_connection_count`: Gauge of the number of registry clients
connected to the node

## Go Metrics
Each node registers Golang system metrics, see [`NewGoCollector`](https://pkg.go.dev/github.com/prometheus/client_golang@v1.14.0/prometheus/collectors#NewGoCollector).
