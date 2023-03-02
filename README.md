<p align="center">
  <img src='assets/images/logo.png?raw=true' width='80%'>
</p>

---

> :warning: **Fuddle is still in development**

---

Fuddle is a service to manage your cluster.

# Getting Started

Start by downloading the appropriate Fuddle binary for your platform from the
[releases](https://github.com/andydunstall/fuddle/releases) page.

The quickest way to start using Fuddle is to run one of the [demos](./demos):

### [Random Number Service](./demos/random)
This is a toy example that starts a cluster to serve random numbers over HTTP.
The cluster uses Fuddle to discover the other nodes in the cluster and their
state to load balance requests evenly among the nodes.

Start the service using `fuddle demo random`.

### [Messaging Service](./demos/messaging)
The messaging service aims to provide a more real world example of using Fuddle.
The cluster runs a publish/subscribe service where channels are partitioned
among the set of nodes in the cluster using Fuddle.

Start the service using `fuddle demo messaging`.
