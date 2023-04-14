# Metrics
Fuddle nodes track metrics describing the health of the cluster.

Each node exposes an admin API `/metrics` endpoint which exports the Prometheus
metrics.

This document describes the available metrics and their labels.

## Cluster
* `fuddle.cluster.nodes.count` (gauge): Number of Fuddle nodes in the cluster
known by each node

## Registry
* `fuddle.registry.members.count` (gauge): Number of known members registered 
by each node. Labels:
  * `status`: The members status (either `up`, `down`, or `left`)
  * `owner`: The ID of the node that owns the member

* `fuddle.registry.members.owned` (gauge): Number of members owned by the node.
labels:
  * `status`: The members status (either `up`, `down`, or `left`)

## Errors
* `fuddle.errors` (counter): Number of errors logged on the node. Labels:
  * `subsystem`: The subsystem that logged the warning

* `fuddle.warnings` (counter): Number of warnings logged on the node. Labels:
  * `subsystem`: The subsystem that logged the warning
