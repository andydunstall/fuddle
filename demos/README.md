# Demos

These demos provide services that use Fuddle to manage the cluster, which can
be run using `fuddle demo`.

## [Counter Service](./counter)

> :warning: **Counter service is still in development**

The counter service is a demo cluster that shows how Fuddle can be used for
application specific routing between nodes, rather than just basic round robin
load balancing.

Users register an ID, then the service streams updates on how many other users
are registered with the same ID. So if a user registers ID `foo`, the service
will increment the count and broadcast the updated count to all users registered
with ID `foo`. Similarly when a user unregisters the count is decremented and
broadcast.

To scale the cluster horizontally, each node in the cluster is responsible for a
range of IDs using consistent hashing. Therefore Fuddle is used to build the
hash ring of nodes, and receive updates when nodes join and leave to trigger a
rebalance.

Although this is a simple service, it show how Fuddle can be used for:
* Observability: View the nodes in the cluster and their state either through
the Fuddle dashboard or using the Fuddle CLI,
* Cluster discovery: Nodes use Fuddle to discover each other, and are notified
when nodes join, leave or update their state, which can be used for routing
requests to the appropriate node

See [`counter/`](./counter) for more usage information, documentation and the
demo source code.
