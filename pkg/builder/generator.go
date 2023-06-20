package builder

import (
	_ "embed"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"strings"
	"text/template"
)

//go:embed templates/model.go.tmpl
var modelTemplate string

//go:embed templates/cql.go.tmpl
var cqlTemplate string

//go:embed templates/cql-create.cql.tmpl
var cqlCreateTemplate string

//go:embed templates/api.go.tmpl
var apiTemplate string

type Data struct {
	Pkg        string
	Name       string
	Object     Object
	Attributes map[string]Attribute
	TreeNode   TreeNode
}

var templateFuncs = template.FuncMap{
	"tc": func(s string) string {
		// return s with the first letter capitalized
		s = strings.ReplaceAll(s, "_", " ")
		s = cases.Title(language.Und, cases.NoLower).String(s)
		s = strings.ReplaceAll(s, " ", "")
		return s
	},
	"primaryKey": func(table Table) string {
		pk := strings.Join(table.PartitionKey, ", ")
		ck := strings.Join(table.ClusteringKey, ", ")
		return "(" + pk + "), " + ck
	},
}

func GenerateGolang(name string, model Model, tree map[string]TreeNode) error {
	if cqlTemplate == "" {
		return fmt.Errorf("cassandra template is empty")
	}
	modelTmp, err := template.New("Model").Funcs(templateFuncs).Parse(modelTemplate)
	if err != nil {
		return err
	}
	cqlTmpl, err := template.New("Cassandra").Funcs(templateFuncs).Parse(cqlTemplate)
	if err != nil {
		return err
	}
	cqlCreateTmpl, err := template.New("CassandraCreateScript").Funcs(templateFuncs).Parse(cqlCreateTemplate)
	if err != nil {
		return err
	}
	apiTmpl, err := template.New("Api").Funcs(templateFuncs).Parse(apiTemplate)
	if err != nil {
		return err
	}

	_ = os.Mkdir("generated", 0755)
	_ = os.Mkdir("generated/model", 0755)
	_ = os.Mkdir("generated/cql", 0755)
	_ = os.Mkdir("generated/cql/scripts", 0755)
	_ = os.Mkdir("generated/api", 0755)
	if err != nil {
		return err
	}

	err = generate("model/"+name+".go", modelTmp, Data{
		Pkg:      "model",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}
	err = generate("cql/"+name+".go", cqlTmpl, Data{
		Pkg:      "cql",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}
	for _, table := range tree[name].Tables {
		err = generate("cql/scripts/create-"+table.Name+".cql", cqlCreateTmpl, table)
		if err != nil {
			return err
		}
	}
	err = generate("api/"+name+".go", apiTmpl, Data{
		Pkg:      "api",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}
	return nil
}

func generate(name string, tmpl *template.Template, data interface{}) error {
	var file *os.File
	file, err := os.Create("generated/" + name)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return file.Close()
}
