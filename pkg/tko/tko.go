package tko

import (
	"context"
	"flag"
	"reflect"
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

func Execute(ctx context.Context, r Runner) error {
	generateFlags(r)

	err := r.Run(ctx)

	return err
}
