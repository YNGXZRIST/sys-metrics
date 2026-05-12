package pool

import "testing"

type testItem struct {
	val        int
	resetCount int
}

func (t *testItem) Reset() {
	t.val = 0
	t.resetCount++
}

func TestPool_Get_CreatesWithFactory(t *testing.T) {
	var created int
	p := New(func() *testItem {
		created++
		return &testItem{val: 123}
	})

	a := p.Get()
	if a == nil {
		t.Fatalf("Get() returned nil")
	}
	if created != 1 {
		t.Fatalf("expected 1 created item, got %d", created)
	}
	if a.val != 123 {
		t.Fatalf("expected initial val=123, got %d", a.val)
	}

	b := p.Get()
	if b == nil {
		t.Fatalf("Get() returned nil on second call")
	}
	if created != 2 {
		t.Fatalf("expected 2 created items, got %d", created)
	}
}

func TestPool_Put_ResetsBeforeStoring(t *testing.T) {
	p := New(func() *testItem { return &testItem{} })

	x := &testItem{val: 777}
	p.Put(x)

	if x.val != 0 {
		t.Fatalf("expected item to be reset (val=0), got %d", x.val)
	}
	if x.resetCount != 1 {
		t.Fatalf("expected Reset() to be called once, got %d", x.resetCount)
	}
}

func TestPool_Get_AfterPut_ReturnsZeroedItem(t *testing.T) {
	var created int
	p := New(func() *testItem {
		created++
		return &testItem{}
	})

	x := p.Get()
	origCreated := created
	if origCreated < 1 {
		t.Fatalf("expected factory to run at least once, got created=%d", created)
	}

	x.val = 10
	p.Put(x)

	y := p.Get()
	if y.val != 0 {
		t.Fatalf("expected Get after Put to yield val=0 (reset on Put or fresh factory value), got %d", y.val)
	}
	// sync.Pool does not guarantee reuse; a new factory object has resetCount 0, a reused one was Reset on Put.
	if created == origCreated {
		if y.resetCount < 1 {
			t.Fatalf("reused pooled item: expected Reset on Put, resetCount=%d", y.resetCount)
		}
	} else if created == origCreated+1 {
		if y.resetCount != 0 {
			t.Fatalf("new factory item: expected resetCount=0, got %d", y.resetCount)
		}
	} else {
		t.Fatalf("unexpected factory calls: before=%d after=%d", origCreated, created)
	}
}
