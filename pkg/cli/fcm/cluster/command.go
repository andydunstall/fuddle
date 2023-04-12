package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "create an fcm cluster",
	Run:   run,
}

type clusterRequest struct {
	Nodes   int `json:"nodes,omitempty"`
	Members int `json:"members,omitempty"`
}

type nodeResponse struct {
	ID        string `json:"id,omitempty"`
	RPCAddr   string `json:"rpc_addr,omitempty"`
	AdminAddr string `json:"admin_addr,omitempty"`
}

type clusterResponse struct {
	ID    string         `json:"id,omitempty"`
	Nodes []nodeResponse `json:"nodes,omitempty"`
}

func run(cmd *cobra.Command, args []string) {
	url := "http://" + addr + "/cluster"
	req := clusterRequest{
		Nodes:   nodes,
		Members: members,
	}
	b, err := json.Marshal(&req)
	if err != nil {
		fmt.Println("failed to encode request", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("request failed", err)
	}
	defer resp.Body.Close()

	var clusterResp clusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&clusterResp); err != nil {
		fmt.Println("failed to decode response", err)
	}

	fmt.Println("")
	fmt.Println("  ID:", clusterResp.ID)
	fmt.Println("")
	fmt.Println("  Nodes:")
	for _, n := range clusterResp.Nodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      RPC Addr:", n.RPCAddr)
		fmt.Println("      Admin Addr:", n.AdminAddr)
		fmt.Println("")
	}
}
