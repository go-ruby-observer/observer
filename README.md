# go-ruby-observer/observer

![go-ruby-observer/observer](https://raw.githubusercontent.com/go-ruby-observer/brand/main/social/go-ruby-observer-observer.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/go-ruby-observer/observer.svg)](https://pkg.go.dev/github.com/go-ruby-observer/observer)
[![License: BSD-3-Clause](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/go-ruby-observer/observer/actions/workflows/ci.yml/badge.svg)](https://github.com/go-ruby-observer/observer/actions/workflows/ci.yml)

A pure-Go (**CGO=0**) implementation of the core state behind Ruby's
**`Observable`** mixin — the observer registry and the *changed* flag — faithful
to MRI Ruby's `lib/observer.rb` and verified against the reference interpreter
(**Ruby 4.0.5**) with `ruby -robserver`.

It is part of the `go-ruby-*` family of standalone front-end/runtime components
(alongside [go-ruby-parser](https://github.com/go-ruby-parser/parser),
[go-ruby-regexp](https://github.com/go-ruby-regexp/regexp),
[go-ruby-marshal](https://github.com/go-ruby-marshal/marshal) and
[go-ruby-erb](https://github.com/go-ruby-erb/erb)) that
[go-embedded-ruby](https://github.com/go-embedded-ruby) builds on. It has no
dependency on any interpreter: it owns the registry and answers *whether* and *in
what order* observers should be notified, while the actual method dispatch onto
each observer — `update` and friends — stays in the embedding interpreter
(rbgo).

## What it owns vs. what rbgo binds

A [`Registry`](observer.go) backs one object that includes `Observable`:

| MRI `Observable` | this package |
| --- | --- |
| `add_observer(obj, func=:update)` | `Registry.AddObserver` |
| `delete_observer(obj)` | `Registry.DeleteObserver` |
| `delete_observers` | `Registry.DeleteObservers` |
| `count_observers` | `Registry.CountObservers` |
| `changed(state=true)` | `Registry.Changed` |
| `changed?` | `Registry.ChangedQ` |
| `notify_observers(*args)` | `Registry.NotifyObservers` |

`NotifyObservers` returns the ordered `(observer, method, args)` tuples to call;
**rbgo performs the dispatch** (invoking each observer's method through the
interpreter). Responsiveness (`respond_to?`) is delegated to a caller-supplied
callback so the package stays interpreter-agnostic.

## MRI-faithful semantics

- **Insertion order.** Observers notify in the order they were added (MRI's
  Hash-keyed store). Re-adding an existing observer updates its method and keeps
  its position.
- **`add_observer` of a non-responding method raises.** When the supplied
  `respond_to?` callback reports false, `AddObserver` returns
  `*NotRespondingError` with MRI's verbatim message
  ``observer does not respond to `update'`` (rbgo turns it into a
  `NoMethodError`).
- **Changed-flag lifecycle.** `changed?` starts false; `Changed(true)` sets it,
  `Changed(false)` clears it.
- **Notify decision/reset.** `NotifyObservers` is a no-op returning `ok == false`
  when `changed?` is false; when changed it returns the observers to call and
  then resets `changed?` to false.

## Usage

```go
import "github.com/go-ruby-observer/observer"

var r observer.Registry

// add_observer(w)   /   add_observer(w2, :special)
_ = r.AddObserver(w, observer.DefaultFunc, respondTo)
_ = r.AddObserver(w2, "special", respondTo)

r.Changed(true) // changed
entries, args, ok := r.NotifyObservers("event")
if ok {
    for _, e := range entries {
        // rbgo: invoke e.Func on e.Observer with args
        dispatch(e.Observer, e.Func, args)
    }
}
// r.ChangedQ() == false now
```

## License

BSD-3-Clause.
