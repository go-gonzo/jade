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

type Options struct {
	FuncMap map[string]interface{}
	Data    interface{}
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
				t, err := template.New("html").Parse(html)
				if err != nil {
					return err
				}

				if opt.FuncMap != nil {
					t = t.Funcs(opt.FuncMap)
				}
				t.Execute(content, opt.Data)

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
