# FCM (Fuddle Cluster Manager)
FCM is a tool for spinning up local Fuddle clusters, used for development and
testing.

The FCM server manages the set of active clusters, which exposes a REST API to
create and modify clusters. This makes it easy to add FCM clients in different
languages (such as for SDK testing) and add FCM support to the Fuddle CLI.

# Usage
This describes how to use FCM via the Fuddle CLI, though the same functionality
can be achieved by querying the REST API directly.

This is just an overview of the available commands, a full list of commands can
be seen using `fuddle fcm help`.

## Server
Before you can create a cluster, you must start the FCM server.

The easiest way to start FCM is using the docker-compose cluster, which can be
run with `docker-compose -f dev/fcm/docker-compose.yml up`.

This will:
* Start the FCM server listening on port `8220`
* Start a default cluster with ID `default`
* Start Prometheus and Grafana that monitor the nodes in the default cluster

### Grafana
To use Grafana to monitor the default cluster:
* Add a Prometheus data source with address `prometheus:9090`
* Either create you're own dashboard or import the Fuddle dashboard from
`monitoring/grafana/fuddle.json`

## Create A Cluster
Once the FCM server is running, create a new cluster using `fuddle fcm cluster
create`.

Each cluster is given a unique ID since FCM may run multiple clusters at once.

A cluster contains:
* Fuddle nodes that maintain the registry
* Client nodes that connect to Fuddle and registry a random member

By default, a cluster contains 3 Fuddle nodes and 10 clients, which can be
changed with `--fuddle-nodes` and `--client-nodes` flags.

The cluster can be inspected at any time with `fuddle fcm cluster info`.

### Prometheus
Each cluster has a HTTP endpoint that returns an up to date list of Prometheus
targets of the set of nodes in the cluster at
`/cluster/{id}/prometheus/targets`. This should be used when configuring
Prometheus (using `http_sd_config`) instead of a static list of targets since
the set of nodes may change.

For example you could configure Prometheus running locally with:
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

Where `7cde0ff4` is the cluster ID that should be replaced with the ID of your cluster.

## Add/Remove Nodes
Nodes can be added and removed from running clusters with `fuddle fcm nodes add
{cluster ID}` and `fuddle fcm nodes remove {cluster ID}` respectively.

Similar to creating a cluster, each command accepts flags:
* `--fuddle-nodes`: Number of Fuddle nodes to add/remove
* `--client-nodes`: Number of client nodes to add/remove
