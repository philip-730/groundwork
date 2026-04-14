# 02 — Error Handling

The thing that surprises people coming from Python or TypeScript most. Go has no exceptions. Errors are values returned from functions, handled explicitly at every call site.

---

## The pattern

```go
value, err := someFunction()
if err != nil {
    return err  // or wrap it, or handle it
}
// use value safely here
```

That's the whole pattern. `Load` does this three times in a row:

```go
// internal/config/config.go:54-65
func Load(path string) (*Config, error) {
    var cfg Config
    md, err := toml.DecodeFile(path, &cfg)
    if err != nil {
        return nil, fmt.Errorf("parse %s: %w", path, err)
    }
    if len(md.Undecoded()) > 0 {
        return nil, fmt.Errorf("parse %s: unknown keys: %v", path, md.Undecoded())
    }
    cfg.Path = path
    if err := Validate(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

Each step that can fail returns `(value, error)`. If anything fails, return `nil, err`. If everything succeeds, return `&cfg, nil`.

---

## Wrapping errors with context

The `%w` verb in `fmt.Errorf` wraps an error with additional context:

```go
return nil, fmt.Errorf("parse %s: %w", path, err)
```

Like Python's `raise ValueError("context") from original_err`. The original error is preserved and can be unwrapped later with `errors.Is` or `errors.As`. The convention is to build a chain like:

```
config: parse /path/to/groundwork.toml: toml: line 5: ...
```

Each layer adds context.

---

## Collecting multiple errors

Sometimes the right move is to collect all errors instead of bailing on the first one. `validate.go` does this:

```go
// internal/config/validate.go:10-46
func Validate(c *Config) error {
    var errs []string

    if strings.TrimSpace(c.Workspace.Name) == "" {
        errs = append(errs, "workspace.name is required")
    }
    // ... more checks ...

    if len(errs) > 0 {
        return errors.New(strings.Join(errs, "\n"))
    }
    return nil
}
```

Build up a slice of error strings, join them at the end. `SyncAll` in `internal/registry/sync.go` does the same thing but with `[]error` — keeps going even when individual registries fail, collects everything, returns it all at the end.

---

## errors.Is and os.IsNotExist

For checking what kind of error was returned:

```go
if os.IsNotExist(err) { ... }          // stdlib helper (older style)
if errors.Is(err, os.ErrNotExist) { }  // modern style
```

`discover.go` uses this to distinguish "templates directory doesn't exist yet" from a real read error:

```go
// internal/template/discover.go:20-22
if os.IsNotExist(err) {
    return nil, nil  // not an error — registry just has no templates yet
}
```

---

## Why no exceptions

The absence of exceptions is deliberate. The argument is that exceptions make control flow hard to follow — errors get thrown far from where they're handled, and it's not obvious which functions can fail. The explicit `if err != nil` is verbose, but every error is handled exactly where it occurs.

---

## Next

[03 — Structs and Interfaces](./03-structs-and-interfaces.md)
