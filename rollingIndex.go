// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/01/2021

package reqtest

func newRollingIndex(size int) rollingIndex {
	return rollingIndex{
		size:    size - 1,
		curr:    -1,
		hasFull: false,
	}
}

type rollingIndex struct {
	hasFull bool
	size    int
	curr    int
}

func (r rollingIndex) Reset(size int) rollingIndex {
	r.hasFull = false
	if size > 0 {
		r.size = size - 1
	}
	r.curr = -1
	return r
}

func (r rollingIndex) List() (out []int) {
	if r.hasFull {
		for i:=r.curr+1; i <= r.size; i++ {
			out = append(out, i)
		}
	}
	for i:=0; i<r.curr+1; i++ {
		out = append(out, i)
	}
	return out
}

func (r rollingIndex) Curr() int {
	return r.curr
}

func (r rollingIndex) Next() rollingIndex {
	if r.curr >= r.size {
		r.curr = 0
		r.hasFull = true
		return r
	}
	r.curr += 1
	return r
}
