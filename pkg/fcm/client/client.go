package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FuddleNodeInfo struct {
	ID        string `json:"id,omitempty"`
	RPCAddr   string `json:"rpc_addr,omitempty"`
	AdminAddr string `json:"admin_addr,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
}

// nolint
type ClientNodeInfo struct {
	ID      string `json:"id,omitempty"`
	LogPath string `json:"log_path,omitempty"`
}

type ClusterInfo struct {
	ID          string           `json:"id,omitempty"`
	FuddleNodes []FuddleNodeInfo `json:"nodes,omitempty"`
	ClientNodes []ClientNodeInfo `json:"members,omitempty"`
}

type ClusterHealth struct {
	Healthy bool `json:"healthy,omitempty"`
}

type NodesInfo struct {
	FuddleNodes []FuddleNodeInfo `json:"nodes,omitempty"`
	ClientNodes []ClientNodeInfo `json:"members,omitempty"`
}

type clusterRequest struct {
	FuddleNodes int `json:"nodes,omitempty"`
	ClientNodes int `json:"members,omitempty"`
}

type nodesRequest struct {
	FuddleNodes int `json:"nodes,omitempty"`
	ClientNodes int `json:"members,omitempty"`
}

type Client struct {
	addr       string
	httpClient *http.Client
}

func NewClient(addr string) *Client {
	return &Client{
		addr:       addr,
		httpClient: &http.Client{},
	}
}

func (c *Client) ClusterCreate(ctx context.Context, fuddleNodes int, clientNodes int) (ClusterInfo, error) {
	b, err := json.Marshal(&clusterRequest{
		FuddleNodes: fuddleNodes,
		ClientNodes: clientNodes,
	})
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster create: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://"+c.addr+"/cluster",
		bytes.NewReader(b),
	)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster create: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster create: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster create: request failed: bad status: %d", resp.StatusCode)
	}

	var clusterInfo ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&clusterInfo); err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster create: decode response: %w", err)
	}

	return clusterInfo, nil
}

func (c *Client) ClusterInfo(ctx context.Context, id string) (ClusterInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://"+c.addr+"/cluster/"+id,
		nil,
	)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster info: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster info: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster info: request failed: bad status: %d", resp.StatusCode)
	}

	var clusterInfo ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&clusterInfo); err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: cluster info: decode response: %w", err)
	}

	return clusterInfo, nil
}

func (c *Client) ClusterHealth(ctx context.Context, id string) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://"+c.addr+"/cluster/"+id+"/health",
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("fcm client: cluster health: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("fcm client: cluster health: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("fcm client: cluster health: request failed: bad status: %d", resp.StatusCode)
	}

	var clusterHealth ClusterHealth
	if err := json.NewDecoder(resp.Body).Decode(&clusterHealth); err != nil {
		return false, fmt.Errorf("fcm client: cluster health: decode response: %w", err)
	}

	return clusterHealth.Healthy, nil
}

func (c *Client) AddNodes(ctx context.Context, clusterID string, fuddleNodes int, clientNodes int) (NodesInfo, error) {
	nodesInfo, err := c.updateNodes(
		ctx,
		"http://"+c.addr+"/cluster/"+clusterID+"/nodes/add",
		nodesRequest{
			FuddleNodes: fuddleNodes,
			ClientNodes: clientNodes,
		},
	)
	if err != nil {
		return NodesInfo{}, fmt.Errorf("fcm client: add nodes: %w", err)
	}
	return nodesInfo, nil
}

func (c *Client) RemoveNodes(ctx context.Context, clusterID string, fuddleNodes int, clientNodes int) (NodesInfo, error) {
	nodesInfo, err := c.updateNodes(
		ctx,
		"http://"+c.addr+"/cluster/"+clusterID+"/nodes/remove",
		nodesRequest{
			FuddleNodes: fuddleNodes,
			ClientNodes: clientNodes,
		},
	)
	if err != nil {
		return NodesInfo{}, fmt.Errorf("fcm client: remove nodes: %w", err)
	}
	return nodesInfo, nil
}

func (c *Client) updateNodes(ctx context.Context, url string, nodesReq nodesRequest) (NodesInfo, error) {
	b, err := json.Marshal(nodesReq)
	if err != nil {
		return NodesInfo{}, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(b),
	)
	if err != nil {
		return NodesInfo{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NodesInfo{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NodesInfo{}, fmt.Errorf("request failed: bad status: %d", resp.StatusCode)
	}

	var nodesInfo NodesInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodesInfo); err != nil {
		return NodesInfo{}, fmt.Errorf("decode response: %w", err)
	}

	return nodesInfo, nil
}
