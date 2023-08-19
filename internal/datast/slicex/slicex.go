// Package slicex defines various functions useful with slices of any type.
// This is an extension to the slices package in the Go's standard library.
package slicex

// Delete removes the elements s[i:j] from s, modifies s in place.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers Delete is zeroing those elements so that
// objects they reference will be garbage collected.
func Delete[S ~[]E, E any](s *S, i, j int) {
	clear((*s)[i:j])
	*s = append((*s)[:i], (*s)[j:]...)
}
