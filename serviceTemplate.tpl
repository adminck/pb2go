package {{ .PackageName }}

import (
    "context"
    pb "{{ .PbPackage }}"
    . "{{ .ModuleName }}/{{ .UCModuleName }}"
)

type {{ .StructName }} struct {
    pb.Unimplemented{{ .VarStructName }}Server
}

var {{ .VarStructName }} = new({{ .StructName }})