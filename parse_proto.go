package main

import (
	"errors"
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"net/http"
	"strings"
)

type Service struct {
	Name     string
	Methods  []*Method
	Comments string
}

type Method struct {
	Name     string
	Comments string
	Method   string
	Path     string
	Request  string
	Reply    string
}

type ProtoItem struct {
	Services    []*Service
	FilePath    string
	PackageName string
	PackagePath string
}

func ServiceParser(srv *desc.ServiceDescriptor) *Service {
	srvInfo := &Service{
		Name:     srv.GetName(),
		Comments: srv.GetSourceInfo().GetLeadingComments(),
		Methods:  make([]*Method, 0),
	}

	for _, method := range srv.GetMethods() {
		MethodsParser(method, srvInfo)
	}
	return srvInfo
}

func adjustHttpPath(path string) string {
	path = strings.ReplaceAll(path, "{", ":")
	return strings.ReplaceAll(path, "}", "")
}

func GetHttpMethodAndPath(ext interface{}) (string, string, error) {
	switch rule := ext.(type) {
	case *annotations.HttpRule:
		switch httpRule := rule.GetPattern().(type) {
		case *annotations.HttpRule_Get:
			return http.MethodGet, adjustHttpPath(httpRule.Get), nil
		case *annotations.HttpRule_Post:
			return http.MethodPost, adjustHttpPath(httpRule.Post), nil
		case *annotations.HttpRule_Put:
			return http.MethodPut, adjustHttpPath(httpRule.Put), nil
		case *annotations.HttpRule_Delete:
			return http.MethodDelete, adjustHttpPath(httpRule.Delete), nil
		case *annotations.HttpRule_Patch:
			return http.MethodPatch, adjustHttpPath(httpRule.Patch), nil
		default:
			return "", "", errors.New("not http rule")
		}
	default:
		return "", "", errors.New("not http rule")
	}
}

func MethodsParser(method *desc.MethodDescriptor, srv *Service) {
	if method.IsServerStreaming() || method.IsClientStreaming() {
		fmt.Printf("method %s is not supported \n", method.GetName())
		return
	}

	ext := proto.GetExtension(method.GetMethodOptions(), annotations.E_Http)

	methodStr, path, err := GetHttpMethodAndPath(ext)
	if err != nil {
		fmt.Println(err)
		return
	}

	methodInfo := &Method{
		Name:     method.GetName(),
		Comments: method.GetSourceInfo().GetLeadingComments(),
		Method:   methodStr,
		Path:     path,
		Request:  method.GetInputType().GetName(),
		Reply:    method.GetOutputType().GetName(),
	}

	srv.Methods = append(srv.Methods, methodInfo)
}

func GetProtoItem(protoFileList []string, importPath []string) []*ProtoItem {
	pa := &protoparse.Parser{
		IncludeSourceCodeInfo: true,
		ImportPaths:           append(importPath, "./"),
	}

	protoItems := make([]*ProtoItem, 0)

	fds, err := pa.ParseFiles(protoFileList...)
	if err != nil {
		panic(err)
	}

	for _, fd := range fds {
		goPackage := fd.GetFileOptions().GetGoPackage()
		sl := strings.Split(goPackage, ";")
		if goPackage == "" || len(sl) != 2 {
			panic("go_package is not set")
		}

		protoItem := &ProtoItem{
			Services:    make([]*Service, 0),
			FilePath:    fd.GetName(),
			PackageName: sl[1],
			PackagePath: sl[0],
		}

		for _, srv := range fd.GetServices() {
			protoItem.Services = append(protoItem.Services, ServiceParser(srv))
		}

		protoItems = append(protoItems, protoItem)
	}
	return protoItems
}
