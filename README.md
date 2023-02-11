# TKO

tko = Tekton + [ko](https://github.com/google/ko)

This is a WIP tool for generating Tekton tasks from Go types.

The goal is to let developers write Tasks in Go to make it easier to write Tasks
with tests that don't depend on Docker / Kubernetes, and rely on code generation
to handle the rest.

tko generates a ko-compatible Task + main.go, which can then be applied,
published, etc.

## Generate the Task

```sh
$ export KO_DOCKER_REPO="ttl.sh/tko-example"
$ tko github.com/wlynch/tko/example
...
Creating Tekton Bundle:
        - Added Task: MyTask to image

Pushed Tekton Bundle to ttl.sh/tko-example/mytask@sha256:042026b6d4ab01f4eb4d403ebfe17f00f25b69753d9132feff35bf7de54ae88b
```

## Use the Task

```
$ cat run.yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  generateName: hello-
spec:
  params:
    - name: A
      value: "foo"
    - name: B
      value: "1"
  taskRef:
    name: "MyTask"
    bundle: ttl.sh/tko-example/mytask:latest
$ kubectl create -f run.yaml
taskrun.tekton.dev/hello-8cbnm created
$ tkn tr logs hello-8cbnm
[unnamed-0] hello foo 1
```

## Get the SBOM

Because the Task is built with ko, an SBOM is automatically generated at build
time.

```sh
$ tkn bundle list ttl.sh/tko-example/mytask@sha256:042026b6d4ab01f4eb4d403ebfe17f00f25b69753d9132feff35bf7de54ae88b -o yaml
*Warning*: This is an experimental command, it's usage and behavior can change in the next release(s)
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  creationTimestamp: null
  name: MyTask
spec:
  params:
  - name: A
    type: string
  - name: B
    type: string
  results:
  - name: C
    type: string
  - name: D
    type: string
  steps:
  - args:
    - -A
    - $(params.A)
    - -B
    - $(params.B)
    image: ttl.sh/tko-example/example-866e979aaefecbc4e8e26d4fbed651a0@sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086
    name: ""
    resources: {}
$ cosign download sbom ttl.sh/tko-example/example-866e979aaefecbc4e8e26d4fbed651a0@sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086
{
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "sbom-sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
  "spdxVersion": "SPDX-2.2",
  "creationInfo": {
    "created": "2023-02-11T00:09:29Z",
    "creators": [
      "Tool: ko 0.12.0"
    ]
  },
  "dataLicense": "CC0-1.0",
  "documentNamespace": "http://spdx.org/spdxdocs/ko/sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
  "documentDescribes": [
    "SPDXRef-Package-sha256-4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086"
  ],
  "packages": [
    {
      "SPDXID": "SPDXRef-Package-sha256-4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
      "name": "sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
      "filesAnalyzed": false,
      "licenseDeclared": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "downloadLocation": "NOASSERTION",
      "copyrightText": "NOASSERTION",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceLocator": "pkg:oci/image@sha256:4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086?mediaType=application%2Fvnd.oci.image.manifest.v1%2Bjson",
          "referenceType": "purl"
        }
      ]
    },
    {
      "SPDXID": "SPDXRef-Package-sha256-1cbda2ed63073d3e391d559467f54e6febdb0fd26d1a574f0786f30a01f3bba4",
      "name": "distroless.dev/static@sha256:1cbda2ed63073d3e391d559467f54e6febdb0fd26d1a574f0786f30a01f3bba4",
      "versionInfo": "distroless.dev/static:latest",
      "filesAnalyzed": false,
      "licenseDeclared": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "downloadLocation": "NOASSERTION",
      "copyrightText": "NOASSERTION",
      "checksums": [
        {
          "algorithm": "SHA256",
          "checksumValue": "1cbda2ed63073d3e391d559467f54e6febdb0fd26d1a574f0786f30a01f3bba4"
        }
      ],
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceLocator": "pkg:oci/image@sha256:1cbda2ed63073d3e391d559467f54e6febdb0fd26d1a574f0786f30a01f3bba4?repository_url=distroless.dev%2Fstatic\u0026tag=latest",
          "referenceType": "purl"
        }
      ]
    },
    {
      "SPDXID": "SPDXRef-Package-github.com.wlynch.tko-(devel)",
      "name": "github.com/wlynch/tko",
      "filesAnalyzed": false,
      "licenseDeclared": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "downloadLocation": "https://github.com/wlynch/tko",
      "copyrightText": "NOASSERTION",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceLocator": "pkg:golang/github.com/wlynch/tko@(devel)?type=module",
          "referenceType": "purl"
        }
      ]
    }
  ],
  "relationships": [
    {
      "spdxElementId": "SPDXRef-DOCUMENT",
      "relationshipType": "DESCRIBES",
      "relatedSpdxElement": "SPDXRef-Package-sha256-4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086"
    },
    {
      "spdxElementId": "SPDXRef-Package-sha256-4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
      "relationshipType": "DESCENDANT_OF",
      "relatedSpdxElement": "SPDXRef-Package-sha256-1cbda2ed63073d3e391d559467f54e6febdb0fd26d1a574f0786f30a01f3bba4"
    },
    {
      "spdxElementId": "SPDXRef-Package-sha256-4dc3ec2734739c31a31f1485490cc54841941d0bc5ccf4bb764ad77ae6c49086",
      "relationshipType": "CONTAINS",
      "relatedSpdxElement": "SPDXRef-Package-github.com.wlynch.tko-(devel)"
    }
  ]
}
```
