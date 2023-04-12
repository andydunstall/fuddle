package info

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "info",
	Short: "describe an fcm cluster",
	Run:   run,
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

type clusterResponse struct {
	ID      string           `json:"id,omitempty"`
	Nodes   []nodeResponse   `json:"nodes,omitempty"`
	Members []memberResponse `json:"members,omitempty"`
}

func run(cmd *cobra.Command, args []string) {
	url := "http://" + addr + "/cluster/" + clusterID
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("request failed", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("request failed: bad status:", resp.StatusCode)
		return
	}

	var clusterResp clusterResponse
	if err := json.NewDecoder(resp.Body).Decode(&clusterResp); err != nil {
		fmt.Println("failed to decode response", err)
		return
	}

	fmt.Println("")
	fmt.Println("  ID:", clusterResp.ID)
	fmt.Println("")

	fmt.Println("  Nodes:")
	for _, n := range clusterResp.Nodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      RPC Addr:", n.RPCAddr)
		fmt.Println("      Admin Addr:", n.AdminAddr)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}

	fmt.Println("  Members:")
	for _, n := range clusterResp.Members {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}
}
