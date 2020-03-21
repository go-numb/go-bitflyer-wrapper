package executions

import (
	"math"
	"sync"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/types"
)

type Channel struct {
	sync.RWMutex

	highline float64
	lowline  float64

	period int
	signal IsBUY

	prices []float64
}

func NewChannel(period int) *Channel {
	if period <= 1 {
		period = 20
	}

	return &Channel{
		period:  period,
		lowline: math.Inf(1),
		prices:  make([]float64, 0),
	}
}

func (p *Channel) Set(price float64) (isBreak bool) {
	p.Lock()
	defer p.Unlock()

	if len(p.prices) < p.period {
		p.prices = append(p.prices, price)
		return false
	}

	// 期間内価格から高値安値を取得する
	p.highline = 0
	p.lowline = math.Inf(1)
	for i := range p.prices {
		if p.highline < p.prices[i] {
			p.highline = p.prices[i]
		} else if p.prices[i] < p.lowline {
			p.lowline = p.prices[i]
		}
	}

	// 過去期間と比べ、現価格で更新可能か
	if p.highline < price {
		p.highline = price
		if p.signal != 1 {
			p.signal = 1
			isBreak = true
		}
	} else if price < p.lowline {
		p.lowline = price
		if p.signal != -1 {
			p.signal = -1
			isBreak = true
		}
	}

	p.prices = append(p.prices[len(p.prices)-p.period:], price)
	return isBreak
}

func (p *Channel) Signal() IsBUY {
	return p.signal
}

func (p *Channel) Channels() (period int, high, center, low float64) {
	p.RLock()
	defer p.RUnlock()

	return p.period, p.highline, (p.highline + p.lowline) / 2, p.lowline
}

type IsBUY int

func (p IsBUY) IsBuy() bool {
	if p < 0 {
		return false
	}
	return true
}

func (p IsBUY) String() string {
	if 0 < p {
		return types.BUY
	} else if p < 0 {
		return types.SELL
	}
	return types.UNDEFINED
}
