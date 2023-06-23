package builder

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"regexp"
	"strings"
	"text/template"
)

func toTitlecase(s string) string {
	// return s with the first letter capitalized
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	s = strings.ReplaceAll(s, " ", "")
	return s
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(s string) string {
	// For the string "IAmCamel Case", return "i_am_camel_case"
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func getPrimaryKey(table Table) string {
	pk := make([]string, 0, len(table.PartitionKey))
	ck := make([]string, 0, len(table.ClusteringKey))
	for _, attr := range table.PartitionKey {
		pk = append(pk, toSnakeCase(attr))
	}
	for _, attr := range table.ClusteringKey {
		ck = append(ck, toSnakeCase(attr))
	}

	return "(" + strings.Join(pk, ", ") + "), " + strings.Join(ck, ", ")
}

func parseMuxVar(model Model) func(string) string {
	return func(attribute string) string {
		attr, ok := model.Attributes[attribute]
		if !ok {
			return "???"
		}
		varAssign := "query." + toTitlecase(attribute)
		handleError := "\nif err != nil {\nrw.WriteHeader(http.StatusBadRequest)\nreturn\n}\n"
		switch attr.GoType {
		case "string":
			return varAssign + " = vars[\"" + attribute + "\"]"
		case "blob":
			return varAssign + " = []byte(vars[\"" + attribute + "\"])"
		case "uuid.UUID":
			return varAssign + ", err = uuid.Parse(vars[\"" + attribute + "\"])" + handleError
		case "int":
			return varAssign + ", err = strconv.Atoi(vars[\"" + attribute + "\"])" + handleError
		case "int64":
			return varAssign + ", err = strconv.ParseInt(vars[\"" + attribute + "\"], 10, 64)" + handleError
		case "int8":
			return "var " + attribute + " int64\n" + attribute + ", err = strconv.ParseInt(vars[\"" + attribute + "\"], 10, 8)" + handleError + "\n" + varAssign + " = int8(" + attribute + ")"
		case "int16":
			return "var " + attribute + " int64\n" + attribute + ", err = strconv.ParseInt(vars[\"" + attribute + "\"], 10, 16)" + handleError + "\n" + varAssign + " = int16(" + attribute + ")"
		case "int32":
			return "var " + attribute + " int64\n" + attribute + ", err = strconv.ParseInt(vars[\"" + attribute + "\"], 10, 32)" + handleError + "\n" + varAssign + " = int32(" + attribute + ")"
		case "float64":
			return varAssign + ", err = strconv.ParseFloat(vars[\"" + attribute + "\"], 64)" + handleError
		case "float32":
			return varAssign + ", err = strconv.ParseFloat(vars[\"" + attribute + "\"], 32)" + handleError
		case "time.Time":
			return varAssign + ", err = time.Parse(time.RFC3339, vars[\"" + attribute + "\"])" + handleError
		case "bool":
			return varAssign + ", err = strconv.ParseBool(vars[\"" + attribute + "\"])" + handleError

		default:
			return ""
		}
	}
}

func tableName(table Table) string {
	return table.Keyspace + "." + toSnakeCase(table.Name)
}

func getTreeNode(nodes map[string]TreeNode) func(string) TreeNode {
	return func(name string) TreeNode {
		return nodes[name]
	}
}

func getObject(model Model) func(string) Object {
	return func(name string) Object {
		return model.Objects[name]
	}
}

func getTemplateFuncs(model Model, nodes map[string]TreeNode) template.FuncMap {
	return template.FuncMap{
		"tc":          toTitlecase,
		"sc":          toSnakeCase,
		"primaryKey":  getPrimaryKey,
		"tableName":   tableName,
		"parseMuxVar": parseMuxVar(model),
		"getTreeNode": getTreeNode(nodes),
		"getObject":   getObject(model),
	}
}
