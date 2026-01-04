package exceltable

import (
	"iter"
	"maps"
	"sync"

	"github.com/xuri/excelize/v2"
)

var defaultRules = newRuleList()

type rule struct {
	name     string
	style    *excelize.Style
	styleID  int
	priority int
}

func (r *rule) clone() *rule {
	return &rule{
		name:     r.name,
		style:    r.style,
		styleID:  r.styleID,
		priority: r.priority,
	}
}

type ruleList struct {
	sync.Mutex
	v   []*rule        // NOTE: Rules are stored in descending order of priority.
	inv map[string]int // name -> index in v
}

func newRuleList() *ruleList {
	return &ruleList{
		v:   make([]*rule, 0, 2),
		inv: make(map[string]int),
	}
}

func (rl *ruleList) insert(r *rule) {
	rl.Lock()
	defer rl.Unlock()

	idx, ok := rl.inv[r.name]
	if ok {
		rl.v[idx] = r
	} else {
		rl.v = append(rl.v, r)
		idx = len(rl.v) - 1
		rl.inv[r.name] = idx
	}

	next := rl.sort(idx)
	for next != idx {
		idx = next
		next = rl.sort(idx)
	}
}

func (rl *ruleList) sort(i int) int {
	if 0 <= i-1 && rl.v[i-1].priority <= rl.v[i].priority {
		rl.swap(i-1, i)
		return i - 1
	}

	if i+1 < len(rl.v) && rl.v[i].priority < rl.v[i+1].priority {
		rl.swap(i, i+1)
		return i + 1
	}

	return i
}

func (rl *ruleList) swap(i, j int) {
	rl.inv[rl.v[i].name], rl.inv[rl.v[j].name] = j, i
	rl.v[i], rl.v[j] = rl.v[j], rl.v[i]
}

func (rl *ruleList) clone() *ruleList {
	rl.Lock()
	defer rl.Unlock()

	nrl := &ruleList{
		v:   make([]*rule, len(rl.v)),
		inv: maps.Clone(rl.inv),
	}

	for i, r := range rl.v {
		nrl.v[i] = r.clone()
	}

	return nrl
}

func (rl *ruleList) iter() iter.Seq[*rule] {
	return func(yield func(*rule) bool) {
		rl.Lock()
		defer rl.Unlock()

		for _, r := range rl.v {
			if !yield(r) {
				break
			}
		}
	}
}

// RegisterDefaultRule registers a new default rule with the given tag name, *excelize.Style, and priority.
// Rules with higher priority values are applied earlier:
//
//	exceltable.RegisterRule("customTag", &excelize.Style{ ... }, 0)
func RegisterDefaultRule(name string, style *excelize.Style, priority int) {
	defaultRules.insert(&rule{
		name:     name,
		style:    style,
		styleID:  -1,
		priority: priority,
	})
}
