package executions

import (
	"fmt"
	"strings"
	"time"

	v1 "github.com/go-numb/go-bitflyer/v1"

	pex "github.com/go-numb/go-bitflyer/v1/public/executions"
)

type Losscut struct {
	isLosscut bool
	side      int
	volume    float64
	createdAt time.Time
}

func (loss Losscut) revieved(c chan Losscut) {
	c <- loss
}

// isDisadvantage 不利約定の集計
func (p *Losscut) isDisadvantage(e pex.Execution) bool {
	if !strings.HasPrefix(e.BuyChildOrderAcceptanceID, "JRF") {
		p.side = 1
		p.createdAt = time.Now()
		return true
	} else if !strings.HasPrefix(e.SellChildOrderAcceptanceID, "JRF") {
		p.side = -1
		p.createdAt = time.Now()
		return true
	}

	return false
}

func (p Losscut) String() string {
	return fmt.Sprintf("%t,%s,%f,%s", p.isLosscut, v1.ToSide(p.side), p.volume, p.createdAt.Format("2006/01/02 15:04:05"))
}
