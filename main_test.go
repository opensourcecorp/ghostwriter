package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetFiles(t *testing.T) {
	var got interface{}
	var want interface{}

	files := getFiles("testdata/getFiles")
	got = len(files)
	want = 3
	if got != want {
		t.Errorf("Found wrong number of files (got: %d, want: %d)\n", got, want)
	}

	// .gwignore skips anything with "three" in the name, so this should instead
	// return only two files using that
	files = getFiles("testdata/getFiles", "testdata/getFiles/.gwignore")
	got = len(files)
	want = 2
	if got != want {
		t.Errorf("Found wrong number of files (got: %d, want: %d)\n", got, want)
	}
}

func TestGetGWConfig(t *testing.T) {
	// Since gwConfig structs are map[string]interface{}, all nested objects
	// discovered are map[interface{}]interface{}
	want := gwConfig{
		"developer": "ryan",
		"db": map[interface{}]interface{}{
			"host":     "some-host",
			"port":     5432,
			"user":     "postgres",
			"password": "password",
		},
		"app": map[interface{}]interface{}{
			"ip": "0.0.0.0",
		},
	}
	got := getGWConfig("testdata/ghostwriter.yaml")

	if !cmp.Equal(got, want) {
		fmt.Println(cmp.Diff(got, want))
		t.Errorf("got != want\ngot: %v\nwant:%v\n", got, want)
	}
}
