package proc

import "sort"

// ChildrenIndex maps each PPID to its child PIDs (each slice sorted ascending).
func ChildrenIndex(procs []Proc) map[int][]int {
	idx := make(map[int][]int)
	for _, p := range procs {
		idx[p.PPID] = append(idx[p.PPID], p.PID)
	}
	for ppid := range idx {
		sort.Ints(idx[ppid])
	}
	return idx
}

// Descendants returns a BFS set of processes rooted at rootPID, including the
// root when present. maxDepth <= 0 means default 16. Missing root returns an
// empty slice (no error). Child expansion order is ascending PID.
//
// depth(root)=0; children are expanded while depth < maxDepth.
func Descendants(rootPID int, procs []Proc, maxDepth int) []Proc {
	if maxDepth <= 0 {
		maxDepth = 16
	}

	byPID := make(map[int]Proc, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}
	if _, ok := byPID[rootPID]; !ok {
		return []Proc{}
	}

	children := ChildrenIndex(procs)

	type item struct {
		pid   int
		depth int
	}
	out := make([]Proc, 0)
	queue := []item{{pid: rootPID, depth: 0}}
	visited := map[int]bool{rootPID: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		out = append(out, byPID[cur.pid])
		if cur.depth >= maxDepth {
			continue
		}
		for _, childPID := range children[cur.pid] {
			if visited[childPID] {
				continue
			}
			if _, exists := byPID[childPID]; !exists {
				continue
			}
			visited[childPID] = true
			queue = append(queue, item{pid: childPID, depth: cur.depth + 1})
		}
	}
	return out
}
