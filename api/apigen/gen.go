package apigen

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/logrusorgru/aurora"
	"github.com/zeromicro/goctl/util"
	"github.com/urfave/cli"
)

const apiTemplate = `
syntax = "v1"

info(
	title: // TODO: add title
	desc: // TODO: add description
	author: "{{.gitUser}}"
	email: "{{.gitEmail}}"
)

type request {
	// TODO: add members here and delete this comment
}

type response {
	// TODO: add members here and delete this comment
}

service {{.serviceName}} {
	@handler GetUser // TODO: set handler name and delete this comment
	get /users/id/:userId(request) returns(response)

	@handler CreateUser // TODO: set handler name and delete this comment
	post /users/create(request)
}
`

// ApiCommand create api template file
func ApiCommand(c *cli.Context) error {
	apiFile := c.String("o")
	if len(apiFile) == 0 {
		return errors.New("missing -o")
	}

	fp, err := util.CreateIfNotExist(apiFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	baseName := util.FileNameWithoutExt(filepath.Base(apiFile))
	if strings.HasSuffix(strings.ToLower(baseName), "-api") {
		baseName = baseName[:len(baseName)-4]
	} else if strings.HasSuffix(strings.ToLower(baseName), "api") {
		baseName = baseName[:len(baseName)-3]
	}
	t := template.Must(template.New("etcTemplate").Parse(apiTemplate))
	if err := t.Execute(fp, map[string]string{
		"gitUser":     getGitName(),
		"gitEmail":    getGitEmail(),
		"serviceName": baseName + "-api",
	}); err != nil {
		return err
	}

	fmt.Println(aurora.Green("Done."))
	return nil
}
