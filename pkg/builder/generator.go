package builder

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"text/template"
)

//go:embed templates/model.go.tmpl
var modelTemplate string

//go:embed templates/cql.go.tmpl
var cqlTemplate string

//go:embed templates/cql-create-table.cql.tmpl
var cqlCreateTableTemplate string

//go:embed templates/api.common.go.tmpl
var apiCommonTemplate string

//go:embed templates/api.go.tmpl
var apiTemplate string

//go:embed templates/http.go.tmpl
var httpTemplate string

//go:embed templates/routes.go.tmpl
var routesTemplate string

type ObjectData struct {
	Pkg        string
	Name       string
	Object     Object
	Attributes map[string]Attribute
	TreeNode   TreeNode
}

type GlobalData struct {
	Pkg        string
	Objects    map[string]Object
	Attributes map[string]Attribute
	Tree       map[string]TreeNode
}

func GeneratePackages(model Model, tree map[string]TreeNode, out string) (err error) {
	generate := getGenerator(out)
	_ = os.Mkdir(out, 0755)
	_ = os.Mkdir(out+"/model", 0755)
	_ = os.Mkdir(out+"/cql", 0755)
	_ = os.Mkdir(out+"/cql/scripts", 0755)
	_ = os.Mkdir(out+"/api", 0755)
	_ = os.Mkdir(out+"/http", 0755)

	apiCommonTmpl, err := template.New("ApiCommon").Funcs(getTemplateFuncs(model, tree)).Parse(apiCommonTemplate)
	if err != nil {
		return err
	}
	err = generate("api/common.go", apiCommonTmpl, ObjectData{
		Pkg: "api",
	})
	for name := range model.Objects {
		err = generateObjectCode(generate, name, model, tree)
		if err != nil {
			return
		}
	}
	return
}

func generateObjectCode(generate GeneratorFunc, name string, model Model, tree map[string]TreeNode) error {
	templateFuncs := getTemplateFuncs(model, tree)
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
	cqlCreateTmpl, err := template.New("CassandraCreateScript").Funcs(templateFuncs).Parse(cqlCreateTableTemplate)
	if err != nil {
		return err
	}
	apiTmpl, err := template.New("Api").Funcs(templateFuncs).Parse(apiTemplate)
	if err != nil {
		return err
	}
	httpTmpl, err := template.New("Http").Funcs(templateFuncs).Parse(httpTemplate)
	if err != nil {
		return err
	}
	routesTmpl, err := template.New("Routes").Funcs(templateFuncs).Parse(routesTemplate)
	if err != nil {
		return err
	}

	err = generate("model/"+toSnakeCase(name)+".go", modelTmp, ObjectData{
		Pkg:      "model",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}
	err = generate("cql/"+toSnakeCase(name)+".go", cqlTmpl, ObjectData{
		Pkg:      "cql",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}

	for _, table := range tree[name].Tables {
		err = generate("cql/scripts/create_"+toSnakeCase(table.Name)+".cql", cqlCreateTmpl, table)
		if err != nil {
			return err
		}
	}

	err = generate("api/"+toSnakeCase(name)+".go", apiTmpl, ObjectData{
		Pkg:      "api",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}

	err = generate("http/"+toSnakeCase(name)+".go", httpTmpl, ObjectData{
		Pkg:      "http",
		Name:     name,
		Object:   model.Objects[name],
		TreeNode: tree[name],
	})
	if err != nil {
		return err
	}
	err = generate("http/routes.go", routesTmpl, GlobalData{
		Pkg:        "http",
		Objects:    model.Objects,
		Attributes: model.Attributes,
		Tree:       tree,
	})
	if err != nil {
		return err
	}

	return nil
}

type GeneratorFunc func(string, *template.Template, interface{}) error

func getGenerator(out string) GeneratorFunc {
	return func(name string, tmpl *template.Template, data interface{}) error {
		var buf bytes.Buffer
		var err error

		err = tmpl.Execute(&buf, data)
		if err != nil {
			return err
		}

		generatedCode := buf.Bytes()
		// format go code
		if matched, _ := regexp.Match(`.go$`, []byte(name)); matched {
			generatedCode, err = format.Source(buf.Bytes())
			if err != nil {
				return err
			}
		}

		var file *os.File
		file, err = os.Create(out + "/" + name)
		if err != nil {
			return err
		}

		_, err = file.Write(generatedCode)
		if err != nil {
			return err
		}

		return file.Close()
	}
}
