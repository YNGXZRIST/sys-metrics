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

func TestPool_Get_ReusesAfterPutWithoutRecreate(t *testing.T) {
	var created int
	p := New(func() *testItem {
		created++
		return &testItem{}
	})

	x := p.Get()
	if created != 1 {
		t.Fatalf("expected 1 created item after first Get(), got %d", created)
	}

	x.val = 10
	p.Put(x)

	y := p.Get()
	if created != 1 {
		t.Fatalf("expected Get() after Put() to reuse without creating new item, got created=%d", created)
	}
	if y.val != 0 {
		t.Fatalf("expected reused item to be reset (val=0), got %d", y.val)
	}
	if y.resetCount != 1 {
		t.Fatalf("expected Reset() to be called exactly once before reuse, got %d", y.resetCount)
	}
}
