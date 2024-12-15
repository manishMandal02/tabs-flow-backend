package users

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// handler helper
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
			actual1, actual2, err := parseSubPreferencesData(tt.perfBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("err \n[actual] = %v, \n[want]= %v", err, tt.wantErr)
				return
			}
			if actual1 != tt.wantSK {
				t.Errorf("wantSK  \n[actual] = %v, \n[want]= %v", actual1, tt.wantSK)
			}

			if reflect.TypeOf(actual2) == reflect.TypeOf(tt.wantSubPerf) {
				t.Errorf("wantSubPerf \n[actual]  = %v, \n[want]= %v", actual2, tt.wantSubPerf)
			}

		})
	}
}

// repository helper @ repository.go:415
func TestUnMarshalPreferences(t *testing.T) {
	tests := []struct {
		name    string
		input   *dynamodb.QueryOutput
		want    *Preferences
		wantErr bool
	}{
		{
			name: "invalid query item, no SK",
			input: &dynamodb.QueryOutput{
				Items: []map[string]types.AttributeValue{
					{
						"PK": &types.AttributeValueMemberS{Value: "123"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid SK, not a sub preference type",
			input: &dynamodb.QueryOutput{
				Items: []map[string]types.AttributeValue{
					{
						"PK": &types.AttributeValueMemberS{Value: "123"},
						"SK": &types.AttributeValueMemberS{Value: "P#Theme"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			input: &dynamodb.QueryOutput{
				Items: []map[string]types.AttributeValue{
					{
						"PK":                  &types.AttributeValueMemberS{Value: "123"},
						"SK":                  &types.AttributeValueMemberS{Value: "P#General"},
						"OpenSpace":           &types.AttributeValueMemberS{Value: "sameWindow"},
						"DeleteUnsavedSpaces": &types.AttributeValueMemberS{Value: "week"},
					},
				},
			},
			want:    &Preferences{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, _, err := unMarshalPreferences(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("err \n[actual] = %v, \n[want]= %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(actual) != reflect.TypeOf(tt.want) {
				t.Errorf("want \n[actual] = %v \n[want]= %v", actual, tt.want)
			}
		})
	}
}
