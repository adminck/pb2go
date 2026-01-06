package common

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var fset = token.NewFileSet()

func ParseFile(fildePath string, src any) (*ast.File, error) {
	if fildePath != "" {
		return parser.ParseFile(fset, fildePath, nil, parser.ParseComments)
	}
	return parser.ParseFile(fset, "", src, parser.ParseComments)
}

type NewFuncParams struct {
	FnName       string         // 函数名
	ReqName      string         // 请求参数名
	ReplyName    string         // 响应参数名
	UcName       string         // uc名
	Comment      string         // 注释
	ReceiveField *ast.FieldList // 接收参数
}

func NewFunc(param NewFuncParams) *ast.FuncDecl {
	fn := &ast.FuncDecl{
		Name: ast.NewIdent(param.FnName),
		Recv: param.ReceiveField,
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("ctx"),
						},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent("context"),
							Sel: ast.NewIdent("Context"),
						},
					},
					{
						Names: []*ast.Ident{
							ast.NewIdent("req"),
						},
						Type: &ast.StarExpr{
							X: ast.NewIdent(param.ReqName),
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: ast.NewIdent(param.ReplyName),
						},
					},
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}

	if param.Comment != "" {
		fn.Doc.List = append(fn.Doc.List, &ast.Comment{
			Text: fmt.Sprintf("\n//%s %s \r\n ", param.FnName, asComment(param.Comment)),
		})
	}

	var args []ast.Expr
	for _, arg := range fn.Type.Params.List {
		if len(arg.Names) > 0 {
			args = append(args, arg.Names[0])
		} else {
			// 匿名参数
			args = append(args, ast.NewIdent("_"))
		}
	}

	var fun ast.Expr
	if param.UcName == "" {
		fun = ast.NewIdent(param.FnName)
	} else {
		fun = &ast.SelectorExpr{
			X:   ast.NewIdent(param.UcName),
			Sel: ast.NewIdent(param.FnName),
		}
	}

	fn.Body.List = append(fn.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun:  fun,
				Args: args,
			},
		},
	})

	return fn
}

func asComment(c string) string {
	runes := []rune(c)
	if len(runes) == 0 {
		return ""
	}

	str := bytes.Buffer{}

	for i, r := range runes {
		str.WriteString(string(r))
		if r == '\n' && i != len(runes)-1 {
			str.WriteString("//")
		}
	}

	return str.String()
}
