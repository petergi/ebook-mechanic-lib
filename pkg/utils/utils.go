// Package utils provides small helper functions.
//
//revive:disable-next-line var-naming
package utils

// StringInSlice reports whether str is present in slice.
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
