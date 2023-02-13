package tko

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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

		flagKind(t, f)
	}
	flag.Parse()
}

func flagKind(t reflect.StructField, f reflect.Value) {
	fmt.Println(t, f.Kind())
	switch f.Kind() {
	case reflect.String:
		flag.StringVar((*string)(f.Addr().UnsafePointer()), t.Name, f.String(), "")
	case reflect.Int:
		flag.IntVar((*int)(f.Addr().UnsafePointer()), t.Name, int(f.Int()), "")
	case reflect.Uint:
		flag.UintVar((*uint)(f.Addr().UnsafePointer()), t.Name, uint(f.Uint()), "")
	case reflect.Int64:
		flag.Int64Var((*int64)(f.Addr().UnsafePointer()), t.Name, f.Int(), "")
	case reflect.Uint64:
		flag.Uint64Var((*uint64)(f.Addr().UnsafePointer()), t.Name, f.Uint(), "")
	case reflect.Bool:
		// Something's wrong with how bool flags are parsed - it causes the other flags that come after it to be ignored. Need to look into this.
		flag.BoolVar((*bool)(f.Addr().UnsafePointer()), t.Name, f.Bool(), "")
	case reflect.Int8, reflect.Int16, reflect.Int32:
		flag.Func(t.Name, "", func(s string) error {
			v, err := strconv.ParseInt(s, 10, t.Type.Bits())
			if err != nil {
				return err
			}
			f.SetInt(v)
			return nil
		})
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		flag.Func(t.Name, "", func(s string) error {
			v, err := strconv.ParseUint(s, 10, t.Type.Bits())
			if err != nil {
				return err
			}
			f.SetUint(v)
			return nil
		})
	case reflect.Array, reflect.Slice:
		flag.Func(t.Name, "", func(s string) error {
			split := strings.Split(s, ",")
			out := reflect.MakeSlice(t.Type, 0, len(split))
			for _, tok := range split {
				fmt.Println(t.Type.Elem().Kind())
				switch t.Type.Elem().Kind() {
				case reflect.String:
					out = reflect.Append(out, reflect.ValueOf(tok))
				case reflect.Int:
					v, err := strconv.Atoi(tok)
					if err != nil {
						return err
					}
					out = reflect.Append(out, reflect.ValueOf(v))
				default:
					return fmt.Errorf("unsupported type: %v", t.Type.Elem())
				}
			}
			f.Set(out)
			return nil
		})
	case reflect.Struct:
		// TODO
	case reflect.Pointer:
		if f.IsNil() {
			f.Set(reflect.New(t.Type.Elem()))
		}
		flagKind(t, f.Elem())
	default:
		fmt.Println("unknown type", f.Kind())
	}
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
		return os.WriteFile(path, []byte(value), os.ModePerm)
	case "stderr":
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
