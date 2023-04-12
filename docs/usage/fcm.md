# FCM (Fuddle Cluster Manager)

FCM is a tool for spinning up and testing local Fuddle clusters.

Use FCM with the `fuddle fcm` command in the Fuddle CLI.

FCM runs a HTTP server (used by the Fuddle CLI) which can be used to automate
creating clusters.

## Start FCM
FCM runs as a HTTP server which manages all active clusters.

Start FCM with `fuddle fcm start`. By default FCM runs on port `8220`.

## Create A Cluster
Once the FCM server is running, create a new Fuddle cluster with
`fcm cluster create`.

By default this will create a cluster with 3 Fuddle nodes and 10 random members,
which can be changed with `--nodes` and `--members` options.

## Prometheus
To get Prometheus metrics, FCM exposes a `/cluster/{id}/prometheus/targets`
endpoint which returns the set of Prometheus targets in the cluster. The
set of targets will be updated as the nodes in the cluster change.

To integrate with Prometheus running locally, configure `http_sd_config` to
point to the targets URL for you're cluster, such as:

```yaml
global:
  scrape_interval: 10s
  evaluation_interval: 10s

scrape_configs:
  - job_name: 'fuddle'
    metrics_path: '/metrics'
    scheme: 'http'

    http_sd_configs:
      - url: 'http://localhost:8220/cluster/7cde0ff4/prometheus'
        refresh_interval: 10s
```

Where `7cde0ff4` should be replaced by the ID of you're cluster.

Note FCM runs each node in a separate goroutine instead of separate processes,
so the Go runtime metrics will be wrong.
