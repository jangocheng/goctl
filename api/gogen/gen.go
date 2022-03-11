package gogen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
	"github.com/zeromicro/go-zero/core/logx"
	apiformat "github.com/zeromicro/goctl/api/format"
	"github.com/zeromicro/goctl/api/parser"
	apiutil "github.com/zeromicro/goctl/api/util"
	"github.com/zeromicro/goctl/config"
	"github.com/zeromicro/goctl/util"
	"github.com/zeromicro/goctl/util/console"
)

const tmpFile = "%s-%d"

var tmpDir = path.Join(os.TempDir(), "goctl")

// GoCommand gen go project files from command line
func GoCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	namingStyle := c.String("style")
	onlyType := c.Bool("types")

	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	if onlyType {
		return DoGenTypes(apiFile, dir, namingStyle)
	}

	return DoGenProject(apiFile, dir, namingStyle)
}

// DoGenTypes generates golang types from the specified api
func DoGenTypes(apiFile, dir, style string) error {
	apiPath, _, err := util.ParseApiParam(apiFile)
	if err != nil {
		return err
	}

	api, err := parser.Parse(apiPath)
	if err != nil {
		return err
	}

	if len(api.Imports) > 0 {
		return errors.New("the reused type structure file is not allowed to import other files")
	}

	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	err = util.MkdirIfNotExist(dir)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, filepath.Base(apiFile))
	err = onlyGenTypes(filename, cfg, api)
	c := console.NewColorConsole()
	if err == nil {
		c.MarkDone()
	}

	return err
}

// DoGenProject gen go project files with api file
func DoGenProject(apiParam, dir, style string) error {
	apiPath, importMap, err := util.ParseApiParam(apiParam)
	if err != nil {
		return err
	}

	api, err := parser.Parse(apiPath)
	if err != nil {
		return err
	}

	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	logx.Must(util.MkdirIfNotExist(dir))
	logx.Must(genEtc(dir, cfg, api))
	logx.Must(genConfig(dir, cfg, api))
	logx.Must(genMain(dir, cfg, api))
	logx.Must(genServiceContext(dir, cfg, api))
	logx.Must(genTypes(dir, importMap, cfg, api))
	logx.Must(genRoutes(dir, cfg, api))
	logx.Must(genHandlers(dir, cfg, api))
	logx.Must(genLogic(dir, cfg, api))
	logx.Must(genMiddleware(dir, cfg, api))

	if err := backupAndSweep(apiPath); err != nil {
		return err
	}

	if err := apiformat.ApiFormatByPath(apiPath); err != nil {
		return err
	}

	fmt.Println(aurora.Green("Done."))
	return nil
}

func backupAndSweep(apiFile string) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)
	_ = os.MkdirAll(tmpDir, os.ModePerm)

	go func() {
		_, fileName := filepath.Split(apiFile)
		_, e := apiutil.Copy(apiFile, fmt.Sprintf(path.Join(tmpDir, tmpFile), fileName, time.Now().Unix()))
		if e != nil {
			err = e
		}
		wg.Done()
	}()
	go func() {
		if e := sweep(); e != nil {
			err = e
		}
		wg.Done()
	}()
	wg.Wait()

	return err
}

func sweep() error {
	keepTime := time.Now().AddDate(0, 0, -7)
	return filepath.Walk(tmpDir, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pos := strings.LastIndexByte(info.Name(), '-')
		if pos > 0 {
			timestamp := info.Name()[pos+1:]
			seconds, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				// print error and ignore
				fmt.Println(aurora.Red(fmt.Sprintf("sweep ignored file: %s", fpath)))
				return nil
			}

			tm := time.Unix(seconds, 0)
			if tm.Before(keepTime) {
				if err := os.Remove(fpath); err != nil {
					fmt.Println(aurora.Red(fmt.Sprintf("failed to remove file: %s", fpath)))
					return err
				}
			}
		}

		return nil
	})
}
