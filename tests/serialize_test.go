package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/data"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/robertkrimen/otto/underscore"
)

type myTime struct {
	Time     time.Time     `json:"time"`
	Duration time.Duration `json:"duration"`
}

func Test_serialize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      map[string]any
		skip    bool
		want    map[string]any
		wantErr bool
	}{
		{name: "nested uuid",
			in: map[string]any{
				"container": map[string]any{
					"id": uuid.MustParse("962f999c-a9bd-40a4-80bf-47c84b1ad750"),
				},
			},
			want: map[string]any{
				"container": map[string]any{
					"id": "962f999c-a9bd-40a4-80bf-47c84b1ad750",
				},
			},
		},
		{
			name: "duration",
			in: map[string]any{
				"r": time.Second * 100,
				"a": time.Minute,
			},
			want: map[string]any{
				"r": (time.Second * 100),
				"a": time.Minute,
			},
		},
		{
			name: "time",
			in: map[string]any{
				"r": testDateTime,
			},
			want: map[string]any{
				"r": testDateTime,
			},
		},

		{
			name: "floats",
			in: map[string]any{
				"r": float64(100.50),
			},
			want: map[string]any{
				"r": 100.50,
			},
		},
		{
			name: "bytes",
			in: map[string]any{
				"r": []byte("hello world"),
			},
			want: map[string]any{
				"r": "hello world",
			},
		},
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
					"files": []any{
						map[string]any{
							"name":     "test",
							"size":     int64(10),
							"mode":     "drwxr-xr-x",
							"modified": testDateTime,
						},
					},
					"newest": map[string]any{
						"mode":     "drwxr-xr-x",
						"modified": testDateTime,
						"name":     "test",
						"size":     int64(10),
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
				"r": []any{
					map[string]any{"city_name": "Kathmandu"},
					map[string]any{"city_name": "Lalitpur"},
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
					"addresses": []any{
						map[string]any{"city_name": "Kathmandu"},
						map[string]any{"city_name": "Lalitpur"},
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
		{
			name: "nested time.Duration",
			skip: true, // TODO:
			in: map[string]any{
				"a": &myTime{
					Time:     time.Now(),
					Duration: time.Second,
				},
			},
			want: map[string]any{
				"a": map[string]any{
					"time":     time.Now(),
					"duration": time.Second,
				},
			},
		},
		{
			name: "canary checker ctx.Environment",
			in: map[string]any{
				"canary": map[string]any{
					"id": "3ac56bb3-fbf6-4479-bf3b-5f2758e9fdbe",
					"labels": map[string]string{
						"Expected-Fail":                         "true",
						"kustomize.toolkit.fluxcd.io/name":      "canaries",
						"kustomize.toolkit.fluxcd.io/namespace": "default",
					},
					"name":      "http-fail",
					"namespace": "canaries",
				},
				"check": map[string]any{
					"description": "",
					"endpoint":    "",
					"id":          "018c0096-7342-6f80-5ea5-6df245079b7a",
					"labels":      map[string]string(nil),
					"name":        "http fail test expr check",
				},
				"code":     200,
				"content":  "",
				"duration": 681,
				"elapsed":  681153615,
				"headers": map[string]string{
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Allow-Origin":      "*",
					"Content-Length":                   "0",
					"Content-Type":                     "text/html; charset=utf-8",
					"Date":                             "Wed,29 Nov 2023 06:19:30 GMT",
					"Strict-Transport-Security":        "max-age=15724800; includeSubDomains",
				},
				"json":    map[string]any{},
				"results": map[string]any{},
				"sslAge":  7004023369313149,
			},
			want: map[string]any{
				"canary": map[string]any{
					"id": "3ac56bb3-fbf6-4479-bf3b-5f2758e9fdbe",
					"labels": map[string]any{
						"Expected-Fail":                         "true",
						"kustomize.toolkit.fluxcd.io/name":      "canaries",
						"kustomize.toolkit.fluxcd.io/namespace": "default",
					},
					"name":      "http-fail",
					"namespace": "canaries",
				},
				"check": map[string]any{
					"description": "",
					"endpoint":    "",
					"id":          "018c0096-7342-6f80-5ea5-6df245079b7a",
					"labels":      map[string]any{},
					"name":        "http fail test expr check",
				},
				"code":     int64(200),
				"content":  "",
				"duration": int64(681),
				"elapsed":  int64(681153615),
				"headers": map[string]any{
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Allow-Origin":      "*",
					"Content-Length":                   "0",
					"Content-Type":                     "text/html; charset=utf-8",
					"Date":                             "Wed,29 Nov 2023 06:19:30 GMT",
					"Strict-Transport-Security":        "max-age=15724800; includeSubDomains",
				},
				"json":    map[string]any{},
				"results": map[string]any{},
				"sslAge":  int64(7004023369313149),
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		if tt.skip {
			fmt.Printf("Skipping %s\n", tt.name)
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := gomplate.Serialize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%v", diff)
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
