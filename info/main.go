package main

import (
	"fmt"
	"info/utils"
)

func main() {
	_, err := utils.SimpleAnalyze()
	if err != nil {
		fmt.Printf("Ошибка анализа: %v\n", err)
		return
	}
}
