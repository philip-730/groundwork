# 03 — Structs and Interfaces

No classes. No inheritance. Structs for data, interfaces for behaviour, kept deliberately separate.

---

## Structs

A struct is a named collection of fields:

```go
// internal/config/config.go:13-20
type Config struct {
    Registries []Registry `toml:"registries"`
    Path       string     `toml:"-"`
}
```

The backtick strings are **struct tags** — metadata read at runtime by libraries. The `toml:"registries"` tag tells the BurntSushi decoder which TOML key maps to this field. `toml:"-"` means "skip this field during decoding" — `Path` is set by code, not read from the file. The same pattern shows up with `json:"..."` tags when working with JSON APIs.

---

## Methods

Methods are functions attached to a type via a **receiver**:

```go
// internal/topology/topology.go:68
func (f *File) DefaultTopology() (string, Topology, error) {
```

The `(f *File)` part is the receiver — `f` is the instance, like `self` in Python or `this` in TypeScript. The `*` means it's a pointer receiver (can mutate the struct; avoids copying on every call).

---

## No inheritance — composition instead

Go has no `extends`. Instead, structs are nested inside other structs:

```go
// internal/topology/topology.go:26-30
type Topology struct {
    Default      bool              `toml:"default"`
    Environments map[string]string `toml:"environments"`
    Shared       SharedConfig      `toml:"shared"`
}
```

`SharedConfig` is a separate struct composed into `Topology`. Its fields are accessed as `topo.Shared.Project`. This is the Go way of building up complex types — compose, don't inherit.

---

## Interfaces

An interface defines a set of methods. Any type that implements those methods satisfies the interface — no `implements` keyword needed.

The most important interface in use here is `io.Writer`:

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`CollectInputs`, `Scaffold`, `Sync`, and others take an `io.Writer` instead of writing directly to `os.Stdout`:

```go
// internal/scaffold/prompt.go:36
func CollectInputs(t *tmpl.Template, provided map[string]string, r io.Reader, w io.Writer) (map[string]string, error) {
```

In tests, a `bytes.Buffer` (which implements `io.Writer`) gets passed in to capture output without touching stdout. In production, `os.Stdout` gets passed. The function doesn't care — it just calls `.Write()`.

This is Go's answer to dependency injection. No frameworks, no decorators — just pass an interface.

---

## The empty interface and `any`

`any` (alias for `interface{}`) means "any type". Shows up in generic-ish situations. Avoid it when possible — Go added real generics in 1.18 for this reason.

---

## `io.Reader` and `io.Writer` in this codebase

These two interfaces are everywhere in the standard library:

- `io.Reader` — anything readable (files, stdin, network connections, byte buffers)
- `io.Writer` — anything writable

`Scaffold` takes both:

```go
// internal/scaffold/scaffold.go:25-27
Stdin  io.Reader
Stdout io.Writer
```

`cmd/scaffold.go` passes `os.Stdin` and `os.Stdout`. The tests pass `strings.NewReader(...)` and `bytes.Buffer`. Same function, totally different I/O.

---

## Next

[04 — Slices and Maps](./04-slices-and-maps.md)
