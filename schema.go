package multiremoteexec

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

type RemoteExecType string

const (
	RemoteExecInline  RemoteExecType = "inline"
	RemoteExecScripts RemoteExecType = "scripts"
)

type RemoteExec struct {
	Type RemoteExecType `json:"type"`
	// Powershell        bool            `json:"powershell"`
	Values            []string        `json:"values"`
	ContinueOnFailure bool            `json:"continue_on_failure"`
	Scripts           []io.ReadCloser `json:"-"`
}

func (e *RemoteExec) Collect() error {
	switch e.Type {
	case RemoteExecInline:
		scripts := generateScripts(e.Values)
		for _, script := range scripts {
			e.Scripts = append(e.Scripts, ioutil.NopCloser(bytes.NewReader([]byte(script))))
		}
	case RemoteExecScripts:
		for _, s := range e.Values {
			fh, err := os.Open(s)
			if err != nil {
				for _, fh := range e.Scripts {
					fh.Close()
				}
				return fmt.Errorf("failed to open script '%s': %v", s, err)
			}
			e.Scripts = append(e.Scripts, fh)
		}
	}
	return nil
}

var (
	remoteExecSchema = &schema.Schema{
		Type:     schema.TypeList,
		Elem:     &schema.Resource{Schema: baseResources},
		Optional: true,
	}

	typeSchema = &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"inline", "scripts"}, true),
		Required:     true,
	}

	valuesSchema = &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
	}

	continueOnFailureSchema = &schema.Schema{
		Type:     schema.TypeBool,
		Default:  false,
		Optional: true,
	}

	windowsRestartSchema = &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
	}

	powershellSchema = &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	}

	baseResources = map[string]*schema.Schema{
		"type":   typeSchema,
		"values": valuesSchema,
		// "powershell": powershellSchema,
		// "windows_restart": windowsRestartSchema,
		"continue_on_failure": continueOnFailureSchema,
	}
)

func generateScripts(lines []string) []string {
	lines = append(lines, "")
	return []string{strings.Join(lines, "\n")}
}
