package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/karlseguin/liquid"
	"github.com/karlseguin/liquid/core"
)

func main() {

	core.RegisterFilter("divided_by", DividedByFactory)
	core.RegisterFilter("round", RoundFactory)

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

		data, err := readFile(dataDir + f.Name())
		if err != nil {
			log.Println(f.Name() + " : " + err.Error())
			continue FILE_ITERATOR
		}

		template, err := readFile(templDir + strings.TrimSuffix(f.Name(), extension) + "html")
		if err != nil {
			log.Println(f.Name() + " : " + err.Error())
			continue FILE_ITERATOR
		}

		outData := make(map[string]interface{})
		err = json.Unmarshal(data, &outData)
		if err != nil {
			panic(f.Name() + " : " + err.Error())
		}

		render := render(f.Name(), outData, string(template))

		writeFile(render, f.Name(), extension, outDir)
	}
}

func readFile(name string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	_, err = io.Copy(buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeFile(content []byte, fileName string, extension string, out string) {

	err := ioutil.WriteFile(out+strings.TrimSuffix(fileName, extension)+"html", content, 0644)
	if err != nil {
		panic(err)
	}

}

func render(filename string, data map[string]interface{}, templ string) []byte {
	template, err := liquid.ParseString(templ, nil)
	if err != nil {
		panic(filename + " : " + err.Error())
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

func DividedByFactory(parameters []core.Value) core.Filter {
	if len(parameters) == 0 {
		return (&DividedByFilter{}).DividedBy
	}
	return (&DividedByFilter{parameters[0]}).DividedBy
}

type DividedByFilter struct {
	divider core.Value
}

func (f *DividedByFilter) DividedBy(input interface{}, data map[string]interface{}) interface{} {

	var inputFloat float64
	var dividerFloat float64
	var err error

	switch v := input.(type) {
	case float64:
		inputFloat = v
	case string:
		inputFloat, err = strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}
	case int:
		inputFloat = float64(v)
	}

	switch v := f.divider.Underlying().(type) {
	case float64:
		dividerFloat = v
	case string:
		dividerFloat, err = strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}
	case int:
		dividerFloat = float64(v)
	}

	return inputFloat / dividerFloat
}

func RoundFactory(parameters []core.Value) core.Filter {
	if len(parameters) == 0 {
		return (&RoundFilter{}).Round
	}
	return (&RoundFilter{parameters[0]}).Round
}

type RoundFilter struct {
	decimals core.Value
}

func (f *RoundFilter) Round(input interface{}, data map[string]interface{}) interface{} {

	var inputFloat float64
	var decimalInt int
	var err error

	switch v := input.(type) {
	case float64:
		inputFloat = v
	case string:
		inputFloat, err = strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}
	case int:
		inputFloat = float64(v)
	}

	switch v := f.decimals.Underlying().(type) {
	case float64:
		decimalInt = int(v)
	case string:
		temp, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		decimalInt = int(temp)
	case int:
		decimalInt = v
	}

	return fmt.Sprintf("%."+fmt.Sprint(decimalInt)+"f", inputFloat)
}
