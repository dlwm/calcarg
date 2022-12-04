# calcarg

# how to use

```golang
package main

import (
	"fmt"
	"github.com/dlwm/calcarg"
)

func main() {
	calculator, _ := calcarg.Analyse("(100-<age>)*<health>/100")
	res, _ := calculator.Eval("{\"age\":21,\"health\":60}")
	fmt.Println(res)
}
```