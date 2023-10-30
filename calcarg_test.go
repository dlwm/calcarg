package calcarg

import (
	"fmt"
	"testing"
	"time"
)

func TestCalc(t *testing.T) {
	start := time.Now().UnixMilli()
	calculator, err := Analyse("(99-<age>)*<health>/88.88+2^(1/2)^2")
	if err != nil {
		panic(err)
	}
	res, err := calculator.Eval(map[string]float32{
		"age":    21,
		"health": 60,
	})
	_, _ = res, err
	fmt.Println(res, err)
	end := time.Now().UnixMilli()
	fmt.Println(end - start)
}
