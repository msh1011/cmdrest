package cmdrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/ionrock/procs"
	"github.com/robertbakker/swaggerui"
)

type Cmd interface {
	Name() string
}

// CmdHandler implements the http.Handler interface for the given Cmd
type CmdHandler struct {
	cmd           Cmd
	swag          string
	defaultParams map[string]param
}

type param struct {
	name string
	flag string
	pos  int
	val  interface{}
}

// CreateNewHandler creates a new CmdHandler for the given Cmd
func CreateNewHandler(c Cmd) (*CmdHandler, error) {
	h := &CmdHandler{
		defaultParams: getParamMap(c),
		cmd:           c,
	}
	var err error

	h.swag, err = generateSwagger(h)
	if err != nil {
		return nil, err
	}

	return h, nil
}

type resp struct {
	ExitCode int      `json:"exit_code,omitempty"`
	Stdout   []string `json:"stdout"`
	Stderr   []string `json:"stderr,omitempty"`
	Error    error    `json:"error,omitempty"`
	Cmd      string   `json:"cmd"`
}

func (c *CmdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/docs/swagger.yaml") {
		c.handleYaml(w, r)
	} else if strings.Contains(r.URL.Path, "/docs/") {
		c.handleDocs(w, r)
	} else if strings.Contains(r.URL.Path, "/docs") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
	} else if strings.Contains(r.URL.Path, "/run") {
		c.handleCmd(w, r)
	}
}

func (c *CmdHandler) handleDocs(w http.ResponseWriter, r *http.Request) {
	i := strings.LastIndex(r.URL.Path, "/docs")
	if i == -1 {
		return
	}
	pref := r.URL.Path[:i+len("/docs")]
	swaggerfile := r.URL.Path + "/swagger.yaml"
	if strings.Contains(swaggerfile, "docs//") {
		swaggerfile = r.URL.Path + "swagger.yaml"
	}
	http.StripPrefix(pref, swaggerui.SwaggerURLHandler(swaggerfile)).ServeHTTP(w, r)
}

func (c *CmdHandler) handleYaml(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(c.swag))
}

// ServerHTTP implements the http.Handler interface for the REST service.
func (c *CmdHandler) handleCmd(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	cmdParams := c.params()
	posLen := 0

	for k, v := range cmdParams {
		if val := query.Get(k); val != "" {
			if bval, err := strconv.ParseBool(val); err == nil {
				cmdParams[k].val = bval
			} else {
				cmdParams[k].val = val
			}
		}
		if v.pos >= 0 {
			posLen++
		}
	}

	cmdStr := c.cmd.Name()
	posArgs := make([]string, posLen)

	for _, v := range cmdParams {
		if v.pos >= 0 {
			posArgs[v.pos] = fmt.Sprintf("%v", v.val)
			continue
		}
		if val, ok := v.val.(bool); ok {
			if val {
				cmdStr += " -" + v.flag
			} else {
				continue
			}
		} else {
			cmdStr += fmt.Sprintf(" --%v %v", v.flag, v.val)
		}
	}

	for _, v := range posArgs {
		cmdStr += " " + v
	}

	resp := resp{
		Stdout: []string{},
		Stderr: []string{},
		Cmd:    cmdStr,
	}

	defer resp.Render(w)

	p := procs.Process{
		CmdString: cmdStr,
		OutputHandler: func(line string) string {
			resp.Stdout = append(resp.Stdout, line)
			return ""
		},
		ErrHandler: func(line string) string {
			resp.Stderr = append(resp.Stderr, line)
			return ""
		},
	}

	err := p.Start()
	if err != nil {
		resp.Error = err
		return
	}

	err = p.Wait()
	if err != nil {
		resp.Error = err
		return
	}

	return
}

func (r *resp) Render(w http.ResponseWriter) {
	if r.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if val, ok := r.Error.(*exec.ExitError); ok {
			r.ExitCode = val.ExitCode()
		}
	}
	json.NewEncoder(w).Encode(r)
}

func (c *CmdHandler) params() map[string]*param {
	ret := make(map[string]*param)
	for k, v := range c.defaultParams {
		ret[k] = &param{name: v.name, val: v.val, flag: v.flag, pos: v.pos}
	}
	return ret
}

func getParamMap(cmd Cmd) map[string]param {
	params := make(map[string]param)
	// get all struct fields with tag "rest"
	def := reflect.ValueOf(cmd).Elem()
	typeOfT := def.Type()
	for i := 0; i < def.NumField(); i++ {
		f := typeOfT.Field(i)
		// get tag
		tag := f.Tag.Get("rcmd")
		// check if tag is empty
		if tag == "" {
			continue
		}
		var pos int
		if val, err := strconv.Atoi(tag); err == nil {
			pos = val
		} else {
			pos = -1
		}
		// get param name
		name := f.Name
		// check if param name is empty
		if name == "" {
			continue
		}
		// get param value
		value := def.Field(i).Interface()
		// add param to map
		params[name] = param{name: name, val: value, flag: tag, pos: pos}
	}
	return params
}
