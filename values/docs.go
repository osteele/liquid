// Package values is an internal package that defines methods such as sorting, comparison, and type conversion, that apply to interface types.
//
// It is similar to, and makes heavy use of, the reflect package.
//
// Since the intent is to provide runtime services for the Liquid expression interpreter,
// this package does not implement "generic" generics.
// It attempts to implement Liquid semantics (which are largely Ruby semantics).
package values
