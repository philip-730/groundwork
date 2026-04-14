# 05 — Concurrency

Go's killer feature. Goroutines are cheap, lightweight threads managed by the Go runtime — not OS threads. Spinning up thousands of them is normal.

This codebase doesn't use concurrency yet, but `SyncAll` in `internal/registry/sync.go` is the obvious candidate — right now it syncs registries one at a time sequentially. Here's how it would look if parallelized.

---

## Goroutines

Launching a goroutine is just `go` in front of a function call:

```go
go doSomething()
go func() {
    // anonymous function
}()
```

That's it. The function runs concurrently. The problem is coordinating the results.

---

## sync.WaitGroup

The simplest coordination primitive — wait for a set of goroutines to finish:

```go
var wg sync.WaitGroup

for _, reg := range registries {
    wg.Add(1)
    go func(r config.Registry) {
        defer wg.Done()
        // do work
    }(reg)
}

wg.Wait()  // blocks until all goroutines call Done()
```

Note the `reg` being passed as an argument to the anonymous function. If the loop variable was captured directly (`func() { use(reg) }`), all goroutines would see the same final value of `reg` by the time they run — a classic Go gotcha (fixed in Go 1.22, but still worth knowing).

---

## Channels

Channels are typed pipes for communication between goroutines:

```go
results := make(chan SyncResult, len(registries))  // buffered channel

for _, reg := range registries {
    go func(r config.Registry) {
        res, err := Sync(r, cacheRoot, w)
        results <- res  // send
    }(reg)
}

for i := 0; i < len(registries); i++ {
    res := <-results  // receive
}
```

A **buffered** channel (second argument to `make`) won't block the sender until the buffer is full. An **unbuffered** channel blocks the sender until someone receives.

---

## What a concurrent SyncAll might look like

The current sequential version in `internal/registry/sync.go:30-56`:

```go
for _, reg := range registries {
    fmt.Fprintf(w, "syncing %s (%s)\n", reg.Name, reg.URL)
    res, err := Sync(reg, cacheRoot, w)
    if err != nil {
        errs = append(errs, fmt.Errorf("%s: %w", reg.Name, err))
        continue
    }
    results = append(results, res)
}
```

A parallel version would launch all `Sync` calls simultaneously and collect results — useful when there are many registries and each involves a network call.

The tricky part is that concurrent writes to shared slices (`results`, `errs`) aren't safe. The options are:
- Use a `sync.Mutex` to protect shared state
- Use channels to send results back to a single collector goroutine
- Use `errgroup` from `golang.org/x/sync` which handles this pattern cleanly

---

## The rule: don't communicate by sharing memory, share memory by communicating

The Go proverb. Instead of multiple goroutines writing to a shared variable (and adding locks everywhere), pass data through channels. One goroutine owns the data, others send to it.

---

## When to reach for concurrency

Not always. Sequential code is easier to read, test, and debug. Concurrency is worth it when:

- There are independent I/O operations that can overlap (network calls, file reads)
- The sequential version is measurably too slow
- The concurrent version stays readable

`SyncAll` is a good candidate — each registry sync is a separate network call with no dependencies between them.
