// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-observer/observer authors

// Package observer is a pure-Go (CGO=0) implementation of the core state behind
// Ruby's Observable mixin — the observer registry and the "changed" flag — as
// specified by MRI Ruby's lib/observer.rb (verified against the reference
// interpreter, Ruby 4.0.5).
//
// A [Registry] backs one object that includes Observable. It owns the ordered
// set of observers (insertion order, as MRI's Hash-keyed store), the
// changed-flag lifecycle, and the notify decision/reset. It deliberately does
// not invoke observers: NotifyObservers returns the ordered (observer, method,
// args) tuples to call, and the actual method dispatch onto each observer stays
// in the caller (in go-embedded-ruby, the interpreter). Likewise responsiveness
// (Ruby's respond_to?) is delegated to a caller-supplied callback, so this
// package has no dependency on any interpreter.
//
// The MRI mapping is:
//
//	add_observer(obj, func=:update)  ->  Registry.AddObserver
//	delete_observer(obj)             ->  Registry.DeleteObserver
//	delete_observers                 ->  Registry.DeleteObservers
//	count_observers                  ->  Registry.CountObservers
//	changed(state=true)              ->  Registry.Changed
//	changed?                         ->  Registry.ChangedQ
//	notify_observers(*args)          ->  Registry.NotifyObservers
//
// add_observer raises NoMethodError when the observer does not respond to the
// method (reproduced as [*NotRespondingError]); notify_observers is a no-op
// returning nil when changed? is false, and otherwise notifies each observer in
// insertion order and then resets changed? to false.
package observer
