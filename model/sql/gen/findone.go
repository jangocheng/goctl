package gen

import (
	"github.com/zeromicro/goctl/model/sql/template"
	"github.com/zeromicro/goctl/util"
	"github.com/zeromicro/goctl/util/stringx"
)

func genFindOne(table Table, withCache bool) (string, string, error) {
	camel := table.Name.ToCamel()
	text, err := util.LoadTemplate(category, findOneTemplateFile, template.FindOne)
	if err != nil {
		return "", "", err
	}

	output, err := util.With("findOne").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":                 withCache,
			"upperStartCamelObject":     camel,
			"lowerStartCamelObject":     stringx.From(camel).Untitle(),
			"originalPrimaryKey":        wrapWithRawString(table.PrimaryKey.Name.Source()),
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle(),
			"dataType":                  table.PrimaryKey.DataType,
			"cacheKey":                  table.PrimaryCacheKey.KeyExpression,
			"cacheKeyVariable":          table.PrimaryCacheKey.KeyLeft,
		})
	if err != nil {
		return "", "", err
	}

	text, err = util.LoadTemplate(category, findOneMethodTemplateFile, template.FindOneMethod)
	if err != nil {
		return "", "", err
	}

	findOneMethod, err := util.With("findOneMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject":     camel,
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle(),
			"dataType":                  table.PrimaryKey.DataType,
		})
	if err != nil {
		return "", "", err
	}

	return output.String(), findOneMethod.String(), nil
}
