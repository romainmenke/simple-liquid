package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/karlseguin/liquid"
	"github.com/karlseguin/liquid/core"
)

func main() {

	core.RegisterFilter("divided_by", DummyFactory)
	core.RegisterFilter("round", DummyFactory)

	dataSource := flag.String("data", "./", "data directory")
	templateSource := flag.String("templ", "./", "template directory")
	flag.Parse()

	templDir := strings.TrimSuffix(*templateSource, "/") + "/"
	dataDir := strings.TrimSuffix(*dataSource, "/") + "/"
	outDir := dataDir

	exclude := flag.Args()

	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		panic(err)
	}

FILE_ITERATOR:
	for _, f := range files {
		if !isFile(dataDir + f.Name()) {
			continue FILE_ITERATOR
		}

		for _, exc := range exclude {
			if strings.Contains(f.Name(), exc) {
				continue FILE_ITERATOR
			}
		}

		if !strings.Contains(f.Name(), ".json") {
			continue FILE_ITERATOR
		}

		nameComponents := strings.Split(f.Name(), ".")
		extension := nameComponents[len(nameComponents)-1]

		data := readFile(dataDir + f.Name())
		template := readFile(templDir + strings.TrimSuffix(f.Name(), extension) + "html")

		outData := make(map[string]interface{})
		err := json.Unmarshal(data, &outData)
		if err != nil {
			panic(err)
		}

		render := render(outData, string(template))

		writeFile(render, f.Name(), extension, outDir)
	}
}

func readFile(name string) []byte {
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(buf, file)
	if err != nil {
		panic(err)
	}
	file.Close()
	return buf.Bytes()
}

func writeFile(content []byte, fileName string, extension string, out string) {

	err := ioutil.WriteFile(out+strings.TrimSuffix(fileName, extension)+"html", content, 0644)
	if err != nil {
		panic(err)
	}

}

func render(data map[string]interface{}, templ string) []byte {
	template, err := liquid.ParseString(templ, nil)
	if err != nil {
		panic(err)
	}
	writer := new(bytes.Buffer)
	template.Render(writer, data)
	return writer.Bytes()
}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return !fileInfo.IsDir()
}

func createIfMissing(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
}

func DummyFactory(parameters []core.Value) core.Filter {
	if len(parameters) == 0 {
		return (&DummyFilter{}).Dummy
	}
	return (&DummyFilter{parameters[0]}).Dummy
}

type DummyFilter struct {
	divider core.Value
}

func (f *DummyFilter) Dummy(input interface{}, data map[string]interface{}) interface{} {

	return input
}
