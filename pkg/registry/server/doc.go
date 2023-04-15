// Package server receives and send updates to the registry on this node.
//
// The servers are split into
// * Client/Replica: The client server is used by external clients running
// on user applications, and the replica server is used by other Fuddle nodes
// in the cluster
// * Read/Write: The read server is used to stream updates to the registry in
// the local node, and the write server is used to send updates to the registry
//
// The read servers and write servers are split since in the future can run the
// read server on a read-only cache node as its just forwarding updates, and the
// write server on a Fuddle node as it maintains updates.
package server
