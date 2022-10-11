package cwl

// Graphs represents "$graph" field in CWL.
type Graphs []*Root

// Graph represents an element of "steps"
type Graph struct {
	Run *Root
}

// Len for sorting
func (g Graphs) Len() int {
	return len(g)
}

// Less for sorting
func (g Graphs) Less(i, j int) bool {
	return g[i].Process.Base().ID < g[j].Process.Base().ID
}

// Swap for sorting
func (g Graphs) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
