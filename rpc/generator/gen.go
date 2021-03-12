package generator

import (
	"path/filepath"

	conf "github.com/zeromicro/goctl/config"
	"github.com/zeromicro/goctl/rpc/parser"
	"github.com/zeromicro/goctl/util"
	"github.com/zeromicro/goctl/util/console"
	"github.com/zeromicro/goctl/util/ctx"
)

// RPCGenerator defines a generator and configure
type RPCGenerator struct {
	g   Generator
	cfg *conf.Config
}

// NewDefaultRPCGenerator wraps Generator with configure
func NewDefaultRPCGenerator(style string) (*RPCGenerator, error) {
	cfg, err := conf.NewConfig(style)
	if err != nil {
		return nil, err
	}
	return NewRPCGenerator(NewDefaultGenerator(), cfg), nil
}

// NewRPCGenerator creates an instance for RPCGenerator
func NewRPCGenerator(g Generator, cfg *conf.Config) *RPCGenerator {
	return &RPCGenerator{
		g:   g,
		cfg: cfg,
	}
}

// Generate generates an rpc service, through the proto file,
// code storage directory, and proto import parameters to control
// the source file and target location of the rpc service that needs to be generated
func (g *RPCGenerator) Generate(src, target string, protoImportPath []string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	err = util.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	err = g.g.Prepare()
	if err != nil {
		return err
	}

	projectCtx, err := ctx.Prepare(abs)
	if err != nil {
		return err
	}

	p := parser.NewDefaultProtoParser()
	proto, err := p.Parse(src)
	if err != nil {
		return err
	}

	dirCtx, err := mkdir(projectCtx, proto)
	if err != nil {
		return err
	}

	err = g.g.GenEtc(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenPb(dirCtx, protoImportPath, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenConfig(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenSvc(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenLogic(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenServer(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenMain(dirCtx, proto, g.cfg)
	if err != nil {
		return err
	}

	err = g.g.GenCall(dirCtx, proto, g.cfg)

	console.NewColorConsole().MarkDone()

	return err
}
