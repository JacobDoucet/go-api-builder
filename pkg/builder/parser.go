package builder

import (
	"fmt"
	"sort"
)

// Cassandra type to Go type
var cqlTypeToGoType = map[string]string{
	"ascii":     "string",
	"bigint":    "int64",
	"blob":      "[]byte",
	"boolean":   "bool",
	"counter":   "int64",
	"date":      "time.Time",
	"decimal":   "float64",
	"double":    "float64",
	"duration":  "int64",
	"float":     "float32",
	"inet":      "string",
	"int":       "int32",
	"smallint":  "int16",
	"text":      "string",
	"time":      "time.Time",
	"timestamp": "time.Time",
	"timeuuid":  "uuid.UUID",
	"tinyint":   "int8",
	"uuid":      "uuid.UUID",
	"varchar":   "string",
	"varint":    "int64",
}

func CreateTreeFrom(model Model) (nodes map[string]TreeNode, err error) {
	model, err = processModel(model)
	if err != nil {
		return
	}
	return createTree(model.Objects, model.Attributes, model.Relations)
}

func processModel(model Model) (_ Model, err error) {
	model.Attributes["createdAt"] = Attribute{
		Name:    "createdAt",
		CqlType: "timestamp",
	}
	model.Attributes["updatedAt"] = Attribute{
		Name:    "updatedAt",
		CqlType: "timestamp",
	}
	for name := range model.Attributes {
		err = processAttribute(name, model.Objects, model.Attributes)
		if err != nil {
			return
		}
	}
	for name := range model.Objects {
		err = processObject(name, model.Objects, model.Attributes)
		if err != nil {
			return
		}
	}
	for objectName := range model.Relations {
		err = processRelations(objectName, model.Objects, model.Relations)
	}
	for name := range model.Objects {
		err = processObjectAttributes(name, model.Objects, model.Attributes)
		if err != nil {
			return
		}
	}
	return model, nil
}

func processObject(name string, objects map[string]Object, attributes map[string]Attribute) (err error) {
	object := objects[name]
	defer func() {
		if err == nil {
			objects[name] = object
		}
	}()
	if object.Name == "" {
		object.Name = name
	}
	if object.Identifier == nil {
		identifier := name + "Id"
		object.Identifier = []string{identifier}
		attributes[identifier] = Attribute{
			Name:    identifier,
			CqlType: "uuid",
		}
		err = processAttribute(identifier, objects, attributes)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func getAttributeCheck(object Object) func(attr string) bool {
	objectAttrMap := make(map[string]interface{})
	for _, attributeName := range object.Identifier {
		objectAttrMap[attributeName] = struct{}{}
	}
	for _, attributeName := range object.Metadata {
		objectAttrMap[attributeName] = struct{}{}
	}
	return func(attr string) bool {
		if _, ok := objectAttrMap[attr]; ok {
			return true
		}
		objectAttrMap[attr] = struct{}{}
		return false
	}
}

func processRelations(name string, objects map[string]Object, relations map[string]Relation) (err error) {
	object := objects[name]
	objectHasAttribute := getAttributeCheck(object)
	defer func() {
		if err == nil {
			objects[name] = object
		}
	}()
	for _, relatedObjectName := range relations[name].Many {
		relatedObject := objects[relatedObjectName]
		relatedObjectHasAttr := getAttributeCheck(relatedObject)
		for _, attributeName := range object.Identifier {
			if !relatedObjectHasAttr(attributeName) {
				relatedObject.Metadata = append(relatedObject.Metadata, attributeName)
			}
		}
		objects[relatedObjectName] = relatedObject
	}
	for _, relatedObjectName := range relations[name].One {
		relatedObject := objects[relatedObjectName]
		relatedObjectHasAttr := getAttributeCheck(relatedObject)
		for _, attributeName := range relatedObject.Identifier {
			if !objectHasAttribute(attributeName) {
				object.Metadata = append(object.Metadata, attributeName)
			}
		}
		for _, attributeName := range relatedObject.Metadata {
			if !objectHasAttribute(attributeName) {
				object.Metadata = append(object.Metadata, attributeName)
			}
		}
		for _, attributeName := range object.Identifier {
			if !relatedObjectHasAttr(attributeName) {
				relatedObject.Metadata = append(relatedObject.Metadata, attributeName)
			}
		}
		objects[relatedObjectName] = relatedObject
	}
	return nil
}

func processObjectAttributes(name string, objects map[string]Object, attributes map[string]Attribute) (err error) {
	object := objects[name]
	defer func() {
		if err == nil {
			objects[name] = object
		}
	}()
	object.Attributes = make([]Attribute, 0, len(object.Identifier)+len(object.Metadata))
	for _, attributeName := range object.Identifier {
		attribute := attributes[attributeName]
		object.Attributes = append(object.Attributes, attribute)
		object.IdentifierAttributes = append(object.IdentifierAttributes, attribute)
	}
	object.Metadata = append(object.Metadata, "createdAt", "updatedAt")
	for _, attributeName := range object.Metadata {
		attribute := attributes[attributeName]
		object.Attributes = append(object.Attributes, attribute)
	}
	sortAttributes(object.Attributes)
	sortAttributes(object.IdentifierAttributes)
	return nil
}

func sortAttributes(attributes []Attribute) {
	sort.Slice(attributes, func(i, j int) bool {
		attrI := attributes[i]
		attrJ := attributes[j]
		if attrI.Name == "createdAt" {
			return true
		}
		if attrJ.Name == "createdAt" {
			return false
		}
		if attrI.Name == "updatedAt" {
			return true
		}
		if attrJ.Name == "updatedAt" {
			return false
		}
		return attrI.Name < attrJ.Name
	})
}

func processAttribute(name string, _ map[string]Object, attributes map[string]Attribute) (err error) {
	attribute := attributes[name]
	defer func() {
		attributes[name] = attribute
	}()
	if attribute.Name == "" {
		attribute.Name = name
	}
	if goType, ok := cqlTypeToGoType[attribute.CqlType]; ok {
		attribute.GoType = goType
	} else {
		return fmt.Errorf("unknown type %s", attribute.CqlType)
	}
	return nil
}

func createTree(
	objects map[string]Object,
	attributes map[string]Attribute,
	relations map[string]Relation,
) (nodes map[string]TreeNode, err error) {
	nodes = make(map[string]TreeNode)
	for name := range objects {
		nodes[name] = TreeNode{
			nodes:      nodes,
			objects:    make(map[string]Object),
			attributes: make(map[string]Attribute),
			Name:       name,
			Identifier: objects[name].Identifier,
			Children:   make([]string, 0),
			Peers:      make([]string, 0),
		}
		err = appendAttributes(objects[name], attributes, nodes[name].attributes)
		if err != nil {
			return
		}
	}
	for name := range relations {
		node := nodes[name]
		node.objects[name] = objects[name]
		for _, child := range relations[name].Many {
			node.objects[child] = objects[child]
			node.Children = append(node.Children, child)
			childNode := nodes[child]
			childNode.Parents = append(childNode.Parents, name)
			nodes[child] = childNode
			err = appendAttributes(objects[child], attributes, node.attributes)
			if err != nil {
				return
			}
		}
		for _, peer := range relations[name].One {
			node.objects[peer] = objects[peer]
			node.Peers = append(node.Peers, peer)
			peerNode := nodes[peer]
			peerNode.Peers = append(peerNode.Peers, name)
			nodes[peer] = peerNode
			err = appendAttributes(objects[peer], attributes, node.attributes)
			if err != nil {
				return
			}
		}
		nodes[name] = node
	}
	for name := range nodes {
		node := nodes[name]
		node.Tables, err = createTables(node)
		if err != nil {
			return
		}
		nodes[name] = node
	}
	printTree(nodes)
	return
}

func createTables(treeNode TreeNode) ([]Table, error) {
	object := treeNode.objects[treeNode.Name]
	tables := make([]Table, 0)

	table := Table{
		Name:          object.Name,
		PartitionKey:  object.Identifier,
		ClusteringKey: []string{"object_type"},
	}
	var err error
	objectNames := []string{treeNode.Name}
	objectNames = append(objectNames, treeNode.Peers...)
	table.Attributes, err = buildAttributesList(
		treeNode.objects,
		objectNames,
		treeNode.attributes,
	)
	if err != nil {
		return nil, err
	}
	tables = append(tables, table)

	for _, childName := range treeNode.Children {
		child := treeNode.objects[childName]
		childTable := Table{
			Name:          child.Name + "_by_" + object.Name,
			PartitionKey:  object.Identifier,
			ClusteringKey: child.Identifier,
		}
		childTable.Attributes, err = buildAttributesList(
			treeNode.objects,
			[]string{treeNode.Name, childName},
			treeNode.attributes,
		)
		if err != nil {
			return nil, err
		}
		tables = append(tables, childTable)
	}

	return tables, nil
}

func buildAttributesList(objects map[string]Object, names []string, attributes map[string]Attribute) ([]Attribute, error) {
	attributeSet := make(map[string]Attribute)

	for _, name := range names {
		peer := objects[name]
		err := appendAttributes(peer, attributes, attributeSet)
		if err != nil {
			return nil, err
		}
	}
	attributeList := make([]Attribute, 0, len(attributeSet))
	for attributeName := range attributeSet {
		attributeList = append(attributeList, attributeSet[attributeName])
	}
	sortAttributes(attributeList)
	return attributeList, nil
}

func appendAttributes(object Object, src map[string]Attribute, target map[string]Attribute) (err error) {
	for _, attributeName := range object.Identifier {
		attribute := src[attributeName]
		target[attributeName] = attribute
	}
	for _, attributeName := range object.Metadata {
		attribute := src[attributeName]
		target[attributeName] = attribute
	}
	return nil
}
