package generate

import (
	"errors"
	"path/filepath"

	"github.com/zeromicro/goctl/config"
	"github.com/zeromicro/goctl/model/mongo/template"
	"github.com/zeromicro/goctl/util"
	"github.com/zeromicro/goctl/util/format"
)

// Context defines the model generation data what they needs
type Context struct {
	Types  []string
	Cache  bool
	Output string
	Cfg    *config.Config
}

// Do executes model template and output the result into the specified file path
func Do(ctx *Context) error {
	if ctx.Cfg == nil {
		return errors.New("missing config")
	}

	err := generateModel(ctx)
	if err != nil {
		return err
	}

	return generateError(ctx)
}

func generateModel(ctx *Context) error {
	for _, t := range ctx.Types {
		fn, err := format.FileNamingFormat(ctx.Cfg.NamingFormat, t+"_model")
		if err != nil {
			return err
		}

		text, err := util.LoadTemplate(category, modelTemplateFile, template.Text)
		if err != nil {
			return err
		}

		output := filepath.Join(ctx.Output, fn+".go")
		err = util.With("model").Parse(text).GoFmt(true).SaveTo(map[string]interface{}{
			"Type":  t,
			"Cache": ctx.Cache,
		}, output, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateError(ctx *Context) error {
	text, err := util.LoadTemplate(category, errTemplateFile, template.Error)
	if err != nil {
		return err
	}

	output := filepath.Join(ctx.Output, "error.go")

	return util.With("error").Parse(text).GoFmt(true).SaveTo(ctx, output, false)
}
