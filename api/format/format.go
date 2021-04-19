package format

import (
	"bufio"
	"errors"
	"fmt"
	"go/format"
	"go/scanner"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tal-tech/go-zero/core/errorx"
	"github.com/urfave/cli"
	"github.com/zeromicro/goctl/api/parser"
	"github.com/zeromicro/goctl/api/util"
	ctlutil "github.com/zeromicro/goctl/util"
)

const (
	leftParenthesis  = "("
	rightParenthesis = ")"
	leftBrace        = "{"
	rightBrace       = "}"
)

// GoFormatApi format api file
func GoFormatApi(c *cli.Context) error {
	useStdin := c.Bool("stdin")

	var be errorx.BatchError
	if useStdin {
		if err := apiFormatByStdin(); err != nil {
			be.Add(err)
		}
	} else {
		dir := c.String("dir")
		if len(dir) == 0 {
			return errors.New("missing -dir")
		}

		_, err := os.Lstat(dir)
		if err != nil {
			return errors.New(dir + ": No such file or directory")
		}

		err = filepath.Walk(dir, func(path string, fi os.FileInfo, errBack error) (err error) {
			if strings.HasSuffix(path, ".api") {
				if err := ApiFormatByPath(path); err != nil {
					be.Add(util.WrapErr(err, fi.Name()))
				}
			}
			return nil
		})
		be.Add(err)
	}

	if be.NotNil() {
		scanner.PrintError(os.Stderr, be.Err())
		os.Exit(1)
	}

	return be.Err()
}

func apiFormatByStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	result, err := apiFormat(string(data), pwd)
	if err != nil {
		return err
	}

	_, err = fmt.Print(result)
	return err
}

// ApiFormatByPath format api from file path
func ApiFormatByPath(apiFilePath string) error {
	data, err := ioutil.ReadFile(apiFilePath)
	if err != nil {
		return err
	}

	result, err := apiFormat(string(data), filepath.Dir(apiFilePath))
	if err != nil {
		return err
	}

	_, err = parser.ParseContent(result, filepath.Dir(apiFilePath))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(apiFilePath, []byte(result), os.ModePerm)
}

func apiFormat(data, workDir string) (string, error) {
	_, err := parser.ParseContent(data, workDir)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	s := bufio.NewScanner(strings.NewReader(data))
	var tapCount = 0
	var newLineCount = 0
	var preLine string
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 {
			if newLineCount > 0 {
				continue
			}
			newLineCount++
		} else {
			if preLine == rightBrace {
				builder.WriteString(ctlutil.NL)
			}
			newLineCount = 0
		}

		if tapCount == 0 {
			format, err := formatGoTypeDef(line, s, &builder)
			if err != nil {
				return "", err
			}

			if format {
				continue
			}
		}

		noCommentLine := util.RemoveComment(line)
		if noCommentLine == rightParenthesis || noCommentLine == rightBrace {
			tapCount--
		}
		if tapCount < 0 {
			line := strings.TrimSuffix(noCommentLine, rightBrace)
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, leftBrace) {
				tapCount++
			}
		}
		util.WriteIndent(&builder, tapCount)
		builder.WriteString(line + ctlutil.NL)
		if strings.HasSuffix(noCommentLine, leftParenthesis) || strings.HasSuffix(noCommentLine, leftBrace) {
			tapCount++
		}
		preLine = line
	}

	return strings.TrimSpace(builder.String()), nil
}

func formatGoTypeDef(line string, scanner *bufio.Scanner, builder *strings.Builder) (bool, error) {
	noCommentLine := util.RemoveComment(line)
	tokenCount := 0
	if strings.HasPrefix(noCommentLine, "type") && (strings.HasSuffix(noCommentLine, leftParenthesis) ||
		strings.HasSuffix(noCommentLine, leftBrace)) {
		var typeBuilder strings.Builder
		typeBuilder.WriteString(mayInsertStructKeyword(line, &tokenCount) + ctlutil.NL)
		for scanner.Scan() {
			noCommentLine := util.RemoveComment(scanner.Text())
			typeBuilder.WriteString(mayInsertStructKeyword(scanner.Text(), &tokenCount) + ctlutil.NL)
			if noCommentLine == rightBrace || noCommentLine == rightParenthesis {
				tokenCount--
			}
			if tokenCount == 0 {
				ts, err := format.Source([]byte(typeBuilder.String()))
				if err != nil {
					return false, errors.New("error format \n" + typeBuilder.String())
				}

				result := strings.ReplaceAll(string(ts), " struct ", " ")
				result = strings.ReplaceAll(result, "type ()", "")
				builder.WriteString(result)
				break
			}
		}
		return true, nil
	}

	return false, nil
}

func mayInsertStructKeyword(line string, token *int) string {
	insertStruct := func() string {
		if strings.Contains(line, " struct") {
			return line
		}
		index := strings.Index(line, leftBrace)
		return line[:index] + " struct " + line[index:]
	}

	noCommentLine := util.RemoveComment(line)
	if strings.HasSuffix(noCommentLine, leftBrace) {
		*token++
		return insertStruct()
	}
	if strings.HasSuffix(noCommentLine, rightBrace) {
		noCommentLine = strings.TrimSuffix(noCommentLine, rightBrace)
		noCommentLine = util.RemoveComment(noCommentLine)
		if strings.HasSuffix(noCommentLine, leftBrace) {
			return insertStruct()
		}
	}
	if strings.HasSuffix(noCommentLine, leftParenthesis) {
		*token++
	}

	if strings.Contains(noCommentLine, "`") {
		return util.UpperFirst(strings.TrimSpace(line))
	}

	return line
}
