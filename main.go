package main

import (
	"bytes"
	"flag"
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
var templateSuffix = flag.String("template-suffix", ".gw", "suffix used to discover ghostwriter templates")
var addToGitignore = flag.Bool("gitignore", false, "whether to add output directory to the repo's gitignore")
var inputPath = flag.String("input-dir", ".", "input directory path")
var outputPath = flag.String("output-dir", "gw-rendered", "root of output directory")

type cliConfig struct {
	configFile     string
	templateSuffix string // TODO: not currently used
	addToGitignore bool   // TODO: not currently used
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
		panic(err)
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

// Variadic notation on gwIgnoreFileOptional is an ugly hack to allow for a "default"
// arg value, so we can run tests
func getFiles(root string, cliConfigForIgnoring cliConfig, gwIgnoreFileOptional ...string) []fileData {
	var files []fileData

	// This is such an ugly way to walk directories and get the files
	// (filepath.Glob doesn't recurse deep enough), but... Go things
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		// path, err = filepath.Rel(root, path)
		// if err != nil {
		// 	panic(err)
		// }

		// Skip .git, and already-rendered directories
		if filepath.Base(path) == ".git" || filepath.Base(path) == cliConfigForIgnoring.outputPath {
			return filepath.SkipDir
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		// Strip off the root path (so we can ultimately write uncluttered to
		// the output directory)
		file := fileData{
			Path: strings.Replace(path, root+"/", "", 1),
			Mode: fileInfo.Mode(),
		}
		// Only return files, not directories; and also do a weak skip of the
		// config file
		if !fileInfo.IsDir() && filepath.Base(file.Path) != filepath.Base(cliConfigForIgnoring.configFile) {
			files = append(files, file)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	// Aaand here's the hack
	var gwIgnoreFile string
	if len(gwIgnoreFileOptional) == 0 {
		gwIgnoreFile = ".gwignore"
	} else {
		gwIgnoreFile = gwIgnoreFileOptional[0]
	}
	files = filterIgnoredFiles(files, gwIgnoreFile)

	return files
}

func getGWConfig(configPath string) gwConfig {
	gwConfigRaw, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var gwConfig gwConfig
	err = yaml.Unmarshal(gwConfigRaw, &gwConfig)
	if err != nil {
		panic(err)
	}

	return gwConfig
}

// filePath is just used to include better error info
func render(tplText string, gwConfig gwConfig, filePath string) string {
	tpl, err := template.New("tpl").Parse(tplText)
	if err != nil {
		log.Fatalf("Couldn't process file '%s' for some reason; bad template formatting?\n:%s\n", filePath, tplText)
		panic(err)
	}

	// If you want to return this as a string vs. rendering straight to a file
	// (which is what template.Execute expects), you just need something that
	// implements the io.Writer interface (which bytes.Buffer does) and then
	// call Execute using the object's pointer:
	var rendered bytes.Buffer
	err = tpl.Execute(&rendered, gwConfig)
	if err != nil {
		panic(err)
	}
	return rendered.String()
}

func writeRendered(rendered string, cliConfig cliConfig, file fileData) {
	outDir := filepath.Join(cliConfig.outputPath, filepath.Dir(file.Path))
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(
		filepath.Join(cliConfig.outputPath, file.Path),
		[]byte(rendered),
		file.Mode,
	)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	var cliConfig = cliConfig{
		*configFile,
		*templateSuffix,
		*addToGitignore,
		*inputPath,
		*outputPath,
	}

	gwConfig := getGWConfig(cliConfig.configFile)

	files := getFiles(cliConfig.inputPath, cliConfig)

	for _, file := range files {
		tplText, err := os.ReadFile(filepath.Join(cliConfig.inputPath, file.Path))
		if err != nil {
			panic(err)
		}
		writeRendered(
			render(string(tplText), gwConfig, file.Path),
			cliConfig,
			file,
		)
	}
}
