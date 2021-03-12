package dartgen

import (
	"os"
	"text/template"

	"github.com/zeromicro/goctl/api/spec"
)

const dataTemplate = `// --{{with .Info}}{{.Title}}{{end}}--
{{ range .Types}}
class {{.Name}}{
	{{range .Members}}
	/// {{.Comment}}
	final {{.Type}} {{lowCamelCase .Name}};
	{{end}}
	{{.Name}}({ {{range .Members}}
		this.{{lowCamelCase .Name}},{{end}}
	});
	factory {{.Name}}.fromJson(Map<String,dynamic> m) {
		return {{.Name}}({{range .Members}}
			{{lowCamelCase .Name}}: {{if isDirectType .Type}}m['{{tagGet .Tag "json"}}']{{else if isClassListType .Type}}(m['{{tagGet .Tag "json"}}'] as List<dynamic>).map((i) => {{getCoreType .Type}}.fromJson(i)){{else}}{{.Type}}.fromJson(m['{{tagGet .Tag "json"}}']){{end}},{{end}}
		);
	}
	Map<String,dynamic> toJson() {
		return { {{range .Members}}
			'{{tagGet .Tag "json"}}': {{if isDirectType .Type}}{{lowCamelCase .Name}}{{else if isClassListType .Type}}{{lowCamelCase .Name}}.map((i) => i.toJson()){{else}}{{lowCamelCase .Name}}.toJson(){{end}},{{end}}
		};
	}
}
{{end}}
`

func genData(dir string, api *spec.ApiSpec) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = genTokens(dir)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dir+api.Service.Name+".dart", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	t := template.New("dataTemplate")
	t = t.Funcs(funcMap)
	t, err = t.Parse(dataTemplate)
	if err != nil {
		return err
	}

	return t.Execute(file, api)
}

func genTokens(dir string) error {
	path := dir + "tokens.dart"
	if fileExists(path) {
		return nil
	}

	tokensFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer tokensFile.Close()
	_, err = tokensFile.WriteString(tokensFileContent)
	return err
}
