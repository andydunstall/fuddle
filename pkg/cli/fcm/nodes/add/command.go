package add

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type nodesRequest struct {
	Nodes   int `json:"nodes,omitempty"`
	Members int `json:"members,omitempty"`
}

type nodeResponse struct {
	ID        string `json:"id,omitempty"`
	RPCAddr   string `json:"rpc_addr,omitempty"`
	AdminAddr string `json:"admin_addr,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
}

type memberResponse struct {
	ID      string `json:"id,omitempty"`
	LogPath string `json:"log_path,omitempty"`
}

type nodesResponse struct {
	Nodes   []nodeResponse   `json:"nodes,omitempty"`
	Members []memberResponse `json:"members,omitempty"`
}

var Command = &cobra.Command{
	Use:   "add",
	Short: "add nodes to a cluster",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	url := "http://" + addr + "/cluster/" + clusterID + "/nodes/add"
	req := nodesRequest{
		Nodes:   fuddleNodes,
		Members: clientNodes,
	}
	b, err := json.Marshal(&req)
	if err != nil {
		fmt.Println("failed to encode request", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("request failed", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("request failed: bad status:", resp.StatusCode)
		return
	}

	var nodesResp nodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodesResp); err != nil {
		fmt.Println("failed to decode response", err)
		return
	}

	if nodesResp.Nodes != nil {
		fmt.Println("  Nodes:")
		for _, n := range nodesResp.Nodes {
			fmt.Println("      ID:", n.ID)
			fmt.Println("      RPC Addr:", n.RPCAddr)
			fmt.Println("      Admin Addr:", n.AdminAddr)
			fmt.Println("      Log Path:", n.LogPath)
			fmt.Println("")
		}
	}

	if nodesResp.Members != nil {
		fmt.Println("  Members:")
		for _, n := range nodesResp.Members {
			fmt.Println("      ID:", n.ID)
			fmt.Println("      Log Path:", n.LogPath)
			fmt.Println("")
		}
	}
}
