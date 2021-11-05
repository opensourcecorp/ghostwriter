package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

var configFile = flag.String("config-file", "ghostwriter.yaml", "ghostwriter config file")
var recursive = flag.Bool("recurse", true, "whether to recurse into directories to find templates")
var templateSuffix = flag.String("template-suffix", ".gw", "suffix used to discover ghostwriter templates")
var addToGitignore = flag.Bool("gitignore", false, "whether to add output directory to the repo's gitignore")
var inputPath = flag.String("input", ".", "input directory path")
var outputPath = flag.String("output", "rendered", "root of output directory")

type cliConfig struct {
	configFile     string
	recursive      bool
	templateSuffix string
	addToGitignore bool
	inputPath      string
	outputPath     string
}

type gwConfig map[string]interface{}

type fileData struct {
	Path string
	Mode fs.FileMode
}

func filterIgnoredFiles(files []fileData, gwIgnoreFile string) []fileData {
	var filesOut []fileData

	gwIgnoreRaw, err := os.ReadFile(gwIgnoreFile)
	if err != nil {
		log.Fatal(err)
	}

	// Cleaning up newline-delimited files seems to be kind of gross in Go;
	// Python users beware!
	gwIgnore := strings.Join(strings.Split(strings.Trim(string(gwIgnoreRaw), "\n"), "\n"), "|")
	// Also skip some default stuff
	protectedIgnore := []string{
		".gwignore",
	}
	gwIgnore = gwIgnore + "|" + strings.Join(protectedIgnore, "|")

	ignoreRegex := regexp.MustCompile(gwIgnore)

	for _, file := range files {
		pathIsIgnored := ignoreRegex.MatchString(file.Path)
		if pathIsIgnored {
			// log.Printf("Skipping file %s because it's ignored\n", file.Path)
			continue
		} else {
			filesOut = append(filesOut, file)
		}
	}
	return filesOut
}

// Variadic notation on gwIgnoreFile is an ugly hack to allow for a "default"
// arg value, so we can run tests
func getFiles(root string, gwIgnoreFile ...string) []fileData {
	var files []fileData

	// This is such an ugly way to walk directories and get the files
	// (filepath.Glob doesn't recurse deep enough), but... Go things
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		// path, err = filepath.Rel(root, path)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// Skip .git directories
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}

		// Strip off the root path (so we can ultimately write uncluttered to
		// the output directory)
		file := fileData{
			Path: strings.Replace(path, root+"/", "", 1),
			Mode: fileInfo.Mode(),
		}
		// Only return files, not directories
		if !fileInfo.IsDir() {
			files = append(files, file)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Aaand here's the hack
	var gwIgnoreFile_ string
	if len(gwIgnoreFile) == 0 {
		gwIgnoreFile_ = ".gwignore"
	} else {
		gwIgnoreFile_ = gwIgnoreFile[0]
	}
	files = filterIgnoredFiles(files, gwIgnoreFile_)

	return files
}

func getGWConfig(configPath string) gwConfig {
	gwConfigRaw, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	var gwConfig gwConfig
	err = yaml.Unmarshal(gwConfigRaw, &gwConfig)
	if err != nil {
		log.Fatal(err)
	}

	return gwConfig
}

func render(tplText string, gwConfig gwConfig) string {
	tpl, err := template.New("tpl").Parse(tplText)
	if err != nil {
		log.Fatalf("Couldn't process this string for some reason:\n%s\n", tplText)
		log.Fatal(err)
	}

	// If you want to return this as a string vs. rendering straight to a file
	// (which is what template.Execute expects), you just need something that
	// implements the io.Writer interface (which bytes.Buffer does) and then
	// call Execute using the object's pointer:
	var rendered bytes.Buffer
	err = tpl.Execute(&rendered, gwConfig)
	if err != nil {
		log.Fatal(err)
	}
	return rendered.String()
}

func writeRendered(rendered string, path string, mode fs.FileMode) {
	os.WriteFile(path, []byte(rendered), mode)
}

func main() {
	flag.Parse()
	var cliConfig = cliConfig{
		*configFile,
		*recursive,
		*templateSuffix,
		*addToGitignore,
		*inputPath,
		*outputPath,
	}

	gwConfig := getGWConfig(cliConfig.configFile)
	fmt.Println(gwConfig)

	files := getFiles(cliConfig.inputPath)
	for _, file := range files {
		fmt.Println(file.Path)
	}

	var rendered string
	for i, file := range files {
		tplText, err := os.ReadFile(file.Path)
		if err != nil {
			log.Fatal(err)
		}
		rendered = render(string(tplText), gwConfig)
		writeRendered(rendered, filepath.Join(cliConfig.outputPath, file.Path), file.Mode)
		if i == 0 {
			break
		}
	}
	// fmt.Println(rendered)
}
