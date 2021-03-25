package generator

import (
	"bytes"
	"fmt"
	"github.com/tal-tech/go-zero/core/collection"
	conf "github.com/zeromicro/goctl/config"
	"github.com/zeromicro/goctl/rpc/execx"
	"github.com/zeromicro/goctl/rpc/parser"
	"os"
	"path/filepath"
	"strings"
)

// GenPb generates the pb.go file, which is a layer of packaging for protoc to generate gprc,
// but the commands and flags in protoc are not completely joined in goctl. At present, proto_path(-I) is introduced
func (g *DefaultGenerator) GenPb(ctx DirContext, protoImportPath []string, proto parser.Proto, _ *conf.Config, goOptions ...string) error {
	dir := ctx.GetPb()
	cw := new(bytes.Buffer)
	directory, base := filepath.Split(proto.Src)
	cw.WriteString("protoc ")
	protoImportPathSet := collection.NewSet()
	for _, ip := range protoImportPath {
		pip := " --proto_path=" + ip
		if protoImportPathSet.Contains(pip) {
			continue
		}

		protoImportPathSet.AddStr(pip)
		cw.WriteString(pip)
	}
	currentPath := " --proto_path=" + directory
	if !protoImportPathSet.Contains(currentPath) {
		cw.WriteString(currentPath)
	}
	cw.WriteString(" " + proto.Name)
	if strings.Contains(proto.GoPackage, "/") {
		cw.WriteString(" --go_out=plugins=grpc:" + ctx.GetMain().Filename)
	} else {
		cw.WriteString(" --go_out=plugins=grpc:" + dir.Filename)
	}

	// Compatible with version 1.4.0，github.com/golang/protobuf/protoc-gen-go@v1.4.0
	// --go_out usage please see https://developers.google.com/protocol-buffers/docs/reference/go-generated#package
	optSet := collection.NewSet()
	for _, op := range goOptions {
		opt := " --go_opt=" + op
		if optSet.Contains(opt) {
			continue
		}

		optSet.AddStr(op)
		cw.WriteString(" --go_opt=" + op)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	ret, err := filepath.Rel(workDir, dir.Filename)
	if err != nil {
		return err
	}

	fmt.Println(ctx.GetMain().Filename)
	fmt.Println(dir.Filename)
	fmt.Println(ret)
	currentFileOpt := " --go_opt=M" + base + "=../" + proto.GoPackage
	if !optSet.Contains(currentFileOpt) {
		cw.WriteString(currentFileOpt)
	}

	command := cw.String()
	g.log.Debug(command)
	_, err = execx.Run(command, "")
	return err
}
