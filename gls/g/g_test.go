package g

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestG1(t *testing.T) {
	gp1 := G()
	assert.True(t, gp1 != nil)

	t.Run("G in another goroutine", func(t *testing.T) {
		gp2 := G()
		assert.True(t, gp2 != nil)
		assert.True(t, gp1 != gp2)
	})

	gType := reflect.TypeOf(G0())
	sf, ss := gType.FieldByName("labels")
	assert.True(t, ss && sf.Offset > 0)
}

func TestG2(t *testing.T) {
	gp1 := G()

	if gp1 == nil {
		t.Fatalf("fail to get G.")
	}

	t.Run("G in another goroutine", func(t *testing.T) {
		gp2 := G()

		if gp2 == nil {
			t.Fatalf("fail to get G.")
		}

		if gp2 == gp1 {
			t.Fatalf("every living G must be different. [gp1:%p] [gp2:%p]", gp1, gp2)
		}
	})
}
