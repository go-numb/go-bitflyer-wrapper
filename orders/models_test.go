package orders

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-numb/go-bitflyer/v1/jsonrpc"
)

func TestSet(t *testing.T) {
	var (
		channels = []string{
			"lightning_ticker_BTC_JPY",
			"child_order_events",
			// "parent_order_events",
		}
		ch = make(chan jsonrpc.Response)
	)
	key := os.Getenv("BFKEY")
	secret := os.Getenv("BFSECRET")

	go jsonrpc.Get([]string{"lightning_executions_FX_BTC_JPY"}, ch)
	go jsonrpc.GetPrivate(key, secret, channels, ch)

	m := New()

	for {
		select {
		case v := <-ch:
			switch v.Type {

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
				m.Switch(v.ChildOrders)
				_, size := m.Orders.Sum()
				fmt.Printf("onBoard: %f\n", size)
				_, size = m.Positions.Sum()
				fmt.Printf("hasSize: %f\n", size)
			}
		}
	}
}
