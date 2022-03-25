package main

import (
	"fmt"

	"os"
	"os/signal"

	"github.com/realPy/inputeventsubsystem"
)

func main() {
	var choiceindex int
	inputs := inputeventsubsystem.ScanInputs("/dev/input")

	fmt.Println("Available devices:")
	for index, device := range inputs {

		if dev, err := inputeventsubsystem.Open(device); err == nil {
			fmt.Printf("[%d] %s %s VendorID:%x ProductID:%x\n", index, device, dev.Name, dev.VendorID, dev.ProductID)
			defer dev.Close()
		}

	}
	fmt.Println("Select the device index number")

	fmt.Scanf("%d", &choiceindex)

	if choiceindex >= 0 && choiceindex < len(inputs) {
		if device, err := inputeventsubsystem.Open(inputs[choiceindex]); err == nil {
			defer device.Close()
			device.Grab(true)

			fmt.Printf("You have open %s\n", device)
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			chanevents := device.Read()
		loopevents:
			for {

				select {
				case events := <-chanevents:
					for _, ev := range events {
						fmt.Printf("%s\n", ev)
					}

					device.ReadDone(events)
				case err := <-device.Error():
					fmt.Printf("Error: %s\n", err.Error())
					return
				case <-c:
					fmt.Printf("Stopped\n")
					break loopevents

				}
			}

		} else {
			panic(err)
		}

	}

}
