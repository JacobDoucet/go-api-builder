package builder

type Model struct {
	Namespace  string               `yaml:"namespace"`
	Objects    map[string]Object    `yaml:"objects"`
	Relations  map[string]Relation  `yaml:"relations"`
	Attributes map[string]Attribute `yaml:"attributes"`
}

type TreeNode struct {
	nodes      map[string]TreeNode
	objects    map[string]Object
	attributes map[string]Attribute
	Name       string
	Identifier []string
	Parents    []string
	Children   []string
	Peers      []string
	Tables     []Table
}

type Table struct {
	Name          string
	PartitionKey  []string
	ClusteringKey []string
	Attributes    []Attribute
}

func (tn *TreeNode) GetParents() []TreeNode {
	parents := make([]TreeNode, len(tn.Children))
	for i, child := range tn.Children {
		parents[i] = tn.nodes[child]
	}
	return parents
}

func (tn *TreeNode) GetChildren() []TreeNode {
	children := make([]TreeNode, len(tn.Children))
	for i, child := range tn.Children {
		children[i] = tn.nodes[child]
	}
	return children
}

func (tn *TreeNode) GetPeers() []TreeNode {
	peers := make([]TreeNode, len(tn.Peers))
	for i, peer := range tn.Peers {
		peers[i] = tn.nodes[peer]
	}
	return peers
}

type Object struct {
	Metadata             []string `yaml:"metadata"`
	Identifier           []string `yaml:"identifier"`
	Name                 string
	Attributes           []Attribute
	IdentifierAttributes []Attribute
}

type Relation struct {
	One  []string `yaml:"one"`
	Many []string `yaml:"many"`
}

type Attribute struct {
	Name    string `yaml:"name"`
	CqlType string `yaml:"type"`
	GoType  string `yaml:"-"`
}
