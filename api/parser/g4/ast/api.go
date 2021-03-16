package ast

import (
	"fmt"
	"sort"

	"github.com/tal-tech/go-zero/core/lang"
	"github.com/zeromicro/goctl/api/parser/g4/gen/api"
)

// Api describes syntax for api
type Api struct {
	LinePrefix string
	Syntax     *SyntaxExpr
	Import     []*ImportExpr
	importM    map[string]PlaceHolder
	Info       *InfoExpr
	Type       []TypeExpr
	typeM      map[string]TypeExpr
	Service    []*Service
	serviceM   map[string]PlaceHolder
	handlerM   map[string]PlaceHolder
	routeM     map[string]PlaceHolder
	ImportInfo map[string][]*ImportInfo
}

// VisitApi implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitApi(ctx *api.ApiContext) interface{} {
	var final Api
	final.importM = map[string]PlaceHolder{}
	final.typeM = map[string]TypeExpr{}
	final.serviceM = map[string]PlaceHolder{}
	final.handlerM = map[string]PlaceHolder{}
	final.routeM = map[string]PlaceHolder{}
	for _, each := range ctx.AllSpec() {
		root := each.Accept(v).(*Api)
		v.acceptSyntax(root, &final)
		v.acceptImport(root, &final)
		v.acceptInfo(root, &final)
		v.acceptType(root, &final)
		v.acceptService(root, &final)
	}

	return &final
}

func (v *ApiVisitor) acceptService(root *Api, final *Api) {
	for _, service := range root.Service {
		if _, ok := final.serviceM[service.ServiceApi.Name.Text()]; !ok && len(final.serviceM) > 0 {
			v.panic(service.ServiceApi.Name, fmt.Sprintf("mutiple service declaration"))
		}
		v.duplicateServerItemCheck(service)

		for _, route := range service.ServiceApi.ServiceRoute {
			uniqueRoute := fmt.Sprintf("%s %s", route.Route.Method.Text(), route.Route.Path.Text())
			if _, ok := final.routeM[uniqueRoute]; ok {
				v.panic(route.Route.Method, fmt.Sprintf("duplicate route: %s", uniqueRoute))
			}

			final.routeM[uniqueRoute] = Holder
			var handlerExpr Expr
			if route.AtServer != nil {
				atServerM := map[string]PlaceHolder{}
				for _, kv := range route.AtServer.Kv {
					if _, ok := atServerM[kv.Key.Text()]; ok {
						v.panic(kv.Key, fmt.Sprintf("duplicate key: %s", kv.Key.Text()))
					}
					atServerM[kv.Key.Text()] = Holder
					if kv.Key.Text() == "handler" {
						handlerExpr = kv.Value
					}
				}
			}

			if route.AtHandler != nil {
				handlerExpr = route.AtHandler.Name
			}

			if handlerExpr == nil {
				v.panic(route.Route.Method, fmt.Sprintf("mismtached handler"))
			}

			if handlerExpr.Text() == "" {
				v.panic(handlerExpr, fmt.Sprintf("mismtached handler"))
			}

			if _, ok := final.handlerM[handlerExpr.Text()]; ok {
				v.panic(handlerExpr, fmt.Sprintf("duplicate handler: %s", handlerExpr.Text()))
			}
			final.handlerM[handlerExpr.Text()] = Holder
		}
		final.Service = append(final.Service, service)
	}
}

func (v *ApiVisitor) duplicateServerItemCheck(service *Service) {
	if service.AtServer != nil {
		atServerM := map[string]PlaceHolder{}
		for _, kv := range service.AtServer.Kv {
			if _, ok := atServerM[kv.Key.Text()]; ok {
				v.panic(kv.Key, fmt.Sprintf("duplicate key: %s", kv.Key.Text()))
			}

			atServerM[kv.Key.Text()] = Holder
		}
	}
}

func (v *ApiVisitor) acceptType(root *Api, final *Api) {
	for _, tp := range root.Type {
		if _, ok := final.typeM[tp.NameExpr().Text()]; ok {
			v.panic(tp.NameExpr(), fmt.Sprintf("duplicate type: %s", tp.NameExpr().Text()))
		}

		v.acceptField(tp)
		final.typeM[tp.NameExpr().Text()] = tp
		final.Type = append(final.Type, tp)
	}
}

func (v *ApiVisitor) acceptField(tp TypeExpr) {
	st, ok := tp.(*TypeStruct)
	if !ok {
		return
	}

	m := make(map[string]lang.PlaceholderType)
	for _, f := range st.Fields {
		if f.IsAnonymous {
			continue
		}

		if _, ok := m[f.Name.Text()]; ok {
			v.panic(f.Name, fmt.Sprintf("duplicate field: %s", f.Name.Text()))
		}

		m[f.Name.Text()] = lang.Placeholder
	}
}

func (v *ApiVisitor) acceptInfo(root *Api, final *Api) {
	if root.Info != nil {
		infoM := map[string]PlaceHolder{}
		if final.Info != nil {
			v.panic(root.Info.Info, fmt.Sprintf("mutiple info declaration"))
		}

		for _, value := range root.Info.Kvs {
			if _, ok := infoM[value.Key.Text()]; ok {
				v.panic(value.Key, fmt.Sprintf("duplicate key: %s", value.Key.Text()))
			}
			infoM[value.Key.Text()] = Holder
		}

		final.Info = root.Info
	}
}

func (v *ApiVisitor) acceptImport(root *Api, final *Api) {
	for _, imp := range root.Import {
		importValue := imp.Value.Text()
		if _, ok := final.importM[importValue]; ok {
			v.panic(imp.Import, fmt.Sprintf("duplicate import: %s", importValue))
		}

		final.importM[imp.Value.Text()] = Holder
		final.Import = append(final.Import, imp)
	}
}

func (v *ApiVisitor) acceptSyntax(root *Api, final *Api) {
	if root.Syntax != nil {
		if final.Syntax != nil {
			v.panic(root.Syntax.Syntax, fmt.Sprintf("mutiple syntax declaration"))
		}

		final.Syntax = root.Syntax
	}
}

// VisitSpec implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitSpec(ctx *api.SpecContext) interface{} {
	var root Api
	if ctx.SyntaxLit() != nil {
		root.Syntax = ctx.SyntaxLit().Accept(v).(*SyntaxExpr)
	}

	if ctx.ImportSpec() != nil {
		root.Import = ctx.ImportSpec().Accept(v).([]*ImportExpr)
	}

	if ctx.InfoSpec() != nil {
		root.Info = ctx.InfoSpec().Accept(v).(*InfoExpr)
	}

	if ctx.TypeSpec() != nil {
		tp := ctx.TypeSpec().Accept(v)
		root.Type = tp.([]TypeExpr)
	}

	if ctx.ServiceSpec() != nil {
		root.Service = []*Service{ctx.ServiceSpec().Accept(v).(*Service)}
	}

	return &root
}

// Format provides a formatter for api command, now nothing to do
func (a *Api) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Api are equal
func (a *Api) Equal(v interface{}) bool {
	if v == nil {
		return false
	}

	root, ok := v.(*Api)
	if !ok {
		return false
	}

	if !a.Syntax.Equal(root.Syntax) {
		return false
	}

	if len(a.Import) != len(root.Import) {
		return false
	}

	var expectingImport, actualImport []*ImportExpr
	expectingImport = append(expectingImport, a.Import...)
	actualImport = append(actualImport, root.Import...)

	sort.Slice(expectingImport, func(i, j int) bool {
		return expectingImport[i].Value.Text() < expectingImport[j].Value.Text()
	})

	sort.Slice(actualImport, func(i, j int) bool {
		return actualImport[i].Value.Text() < actualImport[j].Value.Text()
	})

	for index, each := range expectingImport {
		ac := actualImport[index]
		if !each.Equal(ac) {
			return false
		}
	}

	if !a.Info.Equal(root.Info) {
		return false
	}

	if len(a.Type) != len(root.Type) {
		return false
	}

	var expectingType, actualType []TypeExpr
	expectingType = append(expectingType, a.Type...)
	actualType = append(actualType, root.Type...)

	sort.Slice(expectingType, func(i, j int) bool {
		return expectingType[i].NameExpr().Text() < expectingType[j].NameExpr().Text()
	})
	sort.Slice(actualType, func(i, j int) bool {
		return actualType[i].NameExpr().Text() < actualType[j].NameExpr().Text()
	})

	for index, each := range expectingType {
		ac := actualType[index]
		if !each.Equal(ac) {
			return false
		}
	}

	if len(a.Service) != len(root.Service) {
		return false
	}

	var expectingService, actualService []*Service
	expectingService = append(expectingService, a.Service...)
	actualService = append(actualService, root.Service...)
	for index, each := range expectingService {
		ac := actualService[index]
		if !each.Equal(ac) {
			return false
		}
	}

	return true
}
