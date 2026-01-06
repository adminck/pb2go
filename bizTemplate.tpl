package {{ .PackageName }}

import (
	"context"
    pb "{{ .PbPackage }}"
    . "{{ .ModuleName }}/internal/data"
)

type {{ .StructName }} struct {
    tr Transaction
}

var {{ .VarStructName }} *{{ .StructName }}

func New{{ .VarStructName }} (tr Transaction) *{{ .StructName }} {
    if {{ .VarStructName }} == nil {
        {{ .VarStructName }} = &{{ .StructName }}{tr}
    }
    return {{ .VarStructName }}
}