package tickers_test

import (
	"fmt"
	"testing"

	tt "github.com/go-numb/go-bitflyer-wrapper/tickers"
	v1 "github.com/go-numb/go-exchanges/api/bitflyer/v1"
	"github.com/go-numb/go-exchanges/api/bitflyer/v1/public/ticker"
	"github.com/go-numb/go-exchanges/api/bitflyer/v1/types"
)

func TestSet(t *testing.T) {
	client := v1.New(nil)
	tick := tt.New()

	fx, err := client.Ticker(ticker.New(types.FXBTCJPY))
	if err != nil {
		t.Fatal(err)
	}
	tick.Set(false, fx)

	spot, err := client.Ticker(ticker.New(types.BTCJPY))
	if err != nil {
		t.Fatal(err)
	}
	tick.Set(true, spot)

	lv, dev, isSFD := tick.IsSFD()
	if isSFD {
		fmt.Printf("%.f	%.f\n", tick.LTP(true), tick.LTP(false))
		fmt.Printf("Let's SFD party!! Level: %s, %f\n", lv, dev)
		return
	}
	fmt.Printf("%.f	%.f\n", tick.LTP(true), tick.LTP(false))
	fmt.Printf("is done.. Level: %s, %f\n", lv, dev)
}

func TestDeviation(t *testing.T) {
	fx := []float64{104, 105, 106}
	spot := []float64{100, 100, 100}

	tick := tt.New()

	for i := range fx {
		tick.Set(true, &ticker.Response{LTP: spot[i]})
		tick.Set(false, &ticker.Response{LTP: fx[i]})
		fmt.Printf("%f\n", tick.Deviation())
	}
}

func TestVolumeByProductInTerm(t *testing.T) {
	tick := tt.New()
	client := v1.New(nil)
	fx, err := client.Ticker(ticker.New(types.FXBTCJPY))
	if err != nil {
		t.Fatal(err)
	}
	tick.Set(false, fx)

	str := []string{"1s", "1m", "5m", "1h"}
	for i := range str {
		fmt.Printf("%.f	%f/%s\n", tick.VolumeByProduct(false), tick.VolumeAvgInTerm(false, str[i]), str[i])
	}
}
