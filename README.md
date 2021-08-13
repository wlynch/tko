# TKO

tko = Tekton + [ko](https://github.com/google/ko)

This is a WIP tool for generating Tekton tasks from Go types.

The goal is to let developers write Tasks in Go to make it easier to write Tasks
with tests that don't depend on Docker / Kubernetes, and rely on code generation to handle the rest.

tko generates a ko-compatible Task + main.go, which can then be applied, published, etc.