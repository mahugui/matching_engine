package trade

import (
	"math/rand"
	"testing"
)

// A function signature allowing us to switch easily between min and max queues
type popperFun func(*testing.T, *Tree, *Tree, *prioq) (*Order, *Order, *Order)

var maker = NewOrderMaker()

func TestPushPopSimpleMin(t *testing.T) {
	// buys
	testPushPopSimple(t, 100, 1, 1, BUY, maxPopper)
	testPushPopSimple(t, 100, 10, 20, BUY, maxPopper)
	testPushPopSimple(t, 100, 100, 10000, BUY, maxPopper)
	// sells
	testPushPopSimple(t, 100, 1, 1, SELL, minPopper)
	testPushPopSimple(t, 100, 10, 20, SELL, minPopper)
	testPushPopSimple(t, 100, 100, 10000, SELL, minPopper)
}

func testPushPopSimple(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	priceTree := NewTree()
	guidTree := NewTree()
	validate(t, priceTree, guidTree)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	for i := 0; i < pushCount; i++ {
		o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
		priceTree.Push(&o.PriceNode)
		guidTree.Push(&o.GuidNode)
		validate(t, priceTree, guidTree)
		q.push(o)
	}
	for i := 0; i < pushCount; i++ {
		popCheck(t, priceTree, guidTree, q, popper)
	}
}

func TestRandomPushPop(t *testing.T) {
	// buys
	testPushPopRandom(t, 100, 1, 1, BUY, maxPopper)
	testPushPopRandom(t, 100, 10, 20, BUY, maxPopper)
	testPushPopRandom(t, 100, 100, 10000, BUY, maxPopper)
	// sells
	testPushPopRandom(t, 100, 1, 1, SELL, minPopper)
	testPushPopRandom(t, 100, 10, 20, SELL, minPopper)
	testPushPopRandom(t, 100, 100, 10000, SELL, minPopper)
}

func testPushPopRandom(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	priceTree := NewTree()
	guidTree := NewTree()
	validate(t, priceTree, guidTree)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || priceTree.PeekMin() == nil {
			o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
			priceTree.Push(&o.PriceNode)
			guidTree.Push(&o.GuidNode)
			validate(t, priceTree, guidTree)
			q.push(o)
			i++
		} else {
			popCheck(t, priceTree, guidTree, q, popper)
		}
	}
	for priceTree.PeekMin() == nil {
		po := priceTree.PopMax().O
		fo := q.popMax()
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		ensureFreed(t, po)
		validate(t, priceTree, guidTree)
	}
}

func TestAddRemoveSimple(t *testing.T) {
	// Buys
	testAddRemoveSimple(t, 100, 1, 1, BUY)
	testAddRemoveSimple(t, 100, 10, 20, BUY)
	testAddRemoveSimple(t, 100, 100, 10000, BUY)
	// Sells
	testAddRemoveSimple(t, 100, 1, 1, SELL)
	testAddRemoveSimple(t, 100, 10, 20, SELL)
	testAddRemoveSimple(t, 100, 100, 10000, SELL)
}

func testAddRemoveSimple(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind) {
	priceTree := NewTree()
	guidTree := NewTree()
	validate(t, priceTree, guidTree)
	orderMap := make(map[int64]*Order)
	for i := 0; i < pushCount; i++ {
		o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
		priceTree.Push(&o.PriceNode)
		guidTree.Push(&o.GuidNode)
		validate(t, priceTree, guidTree)
		orderMap[o.Guid()] = o
	}
	drainTree(t, priceTree, guidTree, orderMap)
}

func TestAddRemoveRandom(t *testing.T) {
	// Buys
	testAddRemoveRandom(t, 100, 1, 1, BUY)
	testAddRemoveRandom(t, 100, 10, 20, BUY)
	testAddRemoveRandom(t, 100, 100, 10000, BUY)
	// Sells
	testAddRemoveRandom(t, 100, 1, 1, SELL)
	testAddRemoveRandom(t, 100, 10, 20, SELL)
	testAddRemoveRandom(t, 100, 100, 10000, SELL)
}

func testAddRemoveRandom(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind) {
	priceTree := NewTree()
	guidTree := NewTree()
	validate(t, priceTree, guidTree)
	orderMap := make(map[int64]*Order)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || guidTree.PeekMin() == nil {
			o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
			priceTree.Push(&o.PriceNode)
			guidTree.Push(&o.GuidNode)
			validate(t, priceTree, guidTree)
			orderMap[o.Guid()] = o
			i++
		} else {
			for g, o := range orderMap {
				po := guidTree.Pop(g).O
				delete(orderMap, g)
				if po != o {
					t.Errorf("Bad pop")
				}
				ensureFreed(t, po)
				validate(t, priceTree, guidTree)
				break
			}
		}
	}
	drainTree(t, priceTree, guidTree, orderMap)
}

func drainTree(t *testing.T, priceTree, guidTree *Tree, orderMap map[int64]*Order) {
	for g := range orderMap {
		o := orderMap[g]
		po := guidTree.Pop(o.Guid()).O
		if po != o {
			t.Errorf("Bad pop")
		}
		ensureFreed(t, po)
		validate(t, priceTree, guidTree)
	}
}

func ensureFreed(t *testing.T, o *Order) {
	if !o.PriceNode.isFree() {
		t.Errorf("Price Node was not freed")
	}
	if !o.GuidNode.isFree() {
		t.Errorf("Guid Node was not freed")
	}
}

// Quick check to ensure the tree's internal structure is valid
func validate(t *testing.T, priceTree, guidTree *Tree) {
	checkStructure(t, priceTree.root)
	checkStructure(t, guidTree.root)
}

func checkStructure(t *testing.T, n *Node) {
	if n == nil {
		return
	}
	checkQueue(t, n)
	if *n.pp != n {
		t.Errorf("Parent pointer does not point to child node")
	}
	if n.left != nil {
		if n.val <= n.left.val {
			t.Errorf("Left value is greater than or equal to node value. Left value: %d Node value %d", n.left.val, n.val)
		}
		checkStructure(t, n.left)
	}
	if n.right != nil {
		if n.val >= n.right.val {
			t.Errorf("Right value is less than or equal to node value. Right value: %d Node value %d", n.right.val, n.val)
		}
		checkStructure(t, n.right)
	}
}

func checkQueue(t *testing.T, n *Node) {
	curr := n.next
	prev := n
	for curr != n {
		if curr.prev != prev {
			t.Errorf("Bad queue next/prev pair")
		}
		if curr.pp != nil {
			t.Errorf("Internal queue node with non-nil parent pointer")
		}
		if curr.left != nil {
			t.Errorf("Internal queue node has non-nil left child")
		}
		if curr.right != nil {
			t.Errorf("Internal queue node has non-nil right child")
		}
		if curr.O == nil {
			t.Errorf("Internal queue node has nil Order")
		}
		prev = curr
		curr = curr.next
	}
}

// Function to pop and peek and check that everything is in order
func popCheck(t *testing.T, priceTree, guidTree *Tree, q *prioq, popper popperFun) {
	peek, pop, check := popper(t, priceTree, guidTree, q)
	if pop != check {
		t.Errorf("Mismatched push/pop pair")
		return
	}
	if pop != peek {
		t.Errorf("Mismatched peek/pop pair")
		return
	}
	validate(t, priceTree, guidTree)
}

// Helper functions for popping either the max or the min from our queues
func maxPopper(t *testing.T, priceTree, guidTree *Tree, q *prioq) (peek, pop, check *Order) {
	peek = priceTree.PeekMax().O
	if !guidTree.Has(peek.Guid()) {
		t.Errorf("Guid tree does not contain peeked order")
	}
	pop = priceTree.PopMax().O
	if guidTree.Has(peek.Price()) {
		t.Errorf("Guid tree still contains popped order")
		return
	}
	check = q.popMax()
	ensureFreed(t, pop)
	return
}

func minPopper(t *testing.T, priceTree, guidTree *Tree, q *prioq) (peek, pop, check *Order) {
	peek = priceTree.PeekMin().O
	if !guidTree.Has(peek.Guid()) {
		t.Errorf("Guid tree does not contain peeked order")
	}
	pop = priceTree.PopMin().O
	check = q.popMin()
	ensureFreed(t, pop)
	return
}

// An easy to build priority queue
type prioq struct {
	prios               [][]*Order
	lowPrice, highPrice int64
}

func mkPrioq(size int, lowPrice, highPrice int64) *prioq {
	prios := make([][]*Order, highPrice-lowPrice+1)
	return &prioq{prios: prios, lowPrice: lowPrice, highPrice: highPrice}
}

func (q *prioq) push(o *Order) {
	idx := o.Price() - q.lowPrice
	prio := q.prios[idx]
	prio = append(prio, o)
	q.prios[idx] = prio
}

func (q *prioq) popMax() *Order {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMin() *Order {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) pop(i int) *Order {
	prio := q.prios[i]
	o := prio[0]
	prio = prio[1:]
	q.prios[i] = prio
	return o
}
