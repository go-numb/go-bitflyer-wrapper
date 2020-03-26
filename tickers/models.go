package tickers

import (
	"math"
	"sync"
	"time"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/public/ticker"
)

type SFDLevel int

func (p SFDLevel) String() string {
	switch p {
	case SFDLV4:
		return "約定金額の 2%"
	case SFDLV3:
		return "約定金額の 1%"
	case SFDLV2:
		return "約定金額の 0.5%"
	case SFDLV1:
		return "約定金額の 0.25%"
	}
	return "通常"
}

const (
	SFDTHRESHOLD = 0.05
	LEV0_05      = 0.05
	LEV0_1       = 0.1
	LEV0_15      = 0.15
	LEV0_2       = 0.2
)

const (
	SFDLV0 SFDLevel = iota
	SFDLV1
	SFDLV2
	SFDLV3
	SFDLV4
)

type Ticker struct {
	sync.RWMutex

	// Spot
	spot                      float64
	sask, sbid                float64
	sasksize, sbidsize        float64
	saskdepth, sbiddepth      float64
	svolume, svolumeByProduct float64

	// FX
	ltp                     float64
	ask, bid                float64
	asksize, bidsize        float64
	askdepth, biddepth      float64
	volume, volumeByProduct float64

	sfdDeviation float64
	isSFD        bool

	sdelay time.Duration
	delay  time.Duration
}

func New() *Ticker {
	return &Ticker{}
}

func (p *Ticker) Set(isSpot bool, t *ticker.Response) {
	p.Lock()
	defer p.Unlock()

	if !isSpot {
		p.ltp = t.LTP
		p.ask = t.BestAsk
		p.asksize = t.BestAskSize
		p.askdepth = t.TotalAskDepth
		p.bid = t.BestBid
		p.bidsize = t.BestBidSize
		p.biddepth = t.TotalBidDepth
		p.volume = t.Volume
		p.volumeByProduct = t.VolumeByProduct
		p.delay = time.Now().Sub(t.Timestamp.Time)

		return
	}

	p.spot = t.LTP
	p.sask = t.BestAsk
	p.sasksize = t.BestAskSize
	p.saskdepth = t.TotalAskDepth
	p.sbid = t.BestBid
	p.sbidsize = t.BestBidSize
	p.sbiddepth = t.TotalBidDepth
	p.svolume = t.Volume
	p.svolumeByProduct = t.VolumeByProduct
	p.sdelay = time.Now().Sub(t.Timestamp.Time)

}

func (p *Ticker) IsSFD() (SFDLevel, float64, bool) {
	p.RLock()
	defer p.RUnlock()

	ratio := (p.ltp / p.spot) - 1
	dev := math.Abs(ratio)
	switch {
	case LEV0_2 <= dev:
		return SFDLV4, ratio, true
	case LEV0_15 <= dev:
		return SFDLV3, ratio, true
	case LEV0_1 <= dev:
		return SFDLV2, ratio, true
	case LEV0_05 <= dev:
		return SFDLV1, ratio, true
	}

	return SFDLV0, ratio, false
}

func (p *Ticker) Deviation() float64 {
	p.RLock()
	defer p.RUnlock()
	return (p.ltp / p.spot) - 1
}

func (p *Ticker) LTP(isSpot bool) float64 {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.ltp
	}
	return p.spot
}

func (p *Ticker) Best(isSpot bool) (ask, bid float64) {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.ask, p.bid
	}
	return p.sask, p.sbid
}

func (p *Ticker) BestSize(isSpot bool) (ask, bid float64) {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.asksize, p.bidsize
	}
	return p.sasksize, p.sbidsize
}

func (p *Ticker) BestDepth(isSpot bool) (ask, bid float64) {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.askdepth, p.biddepth
	}
	return p.saskdepth, p.sbiddepth
}

func (p *Ticker) Volume(isSpot bool) float64 {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.volume
	}
	return p.svolume
}

func (p *Ticker) Delay(isSpot bool) time.Duration {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.delay
	}
	return p.sdelay
}

func (p *Ticker) VolumeByProduct(isSpot bool) float64 {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		return p.volumeByProduct
	}
	return p.svolumeByProduct
}

// VolumeAvgInTerm return volume par binSize, [1s,1m,5m,1h].
func (p *Ticker) VolumeAvgInTerm(isSpot bool, binSize string) float64 {
	p.RLock()
	defer p.RUnlock()

	if !isSpot {
		switch binSize {
		case "1s":
			return math.Max(0, p.volumeByProduct/86400)
		case "1m":
			return math.Max(0, p.volumeByProduct/1440)
		case "5m":
			return math.Max(0, p.volumeByProduct/288)
		case "1h":
			return math.Max(0, p.volumeByProduct/24)
		}
		return p.volumeByProduct
	}

	switch binSize {
	case "1s":
		return math.Max(0, p.svolumeByProduct/86400)
	case "1m":
		return math.Max(0, p.svolumeByProduct/1440)
	case "5m":
		return math.Max(0, p.svolumeByProduct/288)
	case "1h":
		return math.Max(0, p.svolumeByProduct/24)
	}
	return p.svolumeByProduct
}
