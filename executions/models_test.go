package executions

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
	"golang.org/x/sync/errgroup"
)

func TestSetExecution(t *testing.T) {
	ch := make(chan jsonrpc.Response)
	go jsonrpc.Get([]string{"lightning_executions_FX_BTC_JPY"}, ch)

	loss := make(chan Losscut)
	e := New(loss)

	var eg errgroup.Group

	eg.Go(func() error {
		for {
			select {
			case v := <-ch:
				switch v.Type {
				case jsonrpc.Executions:
					e.HighPerformanceSet(v.Executions)
					ask, bid := e.Best()
					fmt.Printf("%.f	%.f	%v\n", ask, bid, e.Delay().Seconds())
					sum, buy, sell := e.Volume()
					fmt.Printf("%.3f	%.3f	%.3f\n", sum, buy, sell)
					// for i := range e.prices {
					// 	fmt.Printf("%.f	%.4f\n", e.prices[i], e.volumes[i])
					// }

				}
			}

		}
	})

	eg.Go(func() error {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var l Losscut
				l.isLosscut = true
				l.createdAt = time.Now()
				l.received(e.l)

			}
		}
		return fmt.Errorf("")
	})

	eg.Go(func() error {
		for {
			select {
			case v := <-loss:
				fmt.Printf("losscut: %+v\n", v)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
