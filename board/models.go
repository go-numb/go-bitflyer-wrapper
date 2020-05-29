package board

import (
	"fmt"
	"sync"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/public/board"
	"gonum.org/v1/gonum/floats"
)

type Board struct {
	mux      sync.RWMutex
	mid      float64
	ask, bid *Book
}

func New() *Board {
	return &Board{
		ask: new(Book),
		bid: new(Book),
	}
}

func (p *Board) Set(b *board.Response) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.mid = b.MidPrice
	p.ask.set(b.Asks)
	p.bid.set(b.Bids)
}

func (p *Board) Reset() {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.ask.in = 0
	p.ask.out = 0
	p.bid.in = 0
	p.bid.out = 0
}

func (p *Board) Best() (ask, bid float64) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return p.ask.Best(true, p.mid), p.bid.Best(false, p.mid)
}

func (p *Board) InOut() (askin, askout, bidin, bidout int) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return p.ask.in, p.ask.out, p.bid.in, p.bid.out
}

func (p *Board) InOutRatio() (ask, bid float64) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return float64(p.ask.in) / float64(p.ask.out), float64(p.bid.in) / float64(p.bid.out)
}

func (p *Board) Mid() (mid float64) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return p.mid
}

type Book struct {
	in, out int

	prices Smap
}

type Smap struct {
	sync.Map
}

func (p *Book) set(b []board.Book) {
	for i := range b {
		p.insert(b[i].Price, b[i].Size)
		if floats.EqualWithinAbs(b[i].Size, 0, 1e-8) {
			p.out++
		}
	}
}

func (p *Book) insert(price, size float64) {
	v, isThere := p.prices.LoadOrStore(fmt.Sprintf("%.f", price), size)
	if !isThere {
		p.in++
	}
	f, ok := v.(float64)
	if !ok {
		return
	}
	if !floats.EqualWithinAbs(f, 0, 1e-8) {
		p.in++
	}
}

func (p *Book) Best(isAsk bool, mid float64) (price float64) {
	var count = int(mid * 0.001)

	if isAsk {
		for i := 0; i < count; i++ {
			v, ok := p.prices.Load(fmt.Sprintf("%d", int(mid)+i))
			if !ok {
				continue
			}
			size, ok := v.(float64)
			if !ok {
				continue
			}
			if floats.EqualWithinAbs(size, 0, 1e-8) {
				continue
			}

			price = float64(int(mid) + i)
			break
		}
	} else {
		for i := 0; i < count; i++ {
			v, ok := p.prices.Load(fmt.Sprintf("%d", int(mid)-i))
			if !ok {
				continue
			}
			size, ok := v.(float64)
			if !ok {
				continue
			}
			if floats.EqualWithinAbs(size, 0, 1e-8) {
				continue
			}

			price = float64(int(mid) - i)
			break
		}
	}

	return price
}
