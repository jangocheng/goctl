package docgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/zeromicro/goctl/api/parser"
	"github.com/zeromicro/goctl/util"
)

// DocCommand generate markdown doc file
func DocCommand(c *cli.Context) error {
	dir := c.String("dir")
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	outputDir := c.String("o")
	if len(outputDir) == 0 {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if !util.FileExists(dir) {
		return fmt.Errorf("dir %s not exsit", dir)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	files, err := filePathWalkDir(dir)
	if err != nil {
		return err
	}

	for _, path := range files {
		api, err := parser.Parse(path)
		if err != nil {
			return fmt.Errorf("parse file: %s, err: %s", path, err.Error())
		}

		err = genDoc(api, filepath.Dir(filepath.Join(outputDir, path[len(dir):])),
			strings.Replace(path[len(filepath.Dir(path)):], ".api", ".md", 1))
		if err != nil {
			return err
		}
	}
	return nil
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".api") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
