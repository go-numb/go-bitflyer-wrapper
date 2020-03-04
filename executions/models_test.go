package executions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-numb/go-bitflyer/v1/types"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
	"golang.org/x/sync/errgroup"
)

func TestSetExecution(t *testing.T) {
	ch := make(chan jsonrpc.WsWriter)

	ctx, cancel := context.WithCancel(context.Background())
	go jsonrpc.Connect(ctx, ch, []string{"lightning_executions"}, []string{string(types.FXBTCJPY)})
	defer cancel()

	e := New()

	var eg errgroup.Group

	eg.Go(func() error {
		for {
			select {
			case v := <-ch:
				switch v.Types {
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
				e.received(l)

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
