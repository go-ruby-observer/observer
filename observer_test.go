// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-observer/observer authors

package observer

import (
	"errors"
	"reflect"
	"testing"
)

// obs is a stand-in for a Ruby observer object: a comparable identity plus the
// set of method names it "responds to".
type obs struct {
	id      int
	methods map[string]bool
}

func newObs(id int, methods ...string) *obs {
	m := make(map[string]bool, len(methods))
	for _, name := range methods {
		m[name] = true
	}
	return &obs{id: id, methods: m}
}

// respondTo is a [RespondTo] that consults the obs method set.
func respondTo(o Observer, name string) bool {
	ob, ok := o.(*obs)
	if !ok {
		return false
	}
	return ob.methods[name]
}

func ids(entries []Entry) []int {
	out := make([]int, len(entries))
	for i, e := range entries {
		out[i] = e.Observer.(*obs).id
	}
	return out
}

func TestDefaultFunc(t *testing.T) {
	if DefaultFunc != "update" {
		t.Fatalf("DefaultFunc = %q, want %q", DefaultFunc, "update")
	}
}

func TestZeroRegistry(t *testing.T) {
	var r Registry
	if got := r.CountObservers(); got != 0 {
		t.Fatalf("fresh CountObservers = %d, want 0", got)
	}
	if r.ChangedQ() {
		t.Fatal("fresh ChangedQ = true, want false")
	}
}

func TestAddObserverDefaultAndCustomFunc(t *testing.T) {
	var r Registry
	w1 := newObs(1, "update")
	w2 := newObs(2, "special")

	if err := r.AddObserver(w1, DefaultFunc, respondTo); err != nil {
		t.Fatalf("AddObserver default: %v", err)
	}
	if err := r.AddObserver(w2, "special", respondTo); err != nil {
		t.Fatalf("AddObserver custom: %v", err)
	}
	if got := r.CountObservers(); got != 2 {
		t.Fatalf("CountObservers = %d, want 2", got)
	}

	r.Changed(true)
	entries, _, ok := r.NotifyObservers()
	if !ok {
		t.Fatal("NotifyObservers ok = false, want true")
	}
	want := []Entry{{Observer: w1, Func: "update"}, {Observer: w2, Func: "special"}}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("entries = %+v, want %+v", entries, want)
	}
}

func TestAddObserverReAddSameObject(t *testing.T) {
	var r Registry
	w1 := newObs(1, "update", "other")
	w2 := newObs(2, "update")

	_ = r.AddObserver(w1, "update", respondTo)
	_ = r.AddObserver(w2, "update", respondTo)
	// Re-add w1 with a different method: count stays 2, position unchanged,
	// method updated.
	_ = r.AddObserver(w1, "other", respondTo)
	if got := r.CountObservers(); got != 2 {
		t.Fatalf("CountObservers after re-add = %d, want 2", got)
	}

	r.Changed(true)
	entries, _, _ := r.NotifyObservers()
	want := []Entry{{Observer: w1, Func: "other"}, {Observer: w2, Func: "update"}}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("entries = %+v, want %+v", entries, want)
	}
}

func TestAddObserverNotResponding(t *testing.T) {
	var r Registry
	bad := newObs(1) // responds to nothing

	err := r.AddObserver(bad, DefaultFunc, respondTo)
	var nre *NotRespondingError
	if !errors.As(err, &nre) {
		t.Fatalf("err = %v, want *NotRespondingError", err)
	}
	if nre.Func != "update" {
		t.Fatalf("NotRespondingError.Func = %q, want update", nre.Func)
	}
	// Verbatim MRI message.
	if got := err.Error(); got != "observer does not respond to `update'" {
		t.Fatalf("Error() = %q", got)
	}
	if got := r.CountObservers(); got != 0 {
		t.Fatalf("CountObservers after failed add = %d, want 0", got)
	}
}

func TestAddObserverCustomFuncNotResponding(t *testing.T) {
	var r Registry
	w := newObs(1, "update") // responds to update but not nope

	err := r.AddObserver(w, "nope", respondTo)
	var nre *NotRespondingError
	if !errors.As(err, &nre) || nre.Func != "nope" {
		t.Fatalf("err = %v, want NotRespondingError{nope}", err)
	}
	if got := err.Error(); got != "observer does not respond to `nope'" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestAddObserverNilRespondToSkipsCheck(t *testing.T) {
	var r Registry
	bad := newObs(1) // responds to nothing
	if err := r.AddObserver(bad, DefaultFunc, nil); err != nil {
		t.Fatalf("AddObserver with nil respondTo: %v", err)
	}
	if got := r.CountObservers(); got != 1 {
		t.Fatalf("CountObservers = %d, want 1", got)
	}
}

func TestDeleteObserver(t *testing.T) {
	var r Registry
	w1, w2, w3 := newObs(1, "update"), newObs(2, "update"), newObs(3, "update")
	_ = r.AddObserver(w1, "update", respondTo)
	_ = r.AddObserver(w2, "update", respondTo)
	_ = r.AddObserver(w3, "update", respondTo)

	r.DeleteObserver(w2)
	if got := r.CountObservers(); got != 2 {
		t.Fatalf("CountObservers after delete = %d, want 2", got)
	}

	// Deleting an unregistered observer is a no-op.
	r.DeleteObserver(newObs(99, "update"))
	if got := r.CountObservers(); got != 2 {
		t.Fatalf("CountObservers after no-op delete = %d, want 2", got)
	}

	r.Changed(true)
	entries, _, _ := r.NotifyObservers()
	if got := ids(entries); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("order after delete = %v, want [1 3]", got)
	}
}

func TestDeleteObservers(t *testing.T) {
	var r Registry
	_ = r.AddObserver(newObs(1, "update"), "update", respondTo)
	_ = r.AddObserver(newObs(2, "update"), "update", respondTo)
	r.DeleteObservers()
	if got := r.CountObservers(); got != 0 {
		t.Fatalf("CountObservers after DeleteObservers = %d, want 0", got)
	}
	// Idempotent.
	r.DeleteObservers()
	if got := r.CountObservers(); got != 0 {
		t.Fatalf("CountObservers after second DeleteObservers = %d, want 0", got)
	}
}

func TestNotifyOrderingAfterDeleteReAdd(t *testing.T) {
	// MRI: re-adding after delete appends to the end (Hash insertion order).
	var r Registry
	ws := make([]*obs, 5)
	for i := range ws {
		ws[i] = newObs(i+1, "update")
		_ = r.AddObserver(ws[i], "update", respondTo)
	}
	r.DeleteObserver(ws[2]) // remove id 3
	_ = r.AddObserver(ws[2], "update", respondTo)

	r.Changed(true)
	entries, _, _ := r.NotifyObservers()
	if got := ids(entries); !reflect.DeepEqual(got, []int{1, 2, 4, 5, 3}) {
		t.Fatalf("order = %v, want [1 2 4 5 3]", got)
	}
}

func TestChangedFlagLifecycle(t *testing.T) {
	var r Registry
	if r.ChangedQ() {
		t.Fatal("initial ChangedQ = true, want false")
	}
	r.Changed(true)
	if !r.ChangedQ() {
		t.Fatal("after Changed(true) ChangedQ = false, want true")
	}
	r.Changed(false)
	if r.ChangedQ() {
		t.Fatal("after Changed(false) ChangedQ = true, want false")
	}
}

func TestNotifyNoChangeIsNoOp(t *testing.T) {
	var r Registry
	w := newObs(1, "update")
	_ = r.AddObserver(w, "update", respondTo)

	// changed? is false -> no-op, nil entries, ok == false, flag stays false.
	entries, args, ok := r.NotifyObservers("a")
	if ok {
		t.Fatal("NotifyObservers ok = true with no change, want false")
	}
	if entries != nil || args != nil {
		t.Fatalf("NotifyObservers returned (%v, %v), want (nil, nil)", entries, args)
	}
	if r.ChangedQ() {
		t.Fatal("ChangedQ = true after no-op notify, want false")
	}
}

func TestNotifyResetsChangedAndForwardsArgs(t *testing.T) {
	var r Registry
	w := newObs(1, "update")
	_ = r.AddObserver(w, "update", respondTo)

	r.Changed(true)
	entries, args, ok := r.NotifyObservers(1, "two", 3.0)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	wantArgs := []any{1, "two", 3.0}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args = %v, want %v", args, wantArgs)
	}
	if len(entries) != 1 || entries[0].Observer != Observer(w) || entries[0].Func != "update" {
		t.Fatalf("entries = %+v", entries)
	}
	// Flag reset after a successful notify.
	if r.ChangedQ() {
		t.Fatal("ChangedQ = true after notify, want false")
	}
	// A second notify without a new Changed is a no-op.
	if _, _, ok := r.NotifyObservers(); ok {
		t.Fatal("second NotifyObservers ok = true, want false")
	}
}

func TestNotifyChangedNoObservers(t *testing.T) {
	var r Registry
	r.Changed(true)
	entries, _, ok := r.NotifyObservers()
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if len(entries) != 0 {
		t.Fatalf("entries = %+v, want empty", entries)
	}
	if r.ChangedQ() {
		t.Fatal("ChangedQ = true after notify, want false")
	}
}

func TestNotRespondingErrorIs(t *testing.T) {
	err := error(&NotRespondingError{Func: "update"})
	var nre *NotRespondingError
	if !errors.As(err, &nre) {
		t.Fatal("errors.As failed for *NotRespondingError")
	}
}
