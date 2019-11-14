package multiremoteexec

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
)

func TestParseResource(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]interface{}
		want    []RemoteExec
		wantErr bool
	}{
		{
			name: "nil",
		},
		{
			name: "parse empty config",
			in:   map[string]interface{}{},
			want: nil,
		},
		{
			name: "parse type without value",
			in: map[string]interface{}{
				"remote_exec": []interface{}{
					map[string]interface{}{
						"type":                "inline",
						"continue_on_failure": true,
					},
					map[string]interface{}{
						"type": "scripts",
					},
				},
			},
			want: []RemoteExec{
				{
					Type:              RemoteExecInline,
					Values:            []string{},
					ContinueOnFailure: true,
				},
				{
					Type:   RemoteExecScripts,
					Values: []string{},
				},
			},
		},
		{
			name: "parse valid config",
			in: map[string]interface{}{
				"remote_exec": []interface{}{
					map[string]interface{}{
						"type": "inline",
						"values": []interface{}{
							"command_1",
							"other",
						},
						"continue_on_failure": true,
					},
					map[string]interface{}{
						"type": "scripts",
						"values": []interface{}{
							"./script1.ps1",
							"./other/script.sh",
						},
						"continue_on_failure": false,
					},
				},
			},
			want: []RemoteExec{
				{
					Type:              RemoteExecInline,
					Values:            []string{"command_1", "other"},
					ContinueOnFailure: true,
				},
				{
					Type:   RemoteExecScripts,
					Values: []string{"./script1.ps1", "./other/script.sh"},
				},
			},
		},
		{
			name: "parse valid config in order",
			in: map[string]interface{}{
				"remote_exec": []interface{}{
					map[string]interface{}{
						"type": "scripts",
						"values": []interface{}{
							"./script1.ps1",
							"./other/script.sh",
						},
						"continue_on_failure": false,
					},
					map[string]interface{}{
						"type": "inline",
						"values": []interface{}{
							"command_1",
							"other",
						},
						"continue_on_failure": true,
					},
				},
			},
			want: []RemoteExec{
				{
					Type:   RemoteExecScripts,
					Values: []string{"./script1.ps1", "./other/script.sh"},
				},
				{
					Type:              RemoteExecInline,
					Values:            []string{"command_1", "other"},
					ContinueOnFailure: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(schema.TestResourceDataRaw(t, Provisioner().(*schema.Provisioner).Schema, tt.in))
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
