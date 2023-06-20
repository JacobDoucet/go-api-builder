package builder

import (
	"fmt"
	"sort"
	"strings"
)

func printTree(nodes map[string]TreeNode) {
	sortedNames := make([]string, len(nodes))
	i := 0
	for name := range nodes {
		sortedNames[i] = name
		i++
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		printTreeNode(nodes[name])
	}
}

func printTreeNode(node TreeNode) {
	if len(node.Children)+len(node.Peers) == 0 {
		return
	}
	for _, peerNode := range node.GetPeers() {
		printTable(peerNode, node)
	}
	for _, childNode := range node.GetChildren() {
		printTable(node, childNode)
	}
}

func printTable(parentNode, childNode TreeNode) {
	fmt.Println("\n", childNode.Name+"_by_"+parentNode.Name)
	fmt.Println("  primary key:", formatPrimaryKey(parentNode.Identifier, childNode.Identifier))
	fmt.Println("  attributes:")
	for _, attribute := range parentNode.attributes {
		fmt.Println("   - ", attribute.Name, "cql<"+attribute.CqlType+">", "go<"+attribute.GoType+">")
	}
}

func formatPrimaryKey(pk, ck []string) string {
	return fmt.Sprintf("(%s), %s", strings.Join(pk, ", "), strings.Join(ck, ", "))
}
