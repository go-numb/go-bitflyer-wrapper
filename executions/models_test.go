package executions

import (
	"fmt"
	"testing"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
)

func TestSetExecution(t *testing.T) {
	ch := make(chan jsonrpc.Response)
	go jsonrpc.Get([]string{"lightning_executions_FX_BTC_JPY"}, ch)

	e := New()

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
}
