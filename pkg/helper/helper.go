/*
Package helper provides utility functions for creating pointers to common types.
This is particularly useful for handling optional fields in data structures where
nil values indicate the absence of a value.

Functions in this package are designed to facilitate the construction of pointers
to basic types, making it easier to work with data models that use pointer semantics.

The package includes functions for creating pointers to strings, integers, float64,
and a generic function for creating pointers to any type.
*/
package helper

// PtrStr returns a pointer to the given string.
//
// This function is a convenience helper for creating pointers to string values.
// It is commonly used when working with structs that have optional string fields,
// where a *string can be used to represent the presence or absence of a value
// (nil indicating absence).
//
// Parameters:
// - s (string): The string value to which a pointer will be created.
//
// Returns:
// - *string: A pointer to the given string value.
func PtrStr(s string) *string {
	return &s
}

// PtrInt returns a pointer to the given integer.
//
// This function simplifies the creation of integer pointers.  It is useful
// for representing optional integer fields in data structures.  A nil *int
// can signify that the integer value is not present.
//
// Parameters:
// - i (int): The integer value to which a pointer will be created.
//
// Returns:
// - *int: A pointer to the given integer value.
func PtrInt(i int) *int {
	return &i
}

// PtrFloat returns a pointer to the given float64.
//
// This function simplifies the creation of float pointers. It is useful
// for representing optional float64 fields in data structures. A nil *float64
// can signify that the integer value is not present.
//
// Parameters:
// - f (float64): The float64 value to which a pointer will be created.
//
// Returns:
// - *float64: A pointer to the given float64 value.
func PtrFloat(f float64) *float64 {
	return &f
}

// PtrUnit returns a pointer to the given value of any type T.
//
// This is a generic function that simplifies the creation of pointers for any type.
// It is particularly useful for managing optional fields in data structures where
// a nil pointer indicates the absence of a value. This function can replace
// type-specific pointer creation functions, providing a more flexible and reusable
// solution.
//
// Parameters:
// - unit (T): The value of any type T to which a pointer will be created.
//
// Returns:
// - *T: A pointer to the given value of type T.
func PtrUnit[T any](unit T) *T {
	return &unit
}
