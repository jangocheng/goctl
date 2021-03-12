package upgrade

import (
	"fmt"

	"github.com/zeromicro/goctl/rpc/execx"
	"github.com/urfave/cli"
)

// Upgrade gets the latest goctl by
// go get -u github.com/zeromicro/goctl
func Upgrade(_ *cli.Context) error {
	info, err := execx.Run("GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u github.com/zeromicro/goctl", "")
	if err != nil {
		return err
	}

	fmt.Print(info)
	return nil
}
