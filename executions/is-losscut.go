package executions

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/types"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/public/execution"
)

type Losscut struct {
	isLosscut bool
	side      int
	price     float64
	volume    float64
	createdAt time.Time
}

// IsDisadvantage 不利約定の集計
func (p *Losscut) IsDisadvantage(e execution.Execution) bool {
	if !strings.HasPrefix(e.BuyChildOrderAcceptanceID, "JRF") {
		p.side = 1
		p.price = e.Price
		p.isLosscut = true
		p.volume += e.Size
		p.createdAt = time.Now()
		return true
	} else if !strings.HasPrefix(e.SellChildOrderAcceptanceID, "JRF") {
		p.side = -1
		p.price = e.Price
		p.isLosscut = true
		p.volume += e.Size
		p.createdAt = time.Now()
		return true
	}

	return false
}

func (p *Losscut) IsThere() bool {
	return p.isLosscut
}

func (p *Losscut) Side() int {
	return p.side
}

func (p *Losscut) Price() float64 {
	return p.price
}

func (p *Losscut) Volume() float64 {
	return p.volume
}

func (p *Losscut) CreatedAt() time.Time {
	return p.createdAt
}

func (p Losscut) String() string {
	return fmt.Sprintf("%t,%s,%.1f,%f,%s", p.isLosscut, types.ToSide(p.side), p.price, p.volume, p.createdAt.Format("2006/01/02 15:04:05"))
}
