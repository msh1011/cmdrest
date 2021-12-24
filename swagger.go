package cmdrest

import (
	"bytes"

	// embed used for template
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed swagger.tmpl
var swaggerTmpl string

type swagger struct {
	Title       string
	Description string
	Version     string
	Paths       map[string]pathDefs
}

type pathDefs map[string]operationDef

type pathDef map[string]operationDef

type operationDef struct {
	Summary     string
	Description string
	Tags        []string
	Params      []paramDef
}

type paramDef struct {
	Name        string
	Description string
	Type        string
	Default     string
	Required    bool
}

func generateSwagger(c *CmdHandler) (string, error) {
	t, err := template.New("swagger.tmpl").Parse(swaggerTmpl)
	if err != nil {
		return "", err
	}

	params := []paramDef{}
	for k, v := range c.defaultParams {
		p := paramDef{
			Name: k,
		}
		if val, ok := v.val.(string); ok {
			p.Default = fmt.Sprintf("%q", val)
		} else {
			p.Default = fmt.Sprintf("%v", v.val)
		}
		p.Type = fmt.Sprintf("%T", v.val)
		if p.Type == "bool" {
			p.Type = "bool"
		}
		params = append(params, p)
	}

	data := swagger{
		Title:       "Swagger",
		Description: "Swagger",
		Version:     "1.0.0",
		Paths: map[string]pathDefs{
			fmt.Sprintf("/%s/run", c.cmd.Name()): pathDefs{
				"get": operationDef{
					Description: "Get swagger",
					Params:      params,
				},
			},
		},
	}

	// create string io
	b := bytes.NewBuffer([]byte{})

	err = t.Execute(b, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
