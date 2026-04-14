# 01 — The Basics

---

## Packages, not modules

Every `.go` file starts with a `package` declaration. Files in the same directory share a package and can access each other's unexported identifiers freely. The `internal/` convention is enforced by the compiler — nothing outside this repo can import `github.com/philip-730/groundwork/internal/...`.

`main.go` is the simplest case:

```go
package main

import "github.com/philip-730/groundwork/cmd"

func main() {
    cmd.Execute()
}
```

One file, one package, one function. The entry point is always `main.main`.

---

## Variables and `:=`

Two ways to declare variables:

```go
var name string = "foo"  // explicit — rarely used
name := "foo"            // short declaration — the common form
```

`:=` infers the type from the right-hand side. The inlay hints in Helix showing `name string :=` are just gopls making the inferred type visible — it doesn't get written.

The rule: `:=` only works inside a function. Package-level variables need `var`.

---

## Types are explicit, but inference does most of the work

Go is statically typed with no implicit coercion. No `"5" + 5`, no truthiness. An `int` and an `int64` are different types and can't be mixed without a cast.

In practice though, `:=` and function return types do most of the heavy lifting. Type annotations in function bodies are rare.

---

## Functions

```go
func Load(path string) (*Config, error) {
```

`internal/config/config.go:54`

Two things worth noting:

1. **Types come after names** — `path string`, not `string path`
2. **Multiple return values** — Go functions routinely return `(value, error)`. This is idiomatic, not a workaround.

The `*Config` means "pointer to Config". More on that below.

---

## Pointers

Go has pointers, but they're much simpler than C. Two main uses:

**1. Returning a pointer to avoid copying a large struct:**
```go
func Load(path string) (*Config, error) {
    var cfg Config
    // ...
    return &cfg, nil  // & takes the address
}
```
`internal/config/config.go:54-65`

**2. Allowing a method to modify its receiver:**
```go
func (f *File) DefaultTopology() (string, Topology, error) {
```
`internal/topology/topology.go:68`

The `*` in `*Config` means pointer-to-Config. The `&` operator gets a pointer to a value. No pointer arithmetic, no manual memory management — the GC handles it.

---

## Zero values

Every type has a zero value. Uninitialized variables are not null — they're the zero value for their type:

- `string` → `""`
- `int` → `0`
- `bool` → `false`
- pointer → `nil`
- slice → `nil` (safe to append to)
- map → `nil` (NOT safe to write to — needs `make`)

This matters in practice. The empty string checks in `validate.go` like `strings.TrimSpace(c.Workspace.Name) == ""` work because an unset string field is guaranteed to be `""`, not undefined or null.

---

## Next

[02 — Error Handling](./02-error-handling.md)
