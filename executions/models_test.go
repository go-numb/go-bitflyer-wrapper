package executions

import (
	"fmt"
	"testing"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
	"golang.org/x/sync/errgroup"
)

func TestSetExecution(t *testing.T) {
	ch := make(chan jsonrpc.Response)
	go jsonrpc.Get([]string{"lightning_executions_FX_BTC_JPY"}, ch)

	e := New()

	var eg errgroup.Group

	eg.Go(func() error {
		for {
			select {
			case v := <-ch:
				switch v.Type {
				case jsonrpc.Executions:
					e.Set(v.Executions)
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
		for {
			select {
			case v := <-e.l:
				fmt.Printf("losscut: %+v\n", v)
			}
		}
	})

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
