# 04 — Slices and Maps

Go's two workhorse collection types. No lists, no dicts, no sets — just slices and maps.

---

## Slices

A slice is a dynamically-sized view into an array. Think Python list or TypeScript array, but with some specifics worth knowing.

**Declaration:**
```go
var errs []string           // nil slice — zero value, safe to append to
errs := []string{}          // empty slice — initialized but empty
errs := make([]string, 0)   // same, with explicit allocation
```

**Appending:**
```go
errs = append(errs, "registries[%d]: name is required")
```

`internal/config/validate.go:12`

`append` returns a new slice — always reassign. A common mistake:

```go
append(errs, "foo")         // wrong — result is discarded
errs = append(errs, "foo")  // correct
```

**Iterating:**
```go
for i, r := range c.Registries {
    // i is the index, r is the value (a copy)
}
```

`internal/config/validate.go:11`

If the index isn't needed: `for _, r := range c.Registries`. The `_` discards the value — Go won't compile if a variable is declared but unused.

---

## Slice gotcha: nil vs empty

```go
var s []string    // nil — len(s) == 0, s == nil is true
s := []string{}  // empty — len(s) == 0, s == nil is false
```

Rarely matters in practice, but it can affect JSON marshaling (`null` vs `[]`). The error collection pattern throughout this codebase uses nil slices:

```go
var errs []error
// ... conditionally append ...
if len(errs) > 0 { ... }
```

`len(nil)` is 0, so this always works safely.

---

## Maps

```go
// internal/topology/topology.go:26-30
type Topology struct {
    Environments map[string]string `toml:"environments"`
}
```

A `map[K]V` maps keys of type `K` to values of type `V`.

**Two-value lookup** — the safe way to check if a key exists:
```go
t, ok := f.Topologies[name]
if !ok {
    return Topology{}, fmt.Errorf("topology %q not found in %s", name, f.Path)
}
```

`internal/topology/topology.go:83-87`

A single-value lookup on a missing key (`t := c.Topologies[name]`) returns the zero value for `V` — no panic, no error. The two-value form distinguishes "missing" from "present but zero".

**Initializing:** A nil map panics on write. Always `make` a map before writing to it:
```go
out := make(map[string]string, len(vars))
```

`internal/scaffold/prompt.go:15`

The second argument is a size hint — optional, just an optimization.

**Iterating:**
```go
for name, t := range f.Topologies {
    if t.Default {
        defaults++
    }
}
```

`internal/topology/topology.go:97`

Map iteration order is **intentionally random** in Go. If sorted output is needed, collect the keys into a slice and sort them.

---

## When to use which

- **Slice** — ordered list of things, same type
- **Map** — key→value lookups, dynamic keys (like `Environments map[string]string` where env names like "dev", "prod" aren't known at compile time)

There's no built-in set type. The Go idiom for a set is `map[T]struct{}` — the empty struct uses zero bytes.

---

## Next

[05 — Concurrency](./05-concurrency.md)
