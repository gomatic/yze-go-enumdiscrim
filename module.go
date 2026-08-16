// Module locality: a const group is judged only where the analyzed module
// declares it, because only there is the domain the author's to close.
package enumdiscrim

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// localToModule reports whether pkg belongs to the analyzed module.
//
// The instrument is analysis.Pass.Module, the same one yze/ptrparam reads for
// its foreign-convention rule, and it is read the same way so the suite has one
// answer to "is this ours" rather than two.
//
// WITHOUT MODULE METADATA only the analyzed package itself counts as local. A
// driver that loads no module identity (analysistest, and any driver not
// requesting packages.NeedModule) cannot tell a sibling package of the analyzed
// module from a library, and the two readings differ in which direction the
// rule is wrong: treating the unknown as local reports a foreign group as the
// author's domain, treating it as foreign misses a group that is. This is a
// PROBE, so it takes the miss — a finding it never makes costs a review, and a
// finding it makes about go/token costs every review.
//
// A package-less named type is never local. Go declares no constant of such a
// type (the universe's named types are error, any and comparable), so this
// answers a case a real pass does not produce; it is here because the nil is
// reachable in the type system and Path() would panic on it, not because a
// comparison can arrive holding one.
func localToModule(pass *analysis.Pass, pkg *types.Package) bool {
	if pkg == nil {
		return false
	}
	if pkg == pass.Pkg {
		return true
	}
	if pass.Module == nil || pass.Module.Path == "" {
		return false
	}
	return pkg.Path() == pass.Module.Path || strings.HasPrefix(pkg.Path(), pass.Module.Path+"/")
}
