# groundwork

[![CI](https://github.com/philip-730/groundwork/actions/workflows/ci.yml/badge.svg)](https://github.com/philip-730/groundwork/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/philip-730/groundwork/branch/main/graph/badge.svg)](https://codecov.io/gh/philip-730/groundwork)
[![Go](https://img.shields.io/github/go-mod/go-version/philip-730/groundwork)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/philip-730/groundwork)](https://goreportcard.com/report/github.com/philip-730/groundwork)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)

Topology-aware GCP service scaffolding. Get files, not a framework.

---

Groundwork generates boilerplate for GCP service archetypes from a template registry. Define your org's topology once — environment project IDs, shared Artifact Registry, default region — and every scaffold resolves those values automatically. No placeholders to fill in by hand.

This is the same model as [shadcn/ui](https://ui.shadcn.com/): the tool is the delivery mechanism. Once the files are written they're yours — edit them, diverge from the template, delete what you don't need. Groundwork has no runtime presence and no opinion about what you do after it runs.

No lockfile. No upgrade command. No drift detection.

## Quickstart

```sh
groundwork init https://github.com/my-org/groundwork-registry
groundwork scaffold cloud-run-service --var service_name=payments-api
```

```
Inputs:
  region (string, default: us-east1):

Files to be written to ./payments-api:

  terraform/main.tf
  terraform/variables.tf
  terraform/outputs.tf
  terraform/backend.tf
  terraform/environments/dev.tfvars
  terraform/environments/prod.tfvars
  cloudbuild/build.yaml
  nix/flake.nix

Proceed? [y/N]
```

Every project ID, region, and Artifact Registry path is resolved from your topology. No placeholders.

## Why not Backstage?

Backstage solves discoverability — a catalog of services, owners, runbooks. If you have that problem, use Backstage.

Groundwork solves boilerplate — you know what you're building, you just don't want to copy-paste Terraform and Cloud Build config for the fifth time. It's a CLI that runs once and gets out of the way.

## Concepts

### Registry

A git repo with a `templates/` directory and a `topologies.toml` at the root. Your org maintains one. Groundwork clones it locally on `init` or `registry sync` and reads from the local cache.

```
my-registry/
  topologies.toml
  templates/
    cloud-run-service/
      template.toml
      terraform/
        main.tf.tmpl
        variables.tf.tmpl
      cloudbuild/
        build.yaml.tmpl
```

### Topology

A named mapping of your GCP org structure defined in `topologies.toml` at the root of the registry. Templates are topology-aware — they receive your actual project IDs and shared resource references at render time.

```toml
# topologies.toml
[topologies.primary]
default = true

  [topologies.primary.environments]
  dev  = "my-org-dev"
  prod = "my-org-prod"

  [topologies.primary.shared]
  project = "my-org-shared"
  region  = "us-east1"

    [topologies.primary.shared.artifact_registry]
    location   = "us-east1"
    repository = "my-org-images"
```

What belongs in topology: environment project IDs, shared Artifact Registry, default region. Structural, stable, org-wide values that every service will need.

What doesn't: per-service inputs like dataset names, bucket names, secret prefixes. Those are prompted at scaffold time.

### Template

A directory of `*.tmpl` files rendered with Go's `text/template`. The render context:

```
{{ .Inputs.service_name }}
{{ .Topology.Environments.dev }}
{{ .Topology.Environments.prod }}
{{ .Topology.Shared.Project }}
{{ .Topology.Shared.Region }}
{{ .Topology.Shared.ArtifactRegistry.Location }}
{{ .Topology.Shared.ArtifactRegistry.Repository }}
```

Non-`.tmpl` files are copied verbatim. `template.toml` is never included in output.

## CLI

```
groundwork init <url> [--name name]

groundwork scaffold <template> [flags]

  --topology  name    topology to use (defaults to the one marked default)
  --out       dir     output directory (defaults to ./<service_name>)
  --dry-run           print rendered files without writing
  --var       key=val pass input values, skipping the interactive prompt
                      (repeatable: --var a=1 --var b=2)

groundwork registry sync
groundwork registry list
groundwork registry inspect <template>
```

## template.toml reference

```toml
[template]
name        = "cloud-run-service"
description = "HTTP service on Cloud Run with Cloud Build CI/CD."

[inputs]
service_name = { type = "string", required = true }
region       = { type = "string", default = "us-east1" }

[topology]
requires = ["environments.dev", "environments.prod", "shared.artifact_registry"]
```

`topology.requires` is validated before rendering. If your topology is missing a required key, scaffold fails with a clear error.

## Development

```sh
nix develop   # drop into dev shell with go, gopls, gotools, git, golangci-lint
go test ./...
go build -o gw .
```
