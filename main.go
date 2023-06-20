package main

import (
	"eflux-go-cassandra/pkg/builder"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var ModelFile = os.Getenv("MODEL_FILE")

func main() {
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
	for name := range model.Objects {
		err = builder.GenerateGolang(name, model, tree)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Builder completed successfully")
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
