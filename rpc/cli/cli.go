package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/zeromicro/goctl/rpc/generator"
	"github.com/urfave/cli"
)

// RPC is to generate rpc service code from a proto file by specifying a proto file using flag src,
// you can specify a target folder for code generation, when the proto file has import, you can specify
// the import search directory through the proto_path command, for specific usage, please refer to protoc -h
func RPC(c *cli.Context) error {
	src := c.String("src")
	out := c.String("dir")
	style := c.String("style")
	protoImportPath := c.StringSlice("proto_path")
	if len(src) == 0 {
		return errors.New("missing -src")
	}
	if len(out) == 0 {
		return errors.New("missing -dir")
	}

	g, err := generator.NewDefaultRPCGenerator(style)
	if err != nil {
		return err
	}

	return g.Generate(src, out, protoImportPath)
}

// RPCNew is to generate rpc greet service, this greet service can speed
// up your understanding of the zrpc service structure
func RPCNew(c *cli.Context) error {
	rpcname := c.Args().First()
	ext := filepath.Ext(rpcname)
	if len(ext) > 0 {
		return fmt.Errorf("unexpected ext: %s", ext)
	}
	style := c.String("style")

	protoName := rpcname + ".proto"
	filename := filepath.Join(".", rpcname, protoName)
	src, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = generator.ProtoTmpl(src)
	if err != nil {
		return err
	}

	g, err := generator.NewDefaultRPCGenerator(style)
	if err != nil {
		return err
	}

	return g.Generate(src, filepath.Dir(src), nil)
}

// RPCTemplate is the entry for generate rpc template
func RPCTemplate(c *cli.Context) error {
	protoFile := c.String("o")
	if len(protoFile) == 0 {
		return errors.New("missing -o")
	}

	return generator.ProtoTmpl(protoFile)
}
