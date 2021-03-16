package ast

import (
	"github.com/zeromicro/goctl/api/parser/g4/gen/api"
)

// ImportExpr defines import syntax for api
type ImportExpr struct {
	Import      Expr
	Value       Expr
	As          Expr
	Package     Expr
	DocExpr     []Expr
	CommentExpr Expr
}

// ImportValue describes the value and package of import
type ImportValue struct {
	Value   Expr
	As      Expr
	Package Expr
}

// ImportPackage describes the package alias
type ImportPackage struct {
	As      Expr
	Package Expr
}

// VisitImportSpec implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportSpec(ctx *api.ImportSpecContext) interface{} {
	var list []*ImportExpr
	if ctx.ImportLit() != nil {
		lits := ctx.ImportLit().Accept(v).([]*ImportExpr)
		list = append(list, lits...)
	}
	if ctx.ImportBlock() != nil {
		blocks := ctx.ImportBlock().Accept(v).([]*ImportExpr)
		list = append(list, blocks...)
	}

	return list
}

// VisitImportLit implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportLit(ctx *api.ImportLitContext) interface{} {
	importToken := v.newExprWithToken(ctx.GetImportToken())
	valueExpr := ctx.ImportValue().Accept(v).(*ImportValue)
	return []*ImportExpr{
		{
			Import:      importToken,
			Value:       valueExpr.Value,
			As:          valueExpr.As,
			Package:     valueExpr.Package,
			DocExpr:     v.getDoc(ctx),
			CommentExpr: v.getComment(ctx),
		},
	}
}

// VisitImportBlock implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportBlock(ctx *api.ImportBlockContext) interface{} {
	importToken := v.newExprWithToken(ctx.GetImportToken())
	values := ctx.AllImportBlockValue()
	var list []*ImportExpr

	for _, value := range values {
		importExpr := value.Accept(v).(*ImportExpr)
		importExpr.Import = importToken
		list = append(list, importExpr)
	}

	return list
}

// VisitImportBlockValue implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportBlockValue(ctx *api.ImportBlockValueContext) interface{} {
	value := ctx.ImportValue().Accept(v).(*ImportValue)
	return &ImportExpr{
		As:          value.As,
		Package:     value.Package,
		Value:       value.Value,
		DocExpr:     v.getDoc(ctx),
		CommentExpr: v.getComment(ctx),
	}
}

// VisitImportValue implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportValue(ctx *api.ImportValueContext) interface{} {
	value := v.newExprWithTerminalNode(ctx.STRING())
	ret := &ImportValue{
		Value: value,
	}
	if ctx.ImportPackage() != nil {
		p := ctx.ImportPackage().Accept(v)
		pkg := p.(*ImportPackage)
		ret.As = pkg.As
		ret.Package = pkg.Package
	}

	return ret
}

// VisitImportPackage implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitImportPackage(ctx *api.ImportPackageContext) interface{} {
	return &ImportPackage{
		As:      v.newExprWithToken(ctx.GetAsToken()),
		Package: v.newExprWithToken(ctx.GetPackageName()),
	}
}

// Format provides a formatter for api command, now nothing to do
func (i *ImportExpr) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two ImportExpr are equal
func (i *ImportExpr) Equal(v interface{}) bool {
	if v == nil {
		return false
	}

	imp, ok := v.(*ImportExpr)
	if !ok {
		return false
	}

	if i.Package != nil {
		if !i.Package.Equal(imp.Package) {
			return false
		}
	}

	if !EqualDoc(i, imp) {
		return false
	}

	return i.Import.Equal(imp.Import) && i.Value.Equal(imp.Value)
}

// Doc returns the document of ImportExpr, like // some text
func (i *ImportExpr) Doc() []Expr {
	return i.DocExpr
}

// Comment returns the comment of ImportExpr, like // some text
func (i *ImportExpr) Comment() Expr {
	return i.CommentExpr
}
