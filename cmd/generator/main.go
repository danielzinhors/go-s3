package main

import (
	"fmt"
	"os"
)

func main() {
	//Gerar aqruivos aleatórios
	for i := 0; i < 10; i++ {
		f, err := os.Create(fmt.Sprintf("../../tmp/file%d.txt", i))
		if err != nil {
			panic(err)
		}

		defer f.Close()
		f.WriteString("Hello word")
	}
}
