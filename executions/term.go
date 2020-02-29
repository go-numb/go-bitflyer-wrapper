package executions

import (
	"math"
	"sync"

	"gonum.org/v1/gonum/stat"
)

type Term struct {
	sync.RWMutex

	isBuy       int
	first, last float64
	high, low   float64
	volume      float64

	prices, volumes []float64
}

func NewTerm() *Term {
	return &Term{
		prices:  make([]float64, 0),
		volumes: make([]float64, 0),
	}
}

func (p *Term) Set(ltp float64, prices, volumes []float64) {
	p.Lock()
	defer p.Unlock()

	// 現在を過去n秒の配列に追加
	l := len(p.prices)
	if l == 0 {
		p.first = ltp
	}
	p.prices = append(p.prices, prices...)    // 過去に現在を追加
	p.volumes = append(p.volumes, volumes...) // 過去に現在を追加

	if 0 < l { // lenghtを合わせる
		p.prices = p.prices[l:]
		p.volumes = p.volumes[l:]
	}

	if len(p.prices) != len(p.volumes) {
		p.prices = []float64{}
		p.volumes = []float64{}
		return
	}

	p.last = ltp
}

func (p *Term) Reset() {
	p.Lock()
	defer p.Unlock()
	p.highAndLow()
	p.prices = make([]float64, 0)
	p.volumes = make([]float64, 0)
}

func (p *Term) IsBuy() int {
	p.RLock()
	defer p.RUnlock()
	return p.isBuy
}

func (p *Term) Diff() float64 {
	p.RLock()
	defer p.RUnlock()
	return math.Max(0, p.high-p.low)
}

// Change is change price in the term
// defferent spread, first price to last price in term
func (p *Term) Change() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.first - p.last
}

func (p *Term) High() float64 {
	p.RLock()
	defer p.RUnlock()
	return math.Max(0, p.high)
}

func (p *Term) Low() float64 {
	p.RLock()
	defer p.RUnlock()
	return math.Max(0, p.low)
}

func (p *Term) Volume() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.volume
}

func (p *Term) WeightPrice() float64 {
	p.RLock()
	defer p.RUnlock()
	if len(p.prices) != len(p.volumes) {
		return (p.high + p.low) / 2
	}
	return stat.Mean(p.prices, p.volumes)
}

func (p *Term) highAndLow() {
	l := len(p.prices)
	if l < 1 {
		p.isBuy = 0
		p.low = 0
		p.high = 0
		return
	}

	var (
		wg     sync.WaitGroup
		nH, nL int
		high   float64
		low    = p.last
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range p.prices {
			if high < p.prices[i] {
				high = p.prices[i]
				nH = i
			} else if p.prices[i] < low {
				low = p.prices[i]
				nL = i
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range p.volumes {
			p.volume += p.volumes[i]
		}
	}()

	wg.Wait()

	if nH < nL {
		p.isBuy = -1
	} else if nL < nH {
		p.isBuy = 1
	}

	p.high = high
	p.low = low
}
