package main

import (
	"encoding/json"
	"fmt"
	"github.com/voxelbrain/goptions"
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	VERSION = "0.4.0"
)

var (
	options = struct {
		Continue      bool   `goptions:"-c, --continue, description='Continue on error'"`
		FormatFile    string `goptions:"-r, --format-file, description='Read format from file', mutexgroup='input'"`
		FormatString  string `goptions:"-s, --format-string, description='Format string', mutexgroup='input'"`
		Format        string `goptions:"-f, --format, obligatory, description='Name for the formatter'"`
		goptions.Help `goptions:"-h, --help, description='Show this help'`
	}{
		Format: "csv",
	}
)

func main() {
	err := goptions.Parse(&options)
	if err != nil || (len(options.FormatFile) <= 0 && len(options.FormatString) <= 0) {
		if err == nil {
			err = fmt.Errorf("One of --format-file and --format-string must be specified")
		}
		fmt.Printf("Error: %s\n", err)
		goptions.PrintHelp()
		fmt.Println("Formatters:")
		for name, format := range Formats {
			fmt.Printf("\t\"%s\": %s\n", name, format.Description)
		}
		fmt.Printf("Version %s\n", VERSION)
		return
	}

	format, ok := Formats[options.Format]
	if !ok {
		log.Fatalf("Unknown format %s", options.Format)
	}

	f, err := format.Compiler(formatString())
	if err != nil {
		log.Fatalf("Template invalid: %s", err)
	}

	dec := json.NewDecoder(os.Stdin)
	logFn := NewLogFn(options.Continue)
	for {
		var input interface{}
		err := dec.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Could not decode input: %s", err)
			continue
		}
		err = f.Execute(os.Stdout, input)
		if err != nil {
			logFn("Could not apply input to template: %s", err)
			continue
		}
		io.WriteString(os.Stdout, "\n")
	}
}

type LogFn func(format string, v ...interface{})

func NewLogFn(c bool) LogFn {
	if !c {
		return func(format string, v ...interface{}) {
			log.Fatalf(format, v...)
		}
	}
	return func(format string, v ...interface{}) {
		log.Printf(format, v...)
	}
}

func formatString() string {
	if len(options.FormatFile) > 0 {
		d, e := ioutil.ReadFile(options.FormatFile)
		if e != nil {
			log.Fatalf("Could not read file \"%s\": %s", options.FormatFile, e)
		}
		return string(d)
	}
	if len(options.FormatString) > 0 {
		return options.FormatString
	}
	panic("Invalid execution path")
}
