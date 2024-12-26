package tests

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type NoStructTag struct {
	Name  string
	UPPER string
}

type Address struct {
	City    string `json:"city_name"`
	Country string `json:"country,omitempty"`
}

type Person struct {
	Name      string         `json:"name"`
	Age       int            `json:"age,omitempty"`
	Address   *Address       `json:",omitempty"`
	MetaData  map[string]any `json:",omitempty"`
	Codes     []string       `json:",omitempty"`
	Addresses []Address      `json:"addresses,omitempty"`
}

var person = Person{
	Name: "John Doe",
	Age:  25,
	Address: &Address{
		City: "Kathmandu",
	},
	MetaData: map[string]any{
		"key": "value",
	},
	Codes: []string{
		"GO",
		"JS",
	},
	Addresses: []Address{
		{
			City:    "Kathmandu",
			Country: "Nepal",
		},
		{
			City:    "Bhaktapur",
			Country: "Nepal",
		},
		{
			City:    "New York",
			Country: "USA",
		},
	},
}

// A shared test data for all template test
var structEnv = map[string]any{
	"results": Person{
		Name: "Aditya",

		Address: &Address{
			City: "Kathmandu",
		},
	},
}

type inlineJson struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

var inline map[string]interface{}
var structJson string

func init() {
	s, _ := json.MarshalIndent(person, "", "  ")
	encoded := base64.StdEncoding.EncodeToString(s)

	inline = map[string]interface{}{
		"Name": "Jane Doe",
		"Data": encoded,
	}
	s, _ = json.MarshalIndent(inline, "", "  ")
	structJson = string(s)
}

type FolderCheck struct {
	Oldest        *File  `json:"oldest,omitempty"`
	Newest        *File  `json:"newest,omitempty"`
	MinSize       *File  `json:"smallest,omitempty"`
	MaxSize       *File  `json:"largest,omitempty"`
	TotalSize     int64  `json:"size,omitempty"`
	AvailableSize int64  `json:"availableSize,omitempty"`
	Files         []File `json:"files,omitempty"`
}

type File struct {
	Name     string    `json:"name,omitempty"`
	Size     int64     `json:"size,omitempty"`
	Mode     string    `json:"mode,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
	IsDir    bool      `json:"is_dir,omitempty"`
}

var testDate = "2021-10-05T09:00:00Z"
var testDateTime, _ = time.Parse(time.RFC3339Nano, testDate)

func newFile() File {
	t, _ := time.Parse(time.RFC3339, testDate)
	return File{
		Name:     "test",
		Size:     10,
		Mode:     "drwxr-xr-x",
		Modified: t,
	}
}

func newFolderCheck(count int) *FolderCheck {
	f := newFile()
	folder := FolderCheck{
		Newest: &f,
	}
	for i := 0; i < count; i++ {
		folder.Files = append(folder.Files, newFile())
	}
	return &folder
}

type Totals struct {
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Skipped  int     `json:"skipped,omitempty"`
	Error    int     `json:"error,omitempty"`
	Duration float64 `json:"duration"`
}

type JunitTest struct {
	Name string `json:"name" yaml:"name"`
}

type JunitTestSuite struct {
	Name   string `json:"name"`
	Totals `json:",inline"`
	Tests  []JunitTest `json:"tests"`
}

type JunitTestSuites struct {
	Suites []JunitTestSuite `json:"suites,omitempty"`
	Totals `json:",inline"`
}

var junitEnv = JunitTestSuites{
	Totals: Totals{
		Passed: 1,
	},
	Suites: []JunitTestSuite{{Name: "hi", Totals: Totals{Failed: 2}}},
}

type SQLDetails struct {
	Rows  []map[string]interface{} `json:"rows,omitempty"`
	Count int                      `json:"count,omitempty"`
}
