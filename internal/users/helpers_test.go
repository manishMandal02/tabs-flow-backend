package users

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseSubPreferencesData(t *testing.T) {

	tests := []struct {
		name        string
		perfBody    updatePerfBody
		wantSK      string
		wantSubPerf interface{}
		wantErr     bool
	}{{
		name: "valid data",
		perfBody: updatePerfBody{
			Type: "General",
			Data: json.RawMessage(`
					{
					"openSpace": "sameWindow"
					}
				`),
		},
		wantSK: "P#General",
		wantSubPerf: &generalP{
			OpenSpace: "sameWindow",
		},
		wantErr: false,
	},
		{
			name: "invalid data",
			perfBody: updatePerfBody{
				Type: "General1",
				Data: json.RawMessage(`
					{
					"openSpace": "sameWindow"
					}
				`),
			},
			wantSK:      "",
			wantSubPerf: nil,
			wantErr:     true,
		},
		{
			name: "empty data",
			perfBody: updatePerfBody{
				Type: "General",
				Data: json.RawMessage(``),
			},
			wantSK:      "",
			wantSubPerf: nil,
			wantErr:     true,
		},
		{
			name: "invalid preference type",
			perfBody: updatePerfBody{
				Type: "Test",
				Data: json.RawMessage(`
					{
					"openSpace": "sameWindow"
					}
				`),
			},
			wantSK:      "",
			wantSubPerf: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseSubPreferencesData(tt.perfBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("err \n[actual] = %v, \n[want]= %v", err, tt.wantErr)
				return
			}
			if got != tt.wantSK {
				t.Errorf("wantSK  \n[actual] = %v, \n[want]= %v", got, tt.wantSK)
			}

			if reflect.TypeOf(got1) == reflect.TypeOf(tt.wantSubPerf) {
				t.Errorf("wantSubPerf \n[actual]  = %v, \n[want]= %v", got1, tt.wantSubPerf)
			}

		})
	}
}
