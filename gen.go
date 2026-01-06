package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/pb2go/common"
	"go/ast"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed bizTemplate.tpl
var bizTemplate string

//go:embed serviceTemplate.tpl
var serviceTemplate string

var (
	bizPath     *string = nil
	servicePath *string = nil
)

type Item struct {
	PackageName   string
	PbPackage     string
	ModuleName    string
	StructName    string
	VarStructName string
	UCModuleName  string
	UCName        string

	FileName string
	FilePath string
}

func (b *Item) GetAstFile(tmpl *template.Template) (*ast.File, error) {
	ok, err := common.PathExists(b.FilePath)
	if err != nil {
		return nil, err
	}

	if !ok {
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, b); err != nil {
			panic(err)
		}
		return common.ParseFile("", strings.Trim(buf.String(), "\r\n"))
	}

	return common.ParseFile(b.FilePath, nil)
}
func NewBizItem(item *ProtoItem, serviceName string) *Item {
	i := &Item{
		ModuleName:  common.ModuleName,
		PackageName: item.PackageName,
	}

	normalizedPath := filepath.ToSlash(item.FilePath)
	i.FileName = common.ToSnakeCase(serviceName)
	i.StructName = fmt.Sprintf("_%sUc", common.ToFirstLower(strcase.ToCamel(i.FileName)))
	i.VarStructName = fmt.Sprintf("%sUc", strcase.ToCamel(i.FileName))
	sl := strings.Split(normalizedPath, "/")
	i.PbPackage = fmt.Sprintf("%s/%s", i.ModuleName, strings.Join(sl[0:len(sl)-1], "/"))
	i.FilePath = fmt.Sprintf("%s/%s/%s.go", *bizPath, strings.Join(append(sl[1:len(sl)-1], i.PackageName), "/"), i.FileName)

	return i
}
func NewServiceItem(item *ProtoItem, serviceName string) *Item {
	i := &Item{
		ModuleName:  common.ModuleName,
		PackageName: item.PackageName,
	}

	normalizedPath := filepath.ToSlash(item.FilePath)
	i.FileName = common.ToSnakeCase(serviceName)
	i.StructName = fmt.Sprintf("_%s", common.ToFirstLower(strcase.ToCamel(i.FileName)))
	i.VarStructName = strcase.ToCamel(i.FileName)
	sl := strings.Split(normalizedPath, "/")
	i.PbPackage = fmt.Sprintf("%s/%s", i.ModuleName, strings.Join(sl[0:len(sl)-1], "/"))
	i.FilePath = fmt.Sprintf("%s/%s/%s.go", *servicePath, strings.Join(append(sl[1:len(sl)-1], i.PackageName), "/"), i.FileName)
	i.UCName = fmt.Sprintf("%sUc", i.VarStructName)

	bizP := path.Dir(filepath.ToSlash(*bizPath + "/biz.go")) // biz包路径

	i.UCModuleName = fmt.Sprintf("%s/%s", bizP, strings.Join(append(sl[1:len(sl)-1], i.PackageName), "/"))

	return i
}

type genType struct {
	Tmpl    *template.Template
	getItem func(item *ProtoItem, serviceName string) *Item
}

var generateList = []genType{
	{
		Tmpl:    template.Must(template.New("biz").Parse(strings.TrimSpace(bizTemplate))),
		getItem: NewBizItem,
	},
	{
		Tmpl:    template.Must(template.New("service").Parse(strings.TrimSpace(serviceTemplate))),
		getItem: NewServiceItem,
	},
}

func generate(protoItem []*ProtoItem, g genType) (genFileList []common.GenFile, err error) {
	for _, item := range protoItem {
		for _, s := range item.Services {
			i := g.getItem(item, s.Name)
			serviceAst, err := i.GetAstFile(g.Tmpl)
			if err != nil {
				return nil, err
			}

			funcMap := make(map[string]*ast.FuncDecl)
			var receiveField = &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent(common.GetOneChar(i.FileName))},
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: i.StructName, Obj: serviceAst.Scope.Objects[i.StructName]},
						},
					},
				},
			}

			for _, decl := range serviceAst.Decls {
				f, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if common.IsMethodOf(f, i.StructName) {
					funcMap[f.Name.Name] = f
				}
			}

			for _, method := range s.Methods {
				_, ok := funcMap[method.Name]
				if ok {
					continue
				}
				ucName := common.GetOneChar(i.FileName)
				if i.UCName != "" {
					ucName = i.UCName
				}
				serviceAst.Decls = append(serviceAst.Decls, common.NewFunc(common.NewFuncParams{
					FnName:       method.Name,
					ReqName:      fmt.Sprintf("pb.%s", method.Request),
					ReplyName:    fmt.Sprintf("pb.%s", method.Reply),
					UcName:       ucName,
					Comment:      method.Comments,
					ReceiveField: receiveField,
				}))
			}

			genFileList = append(genFileList, common.GenFile{
				AstFile:  serviceAst,
				FilePath: i.FilePath,
			})
		}
	}
	return genFileList, nil
}
