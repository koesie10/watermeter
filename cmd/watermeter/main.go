package main

import (
	"fmt"
	"os"

	"github.com/warthog618/gpio"
)

func main() {
	defer func() {
		if pin != nil {
			pin.Unwatch()
		}
		gpio.Close()
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
