package orders

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-numb/go-bitflyer/v1/types"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
)

func TestSet(t *testing.T) {
	var (
		channels = []string{
			"lightning_ticker_FX_BTC_FX",
			"child_order_events",
			// "parent_order_events",
		}
		symbols = []string{string(types.FXBTCJPY)}
		ch      = make(chan jsonrpc.WsWriter)
	)
	key := os.Getenv("BFKEY")
	secret := os.Getenv("BFSECRET")

	ctx, cancel := context.WithCancel(context.Background())
	go jsonrpc.Connect(ctx, ch, []string{"lightning_executions"}, symbols, nil)
	defer cancel()

	go jsonrpc.ConnectForPrivate(ctx, ch, key, secret, channels, nil)

	m := New()

	for {
		select {
		case v := <-ch:
			switch v.Types {

			// case jsonrpc.Executions:
			// 	func() {
			// 		start := time.Now()
			// 		defer func() {
			// 			end := time.Now()
			// 			fmt.Println("exec time: ", end.Sub(start))
			// 		}()

			// 		for i := range v.Executions {
			// 			m.Check(false, v.Executions[i].BuyChildOrderAcceptanceID, v.Executions[i].Side, v.Executions[i].Size)
			// 			m.Check(false, v.Executions[i].SellChildOrderAcceptanceID, v.Executions[i].Side, v.Executions[i].Size)
			// 		}
			// 	}()

			case jsonrpc.ChildOrders:
				m.Switch(v.ChildOrderEvent)
				_, size := m.Orders.Sum()
				fmt.Printf("onBoard: %f\n", size)
				_, size = m.Positions.Sum()
				fmt.Printf("hasSize: %f\n", size)
			}
		}
	}
}
