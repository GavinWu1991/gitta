package git

import (
	"strings"

	"github.com/gavin/gitta/internal/core"
	"github.com/go-git/go-git/v5/plumbing"
)

// branchFromRef converts a go-git reference into a domain Branch model.
func branchFromRef(ref *plumbing.Reference, isCurrent bool) core.Branch {
	name := ref.Name()

	branch := core.Branch{
		Name:       name.Short(),
		IsCurrent:  isCurrent,
		CommitHash: ref.Hash().String(),
	}

	if name.IsRemote() {
		branch.Type = core.BranchTypeRemote
		branch.RemoteName = remoteNameFromRef(name)
	} else {
		branch.Type = core.BranchTypeLocal
	}

	return branch
}

func remoteNameFromRef(name plumbing.ReferenceName) string {
	const prefix = "refs/remotes/"
	if !strings.HasPrefix(string(name), prefix) {
		return ""
	}
	rest := strings.TrimPrefix(string(name), prefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
