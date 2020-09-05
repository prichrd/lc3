package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/prichrd/lc3/pkg/lc3"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s <image-path>\n", os.Args[0])
		os.Exit(1)
	}
	imagePath := os.Args[1]

	machine := lc3.NewMachine()

	mem, err := lc3.ReadImageFile(imagePath)
	if err != nil {
		panic(err)
	}
	machine.LoadMemory(mem)

	in := make(chan rune)
	defer close(in)
	machine.SetStdin(in)

	out := make(chan rune)
	defer close(out)
	machine.SetStdout(out)

	var wg sync.WaitGroup
	sig := make(chan struct{})

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sig:
				return
			case event := <-keysEvents:
				if event.Err != nil {
					fmt.Printf("Error getting key: %v", event.Err)
				}
				if event.Key == keyboard.KeyCtrlC {
					close(sig)
					return
				}
				if event.Key == keyboard.KeyEnter {
					in <- rune(10)
					continue
				}
				if event.Rune != 0 {
					in <- event.Rune
					continue
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sig:
				return
			case c := <-out:
				fmt.Print(string(c))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		machine.Reset()
		machine.Start(sig, time.Nanosecond)
	}()

	wg.Wait()
}
