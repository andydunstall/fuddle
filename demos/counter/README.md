# Counter Service

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

## Cluster
The cluster contains two types of nodes:

### Frontend Nodes
Frontends expose a WebSocket API for clients to connect to. These nodes are
stateless, and handle routing requests to the correct counter node.

When a request comes in to register for an ID, the frontend will lookup the
appropriate counter node, register with a gRPC stream, then forward count
updates to the WebSocket client.

When counter nodes join or leave the cluster, the frontend receives an update
from Fuddle and will rebalance to its counter service connections to ensure each
ID is registered with the correct node.

### Counter Nodes
Counter nodes aggregate the count for each ID. They expose a gRPC interface
which frontends connect to send send and receive counter updates.

Each node is responsible for a range of IDs using consistent hashing. The nodes
maintain their own view of the cluster using Fuddle, so if a frontend registers
an ID with the wrong node it will respond with an error, or if the counter nodes
join or leave, they will rebalance and return errors to any streams that are now
registered to the wrong node.

## Usage
If you haven’t already, download the `fuddle` binary for your platform from the
[releases](https://github.com/andydunstall/fuddle/releases) page.

Then run `fuddle demo counter` to start the cluster. This will spin up a local
cluster containing multiple Fuddle, frontend and counter service nodes.

The Fuddle dashboard for the cluster can be viewed at
[http://127.0.0.1:8221](http://127.0.0.1:8221). Alternatively you can inspect
the cluster using `fuddle status cluster` or `fuddle status node {node ID}`.

<p align="center">
  <img src='../../assets/images/counter-service-dashboard.png?raw=true' width='80%'>
</p>

Each frontend exposes a WebSocket endpoint at `ws://{addr}/{id}` to register an
ID and stream updates to the number of users registered with that ID.

To connect using [`wscat`](https://www.npmjs.com/package/wscat) use
`wscat -c ws://{addr}/{id}`.
