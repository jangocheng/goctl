package tpl

import (
	"fmt"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/goctl/api/gogen"
	"github.com/zeromicro/goctl/docker"
	"github.com/zeromicro/goctl/kube"
	mongogen "github.com/zeromicro/goctl/model/mongo/generate"
	modelgen "github.com/zeromicro/goctl/model/sql/gen"
	rpcgen "github.com/zeromicro/goctl/rpc/generator"
	"github.com/zeromicro/goctl/util"
)

const templateParentPath = "/"

// GenTemplates writes the latest template text into file which is not exists
func GenTemplates(ctx *cli.Context) error {
	if err := errorx.Chain(
		func() error {
			return gogen.GenTemplates(ctx)
		},
		func() error {
			return modelgen.GenTemplates(ctx)
		},
		func() error {
			return rpcgen.GenTemplates(ctx)
		},
		func() error {
			return docker.GenTemplates(ctx)
		},
		func() error {
			return kube.GenTemplates(ctx)
		},
		func() error {
			return mongogen.Templates(ctx)
		},
	); err != nil {
		return err
	}

	dir, err := util.GetTemplateDir(templateParentPath)
	if err != nil {
		return err
	}

	fmt.Printf("Templates are generated in %s, %s\n", aurora.Green(dir),
		aurora.Red("edit on your risk!"))

	return nil
}

// CleanTemplates deletes all templates
func CleanTemplates(_ *cli.Context) error {
	err := errorx.Chain(
		func() error {
			return gogen.Clean()
		},
		func() error {
			return modelgen.Clean()
		},
		func() error {
			return rpcgen.Clean()
		},
		func() error {
			return docker.Clean()
		},
		func() error {
			return kube.Clean()
		},
		func() error {
			return mongogen.Clean()
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", aurora.Green("template are clean!"))
	return nil
}

// UpdateTemplates writes the latest template text into file,
// it will delete the older templates if there are exists
func UpdateTemplates(ctx *cli.Context) (err error) {
	category := ctx.String("category")
	defer func() {
		if err == nil {
			fmt.Println(aurora.Green(fmt.Sprintf("%s template are update!", category)).String())
		}
	}()
	switch category {
	case docker.Category():
		return docker.Update()
	case gogen.Category():
		return gogen.Update()
	case kube.Category():
		return kube.Update()
	case rpcgen.Category():
		return rpcgen.Update()
	case modelgen.Category():
		return modelgen.Update()
	case mongogen.Category():
		return mongogen.Update()
	default:
		err = fmt.Errorf("unexpected category: %s", category)
		return
	}
}

// RevertTemplates will overwrite the old template content with the new template
func RevertTemplates(ctx *cli.Context) (err error) {
	category := ctx.String("category")
	filename := ctx.String("name")
	defer func() {
		if err == nil {
			fmt.Println(aurora.Green(fmt.Sprintf("%s template are reverted!", filename)).String())
		}
	}()
	switch category {
	case docker.Category():
		return docker.RevertTemplate(filename)
	case kube.Category():
		return kube.RevertTemplate(filename)
	case gogen.Category():
		return gogen.RevertTemplate(filename)
	case rpcgen.Category():
		return rpcgen.RevertTemplate(filename)
	case modelgen.Category():
		return modelgen.RevertTemplate(filename)
	case mongogen.Category():
		return mongogen.RevertTemplate(filename)
	default:
		err = fmt.Errorf("unexpected category: %s", category)
		return
	}
}
