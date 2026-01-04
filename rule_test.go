package exceltable

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func Test_ruleList_iter(t *testing.T) {
	rules := newRuleList()

	size := 1000
	var maxPriority int = 1e9
	mp := make(map[string]int)
	verify := func() {
		prePriority := maxPriority
		preNo := -1
		for r := range rules.iter() {
			no, ok := mp[r.name]
			require.True(t, ok)

			assert.True(t, r.priority <= prePriority)
			if r.priority == prePriority {
				assert.True(t, no > preNo)
			}

			prePriority = r.priority
			preNo = no
		}
	}

	for i := range size {
		r := &rule{
			name:     uuid.NewString(),
			priority: rand.Intn(1e9),
		}
		rules.insert(r)
		mp[r.name] = i
	}

	verify()

	// Update existing rules.
	next := 1000
	for name := range mp {
		r := &rule{
			name:     name,
			priority: rand.Intn(1e9),
		}
		rules.insert(r)
		mp[name] = next
		next++
	}

	verify()
}

func TestRegisterDefaultRule(t *testing.T) {
	RegisterDefaultRule("test", &excelize.Style{}, -1000)
}
