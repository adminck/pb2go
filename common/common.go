package common

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var ModuleName = "github.com/adminck/go-multitenant"

type GenFile struct {
	AstFile  *ast.File
	FilePath string
}

func GetModulePath() (string, error) {
	goModPath := "./go.mod"

	// 读取 go.mod 文件第一行
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("open go.mod file error: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return "", fmt.Errorf("read go.mod file error")
	}
	firstLine := strings.TrimSpace(scanner.Text())

	// 解析模块路径 （module <path>）
	modulePath := strings.SplitN(firstLine, " ", 3)
	if len(modulePath) != 2 || modulePath[0] != "module" {
		return "", fmt.Errorf("invalid go.mod file format")
	}
	return modulePath[1], nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func SaveFile(file GenFile) error {
	dir := filepath.Dir(file.FilePath)
	// 创建目录
	if exists, _ := PathExists(dir); !exists {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.Create(file.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return format.Node(f, fset, file.AstFile)
}

func ToSnakeCase(s string) string {
	var res []rune
	for i, c := range []rune(s) {
		if unicode.IsUpper(c) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				res = append(res, '_')
			}
			res = append(res, unicode.ToLower(c))
		} else {
			res = append(res, c)
		}
	}
	return string(res)
}

func ToFirstLower(s string) string {
	if len(s) == 0 {
		return s
	}
	res := []rune(s)
	res[0] = unicode.ToLower(res[0])
	return string(res)
}

func GetOneChar(s string) string {
	if len(s) == 0 {
		return s
	}
	res := []rune(s)
	return string(res[0])
}

func IsMethodOf(f *ast.FuncDecl, s string) bool {
	// 检查是否有接收者
	if f.Recv == nil || len(f.Recv.List) == 0 {
		return false
	}

	// 提取接收者的类型
	recvType := f.Recv.List[0].Type

	// 将类型转换为字符串
	recvTypeStr := TypeToString(recvType)

	return recvTypeStr == s || recvTypeStr == "*"+s
}

func TypeToString(expr ast.Expr) string {
	var buf strings.Builder
	format.Node(&buf, fset, expr)
	return buf.String()
}
