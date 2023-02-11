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
	"fmt"
	"go/types"
	"os"
	"os/exec"
	"strings"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"golang.org/x/tools/go/packages"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func main() {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo,
	}, os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, p := range pkgs {
		scope := p.Types.Scope()
		for _, n := range scope.Names() {
			if !strings.HasSuffix(n, "Task") {
				continue
			}
			obj := scope.Lookup(n)
			switch s := obj.Type().Underlying().(type) {
			case *types.Struct:
				out := &v1beta1.Task{
					TypeMeta: metav1.TypeMeta{
						APIVersion: v1beta1.SchemeGroupVersion.String(),
						Kind:       "Task",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: n,
					},
					Spec: v1beta1.TaskSpec{
						Steps: []v1beta1.Step{{
							Image: fmt.Sprintf("ko://%s", p.ID),
						}},
					},
				}
				for i := 0; i < s.NumFields(); i++ {
					f := s.Field(i)
					switch f.Name() {
					case "Params":
						sub := f.Type().Underlying().(*types.Struct)
						params := []v1beta1.ParamSpec{}
						args := make([]string, 0, sub.NumFields()*2)
						for j := 0; j < sub.NumFields(); j++ {
							f := sub.Field(j)
							switch sub.Field(j).Type().(type) {
							case *types.Basic:
								params = append(params, v1beta1.ParamSpec{
									Name: f.Name(),
									Type: v1beta1.ParamTypeString,
								})
								args = append(args, fmt.Sprintf("-%s", f.Name()), fmt.Sprintf("$(params.%s)", f.Name()))
							}
						}
						out.Spec.Params = params
						out.Spec.Steps[0].Args = args
					case "Results":
						sub := f.Type().Underlying().(*types.Struct)
						results := []v1beta1.TaskResult{}
						for j := 0; j < sub.NumFields(); j++ {
							f := sub.Field(j)
							switch sub.Field(j).Type().(type) {
							case *types.Basic:
								results = append(results, v1beta1.TaskResult{
									Name: f.Name(),
									Type: v1beta1.ResultsTypeString,
								})
							}
						}
						out.Spec.Results = results
					}
				}

				b, err := yaml.Marshal(out)
				if err != nil {
					panic(err)
				}
				/*
					fmt.Println("Task spec:")
					fmt.Println(string(b))
				*/

				ko := exec.Command("ko", append([]string{"resolve", "-f", "-"}, os.Args[2:]...)...)
				ko.Stdin = bytes.NewBuffer(b)
				ko.Stderr = os.Stderr
				ko.Wait()
				koout, err := ko.Output()
				if err != nil {
					panic(err)
				}
				/*
					fmt.Println("Resolved Task spec:")
					fmt.Println(string(koout))
				*/

				koDockerRepo, ok := os.LookupEnv("KO_DOCKER_REPO")
				if !ok {
					panic("KO_DOCKER_REPO not set")
				}

				bundle := fmt.Sprintf("%s/%s:latest", koDockerRepo, strings.ToLower(n))
				tkn := exec.Command("tkn", "bundle", "push", "-f", "-", bundle)
				tkn.Stdin = bytes.NewBuffer(koout)
				tkn.Stdout = os.Stdout
				tkn.Stderr = os.Stderr
				if err := tkn.Run(); err != nil {
					panic(err)
				}
			}
		}
	}
}
