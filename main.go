package main

import (
	"github.com/pb2go/common"
	"github.com/spf13/pflag"
)

func init() {
	var err error
	common.ModuleName, err = common.GetModulePath()
	if err != nil {
		panic(err)
	}
}

func main() {
	var paths []string
	var pathSlice = pflag.StringSlice("proto_path", nil, "Path to add")
	bizPath = pflag.String("bizPath", "./internal/biz", "bizPath default ./internal/biz")
	servicePath = pflag.String("servicePath", "./internal/service", "servicePath default ./internal/service")

	pflag.Parse()
	paths = *pathSlice
	files := pflag.Args()
	protoItems := GetProtoItem(files, paths)

	var genFileList []common.GenFile
	for _, gen := range generateList {
		genFileL, err := generate(protoItems, gen)
		if err != nil {
			panic(err)
		}
		genFileList = append(genFileList, genFileL...)
	}

	for _, genFile := range genFileList {
		err := common.SaveFile(genFile)
		if err != nil {
			panic(err)
		}
	}
}
