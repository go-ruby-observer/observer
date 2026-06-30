// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-observer/observer authors

package observer

// DefaultFunc is the method name Observable#add_observer uses when the caller
// does not pass an explicit one, matching MRI's default (`:update`).
const DefaultFunc = "update"

// Observer is one registered observer together with the name of the method to
// invoke on it. Observer is an opaque interface value supplied by the caller
// (in go-embedded-ruby it is the underlying Ruby object); this package never
// inspects it beyond identity, so it is used as a map key and must therefore be
// comparable.
type Observer = any

// Entry is a registered (observer, method) pair as returned by
// [Registry.NotifyObservers]. The caller (rbgo) performs the actual dispatch by
// invoking Func on Observer with the notification arguments.
type Entry struct {
	// Observer is the registered observer object.
	Observer Observer
	// Func is the name of the method to call on Observer (the value passed to
	// AddObserver, or DefaultFunc).
	Func string
}

// RespondTo reports whether observer responds to the method named name. The
// caller supplies this because responsiveness is an interpreter concept (Ruby's
// respond_to?); this package is interpreter-agnostic. AddObserver consults it to
// reproduce MRI's behaviour of raising NoMethodError for an observer that does
// not respond to the requested method.
type RespondTo func(observer Observer, name string) bool

// NotRespondingError is returned by [Registry.AddObserver] when the observer
// does not respond to the requested method. It mirrors the NoMethodError MRI's
// Observable#add_observer raises; rbgo translates it back into a Ruby
// NoMethodError. Its Error string matches MRI's message verbatim:
//
//	observer does not respond to `update'
type NotRespondingError struct {
	// Func is the method name the observer failed to respond to.
	Func string
}

func (e *NotRespondingError) Error() string {
	return "observer does not respond to `" + e.Func + "'"
}

// Registry is the state backing one object that includes Ruby's Observable
// mixin: the ordered set of observers and the "changed" flag. A zero Registry is
// ready to use and has no observers with changed? == false.
//
// Registry owns the observer set, insertion ordering, the changed-flag
// lifecycle, and the notify decision/reset; it does not invoke observers. The
// method dispatch onto each observer stays in the caller (rbgo), which acts on
// the [Entry] list [NotifyObservers] returns.
//
// Registry is not safe for concurrent use; callers must synchronise, matching
// MRI's Observable (which is likewise not thread-safe).
type Registry struct {
	// order preserves insertion order of observers, matching MRI where the
	// observer set is a Hash keyed by observer (insertion-ordered).
	order []Observer
	// funcs maps each observer to the method name to invoke on it.
	funcs   map[Observer]string
	changed bool
}

// AddObserver registers observer, recording func name as the method to invoke
// on notification. fn is the method name (use [DefaultFunc] for MRI's default of
// :update). respondTo, when non-nil, is consulted to verify observer responds to
// fn; if it does not, observer is not registered and a [*NotRespondingError] is
// returned, mirroring MRI's NoMethodError. A nil respondTo skips the check (the
// caller is then responsible for responsiveness).
//
// Re-adding an already-registered observer updates its method name and leaves
// its position in the order unchanged, matching MRI's Hash-keyed store.
func (r *Registry) AddObserver(observer Observer, fn string, respondTo RespondTo) error {
	if respondTo != nil && !respondTo(observer, fn) {
		return &NotRespondingError{Func: fn}
	}
	if r.funcs == nil {
		r.funcs = make(map[Observer]string)
	}
	if _, ok := r.funcs[observer]; !ok {
		r.order = append(r.order, observer)
	}
	r.funcs[observer] = fn
	return nil
}

// DeleteObserver removes observer from the set. Removing an observer that is not
// registered is a no-op, matching MRI's Hash#delete.
func (r *Registry) DeleteObserver(observer Observer) {
	if _, ok := r.funcs[observer]; !ok {
		return
	}
	delete(r.funcs, observer)
	for i, o := range r.order {
		if o == observer {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// DeleteObservers removes every observer from the set.
func (r *Registry) DeleteObservers() {
	r.order = nil
	r.funcs = nil
}

// CountObservers returns the number of registered observers.
func (r *Registry) CountObservers() int {
	return len(r.order)
}

// Changed sets the changed flag to state, matching MRI's Observable#changed
// (whose argument defaults to true). Pass true to mark the state changed, false
// to clear it.
func (r *Registry) Changed(state bool) {
	r.changed = state
}

// ChangedQ reports whether the changed flag is set, matching MRI's
// Observable#changed?.
func (r *Registry) ChangedQ() bool {
	return r.changed
}

// NotifyObservers decides whether observers should be notified, following MRI's
// Observable#notify_observers exactly:
//
//   - When changed? is false it is a no-op: it returns a nil [Entry] list and
//     leaves the changed flag false. (ok reports false in this case.)
//   - When changed? is true it returns the registered observers in insertion
//     order, each paired with its method name, then clears the changed flag.
//     (ok reports true.)
//
// args are the notification arguments; they are not used by this package and are
// returned to the caller unchanged so it can forward them to each observer's
// method. The caller (rbgo) performs the dispatch — it invokes each Entry.Func
// on each Entry.Observer with args — which is why the args are echoed here rather
// than acted upon.
//
// ok mirrors MRI's truthiness: MRI returns nil when not changed and the result
// of changed(false) (i.e. false) when it did notify; callers that need the
// notify decision should use ok rather than len(entries).
func (r *Registry) NotifyObservers(args ...any) (entries []Entry, notifyArgs []any, ok bool) {
	if !r.changed {
		return nil, nil, false
	}
	entries = make([]Entry, len(r.order))
	for i, o := range r.order {
		entries[i] = Entry{Observer: o, Func: r.funcs[o]}
	}
	r.changed = false
	return entries, args, true
}
