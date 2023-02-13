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
	"context"
	"encoding/json"
	"fmt"

	"github.com/wlynch/tko/pkg/tko"
)

// MyTask asdf
// tko:Task
type MyTask struct {
	Params  MyTaskParams `tko:"asdf"`
	Results MyTaskResults
}

type MyTaskParams struct {
	A string `tko:"asdf" json:"qwer"`
	B int
	/*
		C *int
		D *string
		E int32
		F int8
		G bool
		I []string
		J []int
		K []uint8
	*/
}

type MyTaskResults struct {
	C string
	D int
}

func (t *MyTask) Run(ctx context.Context) error {
	fmt.Println("hello", t.Params.A, t.Params.B)
	b, _ := json.Marshal(t.Params)
	fmt.Println(string(b))
	t.Results = MyTaskResults{
		C: "tacocat",
		D: 8675309,
	}
	return nil
}

func main() {
	t := &MyTask{}
	tko.Execute(context.Background(), t)
}
