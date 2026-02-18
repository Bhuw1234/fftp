package deparrow

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// NodeTool provides the ability to list and inspect compute nodes.
// Nodes are the compute providers in the DEparrow network.
type NodeTool struct {
	client *Client
}

// NewNodeTool creates a new node tool.
func NewNodeTool(client *Client) *NodeTool {
	return &NodeTool{client: client}
}

// Name returns the tool name.
func (t *NodeTool) Name() string {
	return "deparrow_nodes"
}

// Description returns the tool description.
func (t *NodeTool) Description() string {
	return `List and inspect compute nodes on the DEparrow network.

Nodes are the backbone of the network, providing:
- CPU compute power
- GPU acceleration
- Storage capacity
- Network bandwidth

Use this tool to:
- Discover available compute resources
- Check node health and status
- View node contributions and tiers
`
}

// Parameters returns the JSON schema for tool parameters.
func (t *NodeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"node_id": map[string]interface{}{
				"type":        "string",
				"description": "Specific node ID to inspect (optional, lists all if not provided)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"online", "offline", "maintenance", "all"},
				"description": "Filter by node status",
				"default":     "online",
			},
			"contribution": map[string]interface{}{
				"type":        "boolean",
				"description": "Include contribution details for each node",
				"default":     false,
			},
		},
	}
}

// Execute runs the node tool.
func (t *NodeTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	// Check if specific node is requested
	if nodeID, ok := args["node_id"].(string); ok && nodeID != "" {
		return t.getNode(ctx, nodeID, args)
	}

	// List all nodes
	return t.listNodes(ctx, args)
}

// listNodes retrieves and displays all nodes.
func (t *NodeTool) listNodes(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	nodes, err := t.client.ListNodes(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to list nodes: %v", err))
	}

	// Filter by status
	statusFilter, _ := args["status"].(string)
	if statusFilter == "" {
		statusFilter = "online"
	}

	var filtered []Node
	for _, node := range nodes {
		switch statusFilter {
		case "online":
			if node.Status == NodeStatusOnline {
				filtered = append(filtered, node)
			}
		case "offline":
			if node.Status == NodeStatusOffline {
				filtered = append(filtered, node)
			}
		case "maintenance":
			if node.Status == NodeStatusMaintenance {
				filtered = append(filtered, node)
			}
		default:
			filtered = append(filtered, node)
		}
	}

	if len(filtered) == 0 {
		return tools.UserResult(fmt.Sprintf("No nodes found with status: %s", statusFilter))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🖥️  DEparrow Compute Nodes (%d/%d online)\n", len(filtered), len(nodes)))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	includeContribution, _ := args["contribution"].(bool)

	for i, node := range filtered {
		result.WriteString(fmt.Sprintf("%d. Node: %s\n", i+1, node.ID[:16]+"..."))
		result.WriteString(fmt.Sprintf("   Status: %s | Arch: %s\n", node.Status, node.Arch))

		if node.Resources != nil {
			res := node.Resources
			result.WriteString(fmt.Sprintf("   Resources: %d CPU", res.CPU))
			if res.GPU > 0 {
				result.WriteString(fmt.Sprintf(" | %d GPU", res.GPU))
				if res.GPUModel != "" {
					result.WriteString(fmt.Sprintf(" (%s)", res.GPUModel))
				}
			}
			result.WriteString("\n")
		}

		result.WriteString(fmt.Sprintf("   Credits Earned: %.2f\n", node.CreditsEarned))

		if includeContribution {
			contrib, err := t.client.GetNodeContribution(ctx, node.ID)
			if err == nil {
				result.WriteString(fmt.Sprintf("   Contribution: %.1f%% of network | Rank #%d\n",
					contrib.NetworkPercent, contrib.Rank))
			}
		}

		result.WriteString("\n")
	}

	return tools.UserResult(result.String())
}

// getNode retrieves details for a specific node.
func (t *NodeTool) getNode(ctx context.Context, nodeID string, args map[string]interface{}) *tools.ToolResult {
	node, err := t.client.GetNode(ctx, nodeID)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get node: %v", err))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🖥️  Node: %s\n", node.ID))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("Status:     %s\n", node.Status))
	result.WriteString(fmt.Sprintf("Arch:       %s\n", node.Arch))
	result.WriteString(fmt.Sprintf("Last Seen:  %s\n", node.LastSeen.Format("2006-01-02 15:04:05")))

	if node.Resources != nil {
		result.WriteString("\n📊 Resources:\n")
		result.WriteString(fmt.Sprintf("  CPU:      %d cores\n", node.Resources.CPU))
		if node.Resources.GPU > 0 {
			result.WriteString(fmt.Sprintf("  GPU:      %d", node.Resources.GPU))
			if node.Resources.GPUModel != "" {
				result.WriteString(fmt.Sprintf(" (%s)", node.Resources.GPUModel))
			}
			result.WriteString("\n")
		}
		if node.Resources.Memory != "" {
			result.WriteString(fmt.Sprintf("  Memory:   %s\n", node.Resources.Memory))
		}
	}

	result.WriteString(fmt.Sprintf("\n💰 Credits Earned: %.2f\n", node.CreditsEarned))

	// Get contribution details
	contrib, err := t.client.GetNodeContribution(ctx, nodeID)
	if err == nil {
		result.WriteString("\n📈 Contribution:\n")
		result.WriteString(fmt.Sprintf("  CPU Hours:    %.1f\n", contrib.CPUUsageHours))
		result.WriteString(fmt.Sprintf("  GPU Hours:    %.1f\n", contrib.GPUUsageHours))
		result.WriteString(fmt.Sprintf("  Network %%:    %.2f%%\n", contrib.NetworkPercent))
		result.WriteString(fmt.Sprintf("  Rank:         #%d of %d\n", contrib.Rank, contrib.TotalNodes))
		result.WriteString(fmt.Sprintf("  Tier:         %s\n", node.Tier))
	}

	// Show labels
	if len(node.Labels) > 0 {
		result.WriteString("\n🏷️  Labels:\n")
		for k, v := range node.Labels {
			result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	return tools.UserResult(result.String())
}

// NodeContributionTool provides detailed contribution statistics.
type NodeContributionTool struct {
	client *Client
}

// NewNodeContributionTool creates a new node contribution tool.
func NewNodeContributionTool(client *Client) *NodeContributionTool {
	return &NodeContributionTool{client: client}
}

// Name returns the tool name.
func (t *NodeContributionTool) Name() string {
	return "deparrow_contribution"
}

// Description returns the tool description.
func (t *NodeContributionTool) Description() string {
	return "View detailed contribution statistics for a specific node or your own node."
}

// Parameters returns the JSON schema for tool parameters.
func (t *NodeContributionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"node_id": map[string]interface{}{
				"type":        "string",
				"description": "Node ID to check (uses your node if not specified)",
			},
		},
	}
}

// Execute runs the node contribution tool.
func (t *NodeContributionTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	nodeID, ok := args["node_id"].(string)
	if !ok || nodeID == "" {
		// Try to get the user's own node
		// In a real implementation, this would be derived from the JWT or stored node ID
		return tools.ErrorResult("node_id is required")
	}

	contrib, err := t.client.GetNodeContribution(ctx, nodeID)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get contribution: %v", err))
	}

	node, err := t.client.GetNode(ctx, nodeID)
	if err != nil {
		node = &Node{ID: nodeID}
	}

	var result strings.Builder
	result.WriteString("📊 Node Contribution Report\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("Node ID: %s\n\n", nodeID[:16]+"..."))

	// Resource contribution
	result.WriteString("🖥️  Resources Contributed:\n")
	result.WriteString(fmt.Sprintf("  CPU Time:      %.1f hours\n", contrib.CPUUsageHours))
	result.WriteString(fmt.Sprintf("  GPU Time:      %.1f hours\n", contrib.GPUUsageHours))
	result.WriteString(fmt.Sprintf("  Live GFLOPS:   %.2f\n", contrib.LiveGFlops))

	// Network percentage
	result.WriteString("\n🌐 Network Impact:\n")
	result.WriteString(fmt.Sprintf("  Share:         %.2f%% of network\n", contrib.NetworkPercent))

	// Ranking
	result.WriteString("\n🏆 Ranking:\n")
	result.WriteString(fmt.Sprintf("  Position:      #%d of %d nodes\n", contrib.Rank, contrib.TotalNodes))
	result.WriteString(fmt.Sprintf("  Tier:          %s ", node.Tier))

	tierIcons := map[ContributionTier]string{
		TierBronze:    "🥉",
		TierSilver:    "🥈",
		TierGold:      "🥇",
		TierDiamond:   "💎",
		TierLegendary: "🔥",
	}
	if icon, ok := tierIcons[node.Tier]; ok {
		result.WriteString(icon)
	}
	result.WriteString("\n")

	// Credits
	result.WriteString(fmt.Sprintf("\n💰 Credits Earned: %.2f\n", node.CreditsEarned))

	// Progress to next tier
	result.WriteString("\n📈 Progress:\n")
	totalHours := contrib.CPUUsageHours + contrib.GPUUsageHours
	nextTier, hoursNeeded := getNextTier(node.Tier, totalHours)
	if hoursNeeded > 0 {
		result.WriteString(fmt.Sprintf("  Next Tier:     %s (%.0f more hours needed)\n", nextTier, hoursNeeded))
	} else {
		result.WriteString("  Max Tier:      Legendary! 🎉\n")
	}

	return tools.UserResult(result.String())
}

// getNextTier calculates the next tier and hours needed.
func getNextTier(current ContributionTier, totalHours float64) (ContributionTier, float64) {
	tierRequirements := []struct {
		tier  ContributionTier
		hours float64
	}{
		{TierSilver, 100},
		{TierGold, 1000},
		{TierDiamond, 5000},
		{TierLegendary, 10000},
	}

	for _, req := range tierRequirements {
		if totalHours < req.hours {
			return req.tier, req.hours - totalHours
		}
	}

	return "", 0
}

// OrchestratorTool provides orchestrator node information.
type OrchestratorTool struct {
	client *Client
}

// NewOrchestratorTool creates a new orchestrator tool.
func NewOrchestratorTool(client *Client) *OrchestratorTool {
	return &OrchestratorTool{client: client}
}

// Name returns the tool name.
func (t *OrchestratorTool) Name() string {
	return "deparrow_orchestrators"
}

// Description returns the tool description.
func (t *OrchestratorTool) Description() string {
	return "List orchestrator nodes in the DEparrow network."
}

// Parameters returns the JSON schema for tool parameters.
func (t *OrchestratorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute runs the orchestrator tool.
func (t *OrchestratorTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	// Note: This would call a real orchestrator list endpoint
	// For now, return a placeholder
	return tools.UserResult(`🎛️  DEparrow Orchestrators

Active orchestrators coordinate job distribution across the network.

Use 'deparrow_network' to see network statistics including orchestrator count.`)
}

// Ensure tools implement the Tool interface
var _ tools.Tool = (*NodeTool)(nil)
var _ tools.Tool = (*NodeContributionTool)(nil)
var _ tools.Tool = (*OrchestratorTool)(nil)
