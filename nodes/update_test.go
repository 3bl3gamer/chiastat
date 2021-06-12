package nodes

import (
	"chiastat/chia/network"
	"testing"
)

func TestConnList(t *testing.T) {
	assertEq := func(a interface{}, b interface{}) {
		if a != b {
			t.Errorf("%v != %v", a, b)
		}
	}
	assertNilItem := func(a *ConnListItem) {
		if a != nil {
			t.Errorf("%v != nil", a)
		}
	}
	assertTwoItems := func(list *ConnList, c0, c1 *network.WSChiaConnection) {
		assertEq(list.length, int64(2))
		assertEq(list.start.conn, c0)
		assertEq(list.end.conn, c1)
		assertNilItem(list.start.prev)
		assertEq(list.start.next, list.end)
		assertNilItem(list.end.next)
		assertEq(list.end.prev, list.start)
	}
	assertOneItem := func(list *ConnList, c0 *network.WSChiaConnection) {
		assertEq(list.length, int64(1))
		assertEq(list.start.conn, c0)
		assertEq(list.end.conn, c0)
		assertNilItem(list.start.prev)
		assertNilItem(list.start.next)
	}
	assertEmpty := func(list *ConnList) {
		assertEq(list.length, int64(0))
		assertNilItem(list.start)
		assertNilItem(list.end)
	}

	var list *ConnList
	c0 := &network.WSChiaConnection{}
	c1 := &network.WSChiaConnection{}
	c2 := &network.WSChiaConnection{}

	fill := func() {
		list = &ConnList{limit: 2}
		assertEmpty(list)

		list.PushConn(c0)
		assertOneItem(list, c0)

		list.PushConn(c1)
		assertTwoItems(list, c0, c1)

		list.PushConn(c2)
		assertEq(list.length, int64(3))
		assertEq(list.start.conn, c0)
		assertEq(list.start.next.conn, c1)
		assertEq(list.end.conn, c2)
	}

	fill()
	dest := list.start
	assertEq(list.ShiftIfNeed(), dest)
	assertTwoItems(list, c1, c2)

	assertNilItem(list.ShiftIfNeed())
	assertTwoItems(list, c1, c2)

	// ---

	fill()
	list.DelItem(list.start.next)
	assertTwoItems(list, c0, c2)

	list.DelItem(list.start)
	assertOneItem(list, c2)

	list.DelItem(list.start)
	assertEmpty(list)

	fill()
	list.DelItem(list.end)
	assertTwoItems(list, c0, c1)
}
