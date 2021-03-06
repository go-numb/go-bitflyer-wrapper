package orders

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/realtime/jsonrpc"
	"gonum.org/v1/gonum/stat"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/types"
)

// Managed is orders/positions struct
type Managed struct {
	Orders    *Orders
	Cancels   *Orders
	Positions *Orders

	Result chan string
}

func New() *Managed {
	return &Managed{

		Orders:    new(Orders),
		Cancels:   new(Orders),
		Positions: new(Orders),
	}
}

type Orders struct {
	m sync.Map
}

type StatusType int

const (
	NotExist StatusType = iota
	OnBoard
	Partial
	Completed
	Canceled
	Expired
)

// Order informations
type Order struct {
	ProductCode string
	// 約定判定用ID
	OrderID string

	Side  int
	Price float64
	// Qty 買建玉は正、売建玉は負
	Qty float64

	Commission float64
	SFD        float64

	// 約定の有無
	Status StatusType
}

func toInt(side string) int {
	if side == types.BUY {
		return 1
	} else if side == types.SELL {
		return -1
	}
	return 0
}

func (p *Managed) Switch(childs []jsonrpc.ChildOrderEvent) {
	for i := range childs {
		// ORDER, ORDER_FAILED, CANCEL, CANCEL_FAILED, EXECUTION, EXPIRE
		switch childs[i].EventType {
		case "ORDER":
			p.Orders.Set(Order{
				ProductCode: childs[i].ProductCode,
				OrderID:     childs[i].ChildOrderAcceptanceID,
				Side:        toInt(childs[i].Side),
				Price:       float64(childs[i].Price),
				Qty:         childs[i].Size,
				Status:      OnBoard,
			})

		case "ORDER_FAILED":
			p.Orders.Delete(childs[i].ChildOrderAcceptanceID)
			p.Cancels.Delete(childs[i].ChildOrderAcceptanceID)

		case "CANCEL_FAILED":
			p.Cancels.Delete(childs[i].ChildOrderAcceptanceID)

		case "EXECUTION":
			p.executed(childs[i])

		case "CANCEL":
			p.cancel(childs[i].ChildOrderAcceptanceID)

		case "EXPIRE":
			p.cancel(childs[i].ChildOrderAcceptanceID)
		}
	}
}

func (p *Managed) executed(e jsonrpc.ChildOrderEvent) StatusType {
	o, ok := p.Orders.IsThere(e.ChildOrderAcceptanceID)
	if !ok {
		return NotExist
	}

	if e.Side == types.BUY {
		// Qtyは正, e.Sizeは正
		o.Qty -= e.Size
		if 0 < o.Qty { // 買建玉が残る部分約定
			return p.partial(o, e.Size)
		}

		o.Qty = e.Size
		return p.complete(o)

	} else if e.Side == types.SELL {
		// Qtyは負, e.Sizeは正
		o.Qty += e.Size
		if o.Qty < 0 { // 売建玉が残る部分約定
			return p.partial(o, e.Size)
		}

		o.Qty = e.Size
		return p.complete(o)

	}

	return NotExist
}

func (p *Orders) Set(o Order) {
	o.Qty = math.Abs(o.Qty) * float64(o.Side)
	p.m.Store(o.OrderID, o)
}

func (p *Orders) Reset() {
	p.m = sync.Map{}
}

func (p *Orders) Delete(uuid interface{}) {
	p.m.Delete(uuid)
}

// Deletes 保有注文/建玉/キャンセルを一部削除する
// 削除ルールはorderIDをストリングソートし、古い方から引数分だけ
func (p *Orders) Deletes(parcent int) (deleteCount int) {
	var keys []string
	p.m.Range(func(k, v interface{}) bool {
		keys = append(keys, fmt.Sprintf("%v", k))
		return true
	})

	// 古いものが先頭
	sort.Strings(keys)
	stop := float64(len(keys)) * float64(parcent) / float64(100)
	for range keys {
		p.m.Delete(keys[deleteCount])

		deleteCount++
		if stop < float64(deleteCount) {
			break
		}
	}
	return deleteCount
}

func (p *Orders) IsThere(uuid interface{}) (o Order, isThere bool) {
	v, ok := p.m.Load(uuid)
	if !ok {
		return o, false
	}
	return assert(v)
}

func assert(in interface{}) (o Order, ok bool) {
	o, ok = in.(Order)
	if !ok {
		return o, false
	}
	return o, true
}

const SATOSHI float64 = 0.00000001

func (p *Orders) Sum() (length int, avg, sum float64) {
	var (
		prices, volumes []float64
	)
	p.m.Range(func(key, value interface{}) bool {
		o, ok := p.IsThere(key)
		if !ok {
			return false
		}

		length++
		sum += o.Qty

		size := math.Abs(o.Qty)
		if SATOSHI < size {
			prices = append(prices, o.Price)
			volumes = append(volumes, size)
		}

		return true
	})

	return length, stat.Mean(prices, volumes), sum
}

// Check 約定情報を引数に、mapに保有したordersから約定/部分約定を確認
// 確認後positionsへ移動する
func (p *Managed) Check(isCancel bool, uuid interface{}, side string, qty float64) (status StatusType) {
	if isCancel {
		return p.cancel(uuid)
	}

	o, ok := p.Orders.IsThere(uuid)
	if !ok {
		return NotExist
	}

	if side == types.BUY {
		// Qtyは正, qtyは正
		o.Qty -= qty
		if 0 < o.Qty { // 買建玉が残る部分約定
			return p.partial(o, qty)
		}

		o.Qty = qty
		return p.complete(o)

	} else if side == types.SELL {
		// Qtyは負, qtyは正
		o.Qty += qty
		if o.Qty < 0 { // 売建玉が残る部分約定
			return p.partial(o, qty)
		}

		o.Qty = qty
		return p.complete(o)

	}

	// sideが合わないなど稀有な例
	return NotExist
}

func (p *Managed) partial(o Order, qty float64) StatusType {
	if o.Status == Partial { // 部分約定ならば前約定と合算
		pos, ok := p.Positions.IsThere(o.OrderID)
		if !ok {
			return NotExist
		}
		// 残注文
		p.Orders.m.Store(o.OrderID, o)
		// 約定 -> 建玉
		o.Qty = (math.Abs(pos.Qty) + math.Abs(qty)) * float64(o.Side)
	} else {
		// 残注文
		o.Status = Partial
		p.Orders.m.Store(o.OrderID, o)

		// 約定 -> 建玉
		o.Qty = math.Abs(qty) * float64(o.Side)
	}

	o.Status = Partial
	p.Positions.m.Store(o.OrderID, o)
	return Partial
}

func (p *Managed) complete(o Order) StatusType {
	p.Orders.m.Delete(o.OrderID)

	if o.Status == Partial { // 部分約定ならば前約定と合算
		pos, ok := p.Positions.IsThere(o.OrderID)
		if !ok {
			return NotExist
		}
		o.Qty = (math.Abs(pos.Qty) + math.Abs(o.Qty)) * float64(o.Side)
	} else {
		o.Qty = math.Abs(o.Qty) * float64(o.Side)
	}

	o.Status = Completed
	p.Positions.m.Store(o.OrderID, o)
	return Completed
}

func (p *Managed) cancel(uuid interface{}) StatusType {
	p.Orders.m.Delete(uuid)
	p.Cancels.m.Store(uuid, Order{})
	return Canceled
}

const float64EqualityThreshold = 1e-9

// isFloatEqual
func isFloatEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
