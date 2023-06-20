package main

import (
	"eflux-go-cassandra/pkg/builder"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sort"
)

var ModelFile = os.Getenv("MODEL_FILE")
var OutDir = os.Getenv("OUT_DIR")

func main() {
	log.Println("Builder started")
	var model builder.Model
	var tree map[string]builder.TreeNode
	var err error
	model, err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	tree, err = builder.CreateTreeFrom(model)
	if err != nil {
		log.Fatal(err)
	}
	printTree(tree)
	err = builder.GeneratePackages(model, tree, OutDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Builder completed successfully")
}

func printTree(tree map[string]builder.TreeNode) {
	fmt.Println("tables")
	tables := make([]string, 0)
	for name := range tree {
		if len(tree[name].Tables) > 0 {
			for _, table := range tree[name].Tables {
				tables = append(tables, table.Name)
			}
		}
	}
	sort.Strings(tables)
	for _, table := range tables {
		fmt.Println(" -" + table)
	}
}

func loadConfig() (model builder.Model, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to load config: %v", err)
		}
	}()
	var yamlFile []byte
	yamlFile, err = os.ReadFile(ModelFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlFile, &model)
	return
}
