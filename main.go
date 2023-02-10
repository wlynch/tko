// Copyright 2021 Billy Lynch
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/wlynch/tko/example"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	tmpl = template.Must(template.New("main.go.tmpl").Funcs(template.FuncMap{
		"title": strings.Title,
		"zero": func(t reflect.Type) string {
			out := fmt.Sprint(reflect.Zero(t))
			if out == "" {
				return `""`
			}
			return out
		},
	}).ParseFiles("main.go.tmpl"))
)

func main() {
	in := example.MyTask{}
	t := reflect.TypeOf(in)
	base := filepath.FromSlash(fmt.Sprintf("%s/tko-%s", path.Base(t.PkgPath()), strings.ToLower(t.Name())))
	if err := os.MkdirAll(base, 0755); err != nil {
		log.Fatal(err)
	}

	out, _ := generateYAML(in)
	b, _ := yaml.Marshal(out)
	fmt.Println(string(b))
	if err := os.WriteFile(filepath.Join(base, "task.yaml"), b, 0644); err != nil {
		log.Fatal(err)
	}

	outGo, err := generateGo(in)
	fmt.Println(string(outGo), err)
	if err := os.WriteFile(filepath.Join(base, "main.go"), outGo, 0644); err != nil {
		log.Fatal(err)
	}
}

func generateYAML(in interface{}) (v1beta1.Task, error) {
	out := v1beta1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1beta1",
			Kind:       "Task",
		},
	}

	t := reflect.TypeOf(in)
	out.Name = strings.ToLower(t.Name())

	sf, ok := t.FieldByName("Params")
	if ok {
		pt := sf.Type
		paramSpecs := make([]v1beta1.ParamSpec, 0, pt.NumField())
		args := make([]string, 0, pt.NumField()*2)

		for i := 0; i < pt.NumField(); i++ {
			psf := pt.Field(i)

			pType := v1beta1.ParamTypeString
			if psf.Type.Kind() == reflect.Slice {
				pType = v1beta1.ParamTypeArray
				args = append(args, fmt.Sprintf("--%s", psf.Name), fmt.Sprintf("$(params.%s[*])", psf.Name))
			} else {
				args = append(args, fmt.Sprintf("--%s", psf.Name), fmt.Sprintf("$(params.%s)", psf.Name))

			}

			paramSpecs = append(paramSpecs, v1beta1.ParamSpec{
				Name: psf.Name,
				Type: pType,
			})
		}
		step := v1beta1.Step{
			Name:  strings.ToLower(t.Name()),
			Image: fmt.Sprintf("ko://%s/tko-%s", t.PkgPath(), strings.ToLower(t.Name())),
			Args:  args,
		}
		out.Spec.Steps = []v1beta1.Step{step}

		out.Spec.Params = paramSpecs

	}
	return out, nil
}

type tmplValues struct {
	Import  string
	Package string
	Name    string

	ParamsName string
	Params     []reflect.StructField
}

func generateGo(in interface{}) ([]byte, error) {
	tv := new(tmplValues)

	t := reflect.TypeOf(in)
	tv.Import = t.PkgPath()
	tv.Package = path.Base(t.PkgPath())
	tv.Name = t.Name()
	sf, ok := t.FieldByName("Params")
	if !ok {
		return nil, errors.New("no Params field")
	}

	pt := sf.Type
	tv.ParamsName = sf.Type.Name()
	for i := 0; i < pt.NumField(); i++ {
		tv.Params = append(tv.Params, pt.Field(i))
	}

	b := new(bytes.Buffer)
	if err := tmpl.Execute(b, tv); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
