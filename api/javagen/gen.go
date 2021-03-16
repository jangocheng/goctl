package javagen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/urfave/cli"
	"github.com/zeromicro/goctl/api/parser"
	"github.com/zeromicro/goctl/util"
)

// JavaCommand the generate java code command entrance
func JavaCommand(c *cli.Context) error {
	apiParam := c.String("api")
	dir := c.String("dir")
	if len(apiParam) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	apiPath, importMap, err := util.ParseApiParam(apiParam)
	if err != nil {
		return err
	}

	api, err := parser.Parse(apiPath)
	if err != nil {
		return err
	}

	packetName := api.Service.Name
	if strings.HasSuffix(packetName, "-api") {
		packetName = packetName[:len(packetName)-4]
	}

	logx.Must(util.MkdirIfNotExist(dir))
	logx.Must(genPacket(dir, packetName, api))
	logx.Must(genComponents(dir, packetName, api, importMap))

	fmt.Println(aurora.Green("Done."))
	return nil
}
