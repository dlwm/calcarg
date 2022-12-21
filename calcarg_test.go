package calcarg

import (
	"fmt"
	"testing"
	"time"
)

func TestCalc(t *testing.T) {
	start := time.Now().UnixMilli()
	calculator, err := Analyse("(100-<age>)*<health>/100")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100000; i++ {
		res, err := calculator.Eval(map[string]float32{
			"age":    21,
			"health": 60,
		})
		_, _ = res, err
		//fmt.Println(res, err)
	}
	end := time.Now().UnixMilli()
	fmt.Println(end - start)
}
