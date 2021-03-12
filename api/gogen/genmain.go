package gogen

import (
	"fmt"
	"strings"

	"github.com/zeromicro/goctl/api/spec"
	"github.com/zeromicro/goctl/config"
	ctlutil "github.com/zeromicro/goctl/util"
	"github.com/zeromicro/goctl/util/format"
	"github.com/zeromicro/goctl/vars"
)

const mainTemplate = `package main

import (
	"flag"
	"fmt"

	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
`

func genMain(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	name := strings.ToLower(api.Service.Name)
	if strings.HasSuffix(name, "-api") {
		name = strings.ReplaceAll(name, "-api", "")
	}
	filename, err := format.FileNamingFormat(cfg.NamingFormat, name)
	if err != nil {
		return err
	}

	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          "",
		filename:        filename + ".go",
		templateName:    "mainTemplate",
		category:        category,
		templateFile:    mainTemplateFile,
		builtinTemplate: mainTemplate,
		data: map[string]string{
			"importPackages": genMainImports(parentPkg),
			"serviceName":    api.Service.Name,
		},
	})
}

func genMainImports(parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, configDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, handlerDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"\n", ctlutil.JoinPackages(parentPkg, contextDir)))
	imports = append(imports, fmt.Sprintf("\"%s/core/conf\"", vars.ProjectOpenSourceURL))
	imports = append(imports, fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}
