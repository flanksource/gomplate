package tests

import (
	"testing"

	"github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/data"
	_ "github.com/flanksource/gomplate/v3/js"
	_ "github.com/robertkrimen/otto/underscore"
)

func Test_serialize(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "dates",
			in: map[string]any{
				"r": newFile().Modified,
			},
			want: map[string]any{
				"r": newFile().Modified,
			},
		},
		{
			name: "nested_pointers",
			in: map[string]any{
				"r": newFolderCheck(1),
			},
			want: map[string]any{
				"r": map[string]any{
					"newest": map[string]any{
						"mode":     "drwxr-xr-x",
						"modified": testDateTime,
						"name":     "test",
						"size":     10.0,
					},
					"files": []map[string]any{
						{
							"name":     "test",
							"size":     10.0,
							"mode":     "drwxr-xr-x",
							"modified": testDateTime,
						},
					},
				},
			},
		},
		{name: "nil", in: nil, want: nil, wantErr: false},
		{name: "empty", in: map[string]any{}, want: map[string]any{}, wantErr: false},
		{
			name:    "simple - no struct tags",
			in:      map[string]any{"r": NoStructTag{Name: "Kathmandu", UPPER: "u"}},
			want:    map[string]any{"r": map[string]any{"Name": "Kathmandu", "UPPER": "u"}},
			wantErr: false,
		},
		{name: "simple - struct tags", in: map[string]any{"r": Address{City: "Kathmandu"}}, want: map[string]any{"r": map[string]any{"city_name": "Kathmandu"}}, wantErr: false},
		{
			name:    "nested struct",
			in:      map[string]any{"r": Person{Name: "Aditya", Address: &Address{City: "Kathmandu"}}},
			want:    map[string]any{"r": map[string]any{"name": "Aditya", "Address": map[string]any{"city_name": "Kathmandu"}}},
			wantErr: false,
		},
		{
			name: "slice of struct",
			in: map[string]any{
				"r": []Address{
					{City: "Kathmandu"},
					{City: "Lalitpur"},
				},
			},
			want: map[string]any{
				"r": []map[string]any{
					{"city_name": "Kathmandu"},
					{"city_name": "Lalitpur"},
				},
			},
			wantErr: false,
		},
		{
			name: "nested slice of struct",
			in: map[string]any{
				"r": Person{
					Name: "Aditya",
					Addresses: []Address{
						{City: "Kathmandu"},
						{City: "Lalitpur"},
					},
				},
			},
			want: map[string]any{
				"r": map[string]any{
					"name": "Aditya",
					"addresses": []map[string]any{
						{"city_name": "Kathmandu"},
						{"city_name": "Lalitpur"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pointers",
			in: map[string]any{
				"r": &Address{
					City: "Bhaktapur",
				},
			},
			want: map[string]any{
				"r": map[string]any{
					"city_name": "Bhaktapur",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gomplate.Serialize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_got, _ := data.ToJSONPretty("  ", got)
			_want, _ := data.ToJSONPretty("  ", tt.want)
			if _got != _want {
				t.Errorf("serialize() = \n%s\nwant\n %v", _got, _want)
			}
		})
	}
}
