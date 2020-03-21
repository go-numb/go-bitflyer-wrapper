package executions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/types"

	"github.com/go-numb/go-exchanges/api/bitflyer/v1/realtime/jsonrpc"
	"golang.org/x/sync/errgroup"
)

func TestSetExecution(t *testing.T) {
	ch := make(chan jsonrpc.Response)

	ctx, cancel := context.WithCancel(context.Background())
	go jsonrpc.Connect(ctx, ch, []string{"lightning_executions"}, []string{string(types.FXBTCJPY)}, nil)
	defer cancel()

	e := New()
	term := NewTerm()
	breakout := NewChannel(60)

	var eg errgroup.Group

	eg.Go(func() error {
		for {
			select {
			case v := <-ch:
				switch v.Types {
				case jsonrpc.Executions:
					e.HighPerformanceSet(v.Executions)
					// ask, bid := e.Best()
					// fmt.Printf("%.f	%.f	%v\n", ask, bid, e.Delay().Seconds())
					// sum, buy, sell := e.Volume()
					// fmt.Printf("%.3f	%.3f	%.3f\n", sum, buy, sell)
					// for i := range e.prices {
					// 	fmt.Printf("%.f	%.4f\n", e.prices[i], e.volumes[i])
					// }

					// 期間内集計
					prices, volumes := e.Copy()
					term.Set(e.LTP(), prices, volumes)

				}
			}

		}
	})

	eg.Go(func() error {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// var l Losscut
				// l.isLosscut = true
				// l.createdAt = time.Now()
				// e.received(l)

				if isBreak := breakout.Set(term.WeightPrice()); isBreak {
					fmt.Printf("Breakout: %t,	which side: %s	entry: %.f\n", isBreak, breakout.Signal(), e.LTP())
					fmt.Printf("Contrarian: %t,	which side: %s	entry: %.f\n", isBreak, IsBUY(breakout.Signal()*-1), e.LTP())
					_, high, center, low := breakout.Channels()
					fmt.Printf("出来高荷重価格: %.f,	%.f	%.f	%.f\n", term.WeightPrice(), high, center, low)
				}
				term.Reset()

			}
		}
		return fmt.Errorf("")
	})

	eg.Go(func() error {
		for {
			select {
			case v := <-e.Event:
				fmt.Printf("losscut: %+v\n", v)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
