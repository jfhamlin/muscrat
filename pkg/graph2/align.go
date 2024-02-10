package graph2

import (
	"reflect"

	"github.com/glojurelang/glojure/pkg/lang"
)

type (
	// GraphAlignment is a struct that represents the alignment of two
	// graphs.
	GraphAlignment struct {
		NodeIdentities map[NodeID]NodeID
	}
)

func AlignGraphs(a, b *Graph) GraphAlignment {
	identities := map[NodeID]NodeID{}

	// these are the nodes not yet matched
	aNodes := make([]*Node, 0, len(a.Nodes))
	bNodes := make([]*Node, 0, len(b.Nodes))

	// all constants are identical
	aConstIDs := map[float64]NodeID{}
	for _, n := range a.Nodes {
		if n.Type == "const" {
			aConstIDs[lang.First(n.Args).(float64)] = n.ID
		} else {
			aNodes = append(aNodes, n)
		}
	}
	// map constant nodes in b to constant nodes in a
	for _, n := range b.Nodes {
		if n.Type == "const" {
			if id, ok := aConstIDs[lang.First(n.Args).(float64)]; ok {
				identities[n.ID] = id
			} // else no possible match, so don't add to bNodes
		} else {
			bNodes = append(bNodes, n)
		}
	}

	// todo: match by key

	// now, use the levenshtein distance algorithm to match the remaining nodes
	// nodes are identical if their type and args are equal
	type dist struct {
		dist int
		eq   bool
	}
	// create a grid of distances
	//
	// the distance between two nodes is the minimum number of
	// insertions, deletions, and substitutions to transform one
	// sequence of nodes into the other.
	//
	// the rows of the grid are the nodes of a, and the columns are the
	// nodes of b.
	//
	// dist(a[:i], b[:j]) is in grid[i][j]
	grid := make([][]dist, len(aNodes)+1)
	for i := range grid {
		grid[i] = make([]dist, len(bNodes)+1)
	}
	for i := range grid {
		grid[i][0] = dist{dist: i}
	}
	for i := range grid[0] {
		grid[0][i] = dist{dist: i}
	}
	for i := 1; i <= len(aNodes); i++ {
		for j := 1; j <= len(bNodes); j++ {
			if aNodes[i-1].Type == bNodes[j-1].Type && seqsEqual(aNodes[i-1].Args, bNodes[j-1].Args) {
				grid[i][j] = grid[i-1][j-1]
				grid[i][j].eq = true
			} else {
				grid[i][j] = dist{dist: 1 + min3(
					grid[i-1][j].dist,
					grid[i][j-1].dist,
					grid[i-1][j-1].dist,
				)}
			}
		}
	}
	// now collect the matches
	for i, j := len(aNodes), len(bNodes); i > 0 && j > 0; {
		if grid[i][j].eq {
			identities[bNodes[j-1].ID] = aNodes[i-1].ID
			i--
			j--
		} else if grid[i][j].dist == grid[i-1][j].dist+1 {
			i--
		} else if grid[i][j].dist == grid[i][j-1].dist+1 {
			j--
		} else {
			i--
			j--
		}
	}

	return GraphAlignment{
		NodeIdentities: identities,
	}
}

func seqToSlice(s any) []any {
	var res []any
	for s := lang.Seq(s); s != nil; s = lang.Next(s) {
		res = append(res, lang.First(s))
	}
	return res
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// seqsEqual returns true if the two sequences are equal.
// We wrap lang.Equals to add support for comparing slices, which
// are not comparable in Glojure.
func seqsEqual(a, b any) bool {
	seqA := lang.Seq(a)
	seqB := lang.Seq(b)
	for {
		if seqA == nil {
			return seqB == nil
		}
		if seqB == nil {
			return false
		}
		firstA := lang.First(seqA)
		firstB := lang.First(seqB)
		kindA := reflect.TypeOf(firstA).Kind()
		kindB := reflect.TypeOf(firstB).Kind()
		if kindA == reflect.Slice || kindB == reflect.Slice {
			if kindA != kindB {
				return false
			}
			if !reflect.DeepEqual(firstA, firstB) {
				return false
			}
		} else {
			if !lang.Equals(firstA, firstB) {
				return false
			}
		}

		seqA = lang.Next(seqA)
		seqB = lang.Next(seqB)
	}
	return true
}
