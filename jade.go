package jade

// github.com/Joker/jade binding for gonzo.

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/omeid/gonzo/context"

	"github.com/Joker/jade"
	"github.com/omeid/gonzo"
)

type Delims struct {
	Right string
	Left  string
}

type Options struct {
	FuncMap map[string]interface{}
	Data    interface{}
	Delims  Delims
}

func Compile(opt Options) gonzo.Stage {
	return func(ctx context.Context, in <-chan gonzo.File, out chan<- gonzo.File) error {

		for {
			select {
			case file, ok := <-in:
				if !ok {
					return nil
				}

				src, err := ioutil.ReadAll(file)
				if err != nil {
					return err
				}

				name := strings.TrimSuffix(file.FileInfo().Name(), ".jade") + ".html"
				ctx.Infof("Compiling %s to %s", file.FileInfo().Name(), name)
				html, err := jade.Parse(name, string(src))
				if err != nil {
					return err
				}

				content := new(bytes.Buffer)
				t := template.New("html")

				if opt.FuncMap != nil {
					t = t.Funcs(opt.FuncMap)
				}
				if opt.Delims.Right != "" && opt.Delims.Left != "" {
					t = t.Delims(opt.Delims.Left, opt.Delims.Right)
				}

				t, err = t.Parse(html)
				if err != nil {
					return err
				}

				err = t.Execute(content, opt.Data)
				if err != nil {
					return err
				}

				file = gonzo.NewFile(ioutil.NopCloser(content), file.FileInfo())
				file.FileInfo().SetSize(int64(content.Len()))
				file.FileInfo().SetName(name)

				out <- file
			case <-ctx.Done():
				return nil
			}
		}
	}
}
