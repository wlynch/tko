package tko

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

var (
	flagOutput = flag.String("tko-results", "file", "file|stderr")
)

type Runner interface {
	Run(ctx context.Context) error
}

func generateFlags(r Runner) {
	params := reflect.ValueOf(r).Elem().FieldByName("Params")
	if !params.IsValid() {
		return
	}
	for i := 0; i < params.NumField(); i++ {
		t := params.Type().Field(i)
		f := params.Field(i)

		switch f.Kind() {
		case reflect.String:
			flag.StringVar((*string)(f.Addr().UnsafePointer()), t.Name, f.String(), "")
		case reflect.Int:
			flag.IntVar((*int)(f.Addr().UnsafePointer()), t.Name, int(f.Int()), "")
		}
	}
	flag.Parse()
}

func writeResults(r Runner) error {
	results := reflect.ValueOf(r).Elem().FieldByName("Results")
	if !results.IsValid() {
		return nil
	}
	for i := 0; i < results.NumField(); i++ {
		t := results.Type().Field(i)
		f := results.Field(i)

		if !f.IsZero() {
			var data string
			switch f.Kind() {
			case reflect.String:
				data = f.String()
			case reflect.Int:
				data = fmt.Sprint(data, f.Int())
			}
			if err := writeResult(t.Name, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeResult(name, value string) error {
	switch *flagOutput {
	case "file":
		path := filepath.Join("/tekton/results", name)
		fmt.Println(path)
		err := os.WriteFile(path, []byte(value), os.ModePerm)
		fmt.Println(err)
		return err
	case "stdout":
		fmt.Fprintf(os.Stderr, "%s | %s\n", name, value)
		return nil
	}
	return nil
}

func Execute(ctx context.Context, r Runner) error {
	generateFlags(r)

	if err := r.Run(ctx); err != nil {
		return err
	}

	return writeResults(r)
}
