// Package b stands in for a library the analyzed module merely consumes: its
// const group is declared here, so a comparison written in package a
// discriminates a domain a's author does not govern.
package b

// Tag is a three-member group declared OUTSIDE the package under analysis —
// the same shape as go/token.Token, at a size the threshold accepts, so the
// only thing that can be keeping a's comparison silent is where Tag lives.
type Tag int

// The tags.
const (
	TagOne Tag = iota
	TagTwo
	TagThree
)
