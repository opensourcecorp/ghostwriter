package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const gwConfigFile = "testdata/ghostwriter.yaml"

func TestGetFiles(t *testing.T) {
	var got int
	var want int

	want = 3
	files := getFiles("testdata/getFiles", cliConfig{})
	got = len(files)
	if got != want {
		t.Errorf("Found wrong number of files (got: %d, want: %d)\n", got, want)
	}

	// This implicitly tests filterIgnoredFiles(). The .gwignore provided skips
	// anything with "three" in the name, so this should instead return only two
	// files using that
	want = 2
	files = getFiles("testdata/getFiles", cliConfig{}, "testdata/getFiles/.gwignore")
	got = len(files)
	if got != want {
		t.Errorf("Found wrong number of files (got: %d, want: %d)\n", got, want)
	}
}

func TestGetGWConfig(t *testing.T) {
	// Since gwConfig structs are map[string]interface{}, all nested objects
	// discovered end up being map[interface{}]interface{}
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

	got := getGWConfig(gwConfigFile)

	if !cmp.Equal(got, want) {
		fmt.Println(cmp.Diff(got, want))
		t.Errorf("got: %v\nwant:%v\n", got, want)
	}
}

func TestRender(t *testing.T) {
	want := "db port is 5432, user is postgres"

	got := render(
		"db port is {{ .db.port }}, user is {{ .db.user }}",
		getGWConfig(gwConfigFile),
		"",
	)

	if got != want {
		t.Errorf("Rendered template text wasn't what was expected\ngot: '%v'\nwant: '%v'\n", got, want)
	}
}

// This also implicitly tests main()
func TestWriteRendered(t *testing.T) {
	// First, let's test that we can render directory contents as expected of
	// most usage, but checking a single file for brevity
	want := "db port is 5432, user is postgres\n"

	cliConfig := cliConfig{
		gwConfigFile,
		".gw",
		false,
		"testdata",
		"/tmp/ghostwriter-tests",
	}

	files := getFiles(cliConfig.inputPath, cliConfig)

	var file fileData
	for _, fileI := range files {
		if fileI.Path == "writeRendered/render_me.txt" {
			file = fileI
		}
	}

	tplText, err := os.ReadFile(filepath.Join(cliConfig.inputPath, file.Path))
	if err != nil {
		log.Fatal(err)
	}
	rendered := render(
		string(tplText),
		getGWConfig(gwConfigFile),
		file.Path,
	)

	writeRendered(rendered, cliConfig, file)

	gotRaw, err := os.ReadFile(filepath.Join(cliConfig.outputPath, file.Path))
	if err != nil {
		log.Fatal(err)
	}
	got := string(gotRaw)

	if got != want {
		t.Errorf("got: %v\nwant: %v\n", got, want)
	}

	// Now, let's make sure a user can just render a single file at a time if
	// desired

}
