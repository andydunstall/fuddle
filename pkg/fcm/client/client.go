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

type clusterRequest struct {
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

func (c *Client) CreateCluster(ctx context.Context, fuddleNodes int, clientNodes int) (ClusterInfo, error) {
	b, err := json.Marshal(&clusterRequest{
		FuddleNodes: fuddleNodes,
		ClientNodes: clientNodes,
	})
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: create cluster: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://"+c.addr+"/cluster",
		bytes.NewReader(b),
	)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: create cluster: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: create cluster: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClusterInfo{}, fmt.Errorf("fcm client: create cluster: request failed: bad status: %d", resp.StatusCode)
	}

	var clusterInfo ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&clusterInfo); err != nil {
		return ClusterInfo{}, fmt.Errorf("fcm client: create cluster: decode response: %w", err)
	}

	return clusterInfo, nil
}
