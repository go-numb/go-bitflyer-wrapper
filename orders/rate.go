package orders

import (
	"math"
)

type Rate struct {
	// 発注
	AskOrdersCount, BidOrdersCount   int
	AskOrdersVolume, BidOrdersVolume float64
	// 約定
	AskExecutesCount, BidExecutesCount   int
	AskExecutesVolume, BidExecutesVolume float64
	Rate                                 float64
}

func (p *Rate) Culc() {
	p.Rate = math.Max(0, (p.AskExecutesVolume+p.BidExecutesVolume)/(p.AskOrdersVolume+p.BidOrdersVolume))
}

type Data []Order

func (data Data) ExecutionRate() (rate *Rate) {
	for i := range data {
		if data[i].Side == 1 {
			rate.BidOrdersCount++
			rate.BidOrdersVolume += math.Abs(data[i].Qty)
			if data[i].Status == Partial || data[i].Status == Completed {
				rate.BidOrdersCount++
				rate.BidOrdersVolume += math.Abs(data[i].Qty)
			}
		} else if data[i].Side == -1 {
			rate.AskOrdersCount++
			rate.AskOrdersVolume += math.Abs(data[i].Qty)
			if data[i].Status == Partial || data[i].Status == Completed {
				rate.AskOrdersCount++
				rate.AskOrdersVolume += math.Abs(data[i].Qty)
			}

		}
	}

	rate.Culc()
	return rate
}
