package tko

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"sigs.k8s.io/yaml"
)

type Runner interface {
	Run(ctx context.Context) error
}

func generateFlags(r Runner) {
	params := reflect.ValueOf(r).Elem().FieldByName("Params")
	if !params.IsValid() || params.IsZero() {
		return
	}
	for i := 0; i < params.NumField(); i++ {
		t := params.Type().Field(i)
		f := params.Field(i)

		fmt.Println(f)
		switch f.Kind() {
		case reflect.String:
			flag.StringVar((*string)(f.Addr().UnsafePointer()), t.Name, f.String(), "")
		case reflect.Int:
			flag.IntVar((*int)(f.Addr().UnsafePointer()), t.Name, int(f.Int()), "")
		}
	}
	flag.Parse()
}

func generateParamSpec(r Runner) []*v1beta1.ParamSpec {
	params := reflect.ValueOf(r).Elem().FieldByName("Params")
	if !params.IsValid() || params.IsZero() {
		return nil
	}
	out := make([]*v1beta1.ParamSpec, 0, params.NumField())
	for i := 0; i < params.NumField(); i++ {
		t := params.Type().Field(i)
		f := params.Field(i)

		fmt.Println(f)
		switch f.Kind() {
		case reflect.String:
			out = append(out, &v1beta1.ParamSpec{
				Name:    t.Name,
				Type:    v1beta1.ParamTypeString,
				Default: v1beta1.NewArrayOrString(f.String()),
			})
		case reflect.Int:
			//flag.IntVar((*int)(f.Addr().UnsafePointer()), t.Name, int(f.Int()), "")
		}
	}
	return out
}

func generateResultSpec(r Runner) []*v1beta1.TaskResult {
	in := reflect.ValueOf(r).Elem().FieldByName("Results")
	if !in.IsValid() || in.IsZero() {
		return nil
	}
	out := make([]*v1beta1.TaskResult, 0, in.NumField())
	for i := 0; i < in.NumField(); i++ {
		t := in.Type().Field(i)
		f := in.Field(i)

		fmt.Println(f)
		switch f.Kind() {
		case reflect.String, reflect.Int:
			out = append(out, &v1beta1.TaskResult{
				Name: t.Name,
				Type: v1beta1.ResultsTypeString,
			})
		}
	}
	return out
}

func Execute(ctx context.Context, r Runner) error {
	generateFlags(r)

	err := r.Run(ctx)

	ps, _ := yaml.Marshal(generateParamSpec(r))
	fmt.Println(string(ps))

	rs, _ := yaml.Marshal(generateResultSpec(r))
	fmt.Println(string(rs))

	return err
}
