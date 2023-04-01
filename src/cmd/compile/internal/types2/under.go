// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

// under returns the true expanded underlying type.
// If it doesn't exist, the result is Typ[Invalid].
// under must only be called when a type is known
// to be fully set up.
func under(t Type) Type {
	if t, _ := t.(*Named); t != nil {
		return t.under()
	}
	return t.Underlying()
}

// If t is not a type parameter, coreType returns the underlying type.
// If t is a type parameter, coreType returns the single underlying
// type of all types in its type set if it exists, or nil otherwise. If the
// type set contains only unrestricted and restricted channel types (with
// identical element types), the single underlying type is the restricted
// channel type if the restrictions are always the same, or nil otherwise.
func coreType(t Type) Type {
	tpar, _ := t.(*TypeParam)
	if tpar == nil {
		return under(t)
	}

	var su Type
	if tpar.underIs(func(u Type) bool {
		if u == nil {
			return false
		}
		if su != nil {
			u = match(su, u)
			if u == nil {
				return false
			}
		}
		// su == nil || match(su, u) != nil
		su = u
		return true
	}) {
		return su
	}
	return nil
}

// coreString is like coreType but also considers []byte
// and strings as identical. In this case, if successful and we saw
// a string, the result is of type (possibly untyped) string.
func coreString(t Type) Type {
	tpar, _ := t.(*TypeParam)
	if tpar == nil {
		return under(t) // string or untyped string
	}

	var su Type
	hasString := false
	if tpar.underIs(func(u Type) bool {
		if u == nil {
			return false
		}
		if isString(u) {
			u = NewSlice(universeByte)
			hasString = true
		}
		if su != nil {
			u = match(su, u)
			if u == nil {
				return false
			}
		}
		// su == nil || match(su, u) != nil
		su = u
		return true
	}) {
		if hasString {
			return Typ[String]
		}
		return su
	}
	return nil
}

// If x and y are identical, match returns x.
// If x and y are identical channels but for their direction
// and one of them is unrestricted, match returns the channel
// with the restricted direction.
// In all other cases, match returns nil.
func match(x, y Type) Type {
	// Common case: we don't have channels.
	if Identical(x, y) {
		return x
	}

	// We may have channels that differ in direction only.
	if x, _ := x.(*Chan); x != nil {
		if y, _ := y.(*Chan); y != nil && Identical(x.elem, y.elem) {
			// We have channels that differ in direction only.
			// If there's an unrestricted channel, select the restricted one.
			switch {
			case x.dir == SendRecv:
				return y
			case y.dir == SendRecv:
				return x
			}
		}
	}

	// types are different
	return nil
}
