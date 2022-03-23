package main

import (
	"fmt"
	"inputeventsubsystem"
	"os"
	"os/signal"
)

func main() {
	var choiceindex int
	inputs := inputeventsubsystem.ScanInputs("/dev/input")

	fmt.Println("Available devices:")
	for index, device := range inputs {

		if dev, err := inputeventsubsystem.Open(device); err == nil {
			fmt.Printf("[%d] %s %s\n", index, device, dev.Name)
			defer dev.Close()
		}

	}
	fmt.Println("Select the device index number")

	fmt.Scanf("%d", &choiceindex)

	if choiceindex >= 0 && choiceindex < len(inputs) {
		if device, err := inputeventsubsystem.Open(inputs[choiceindex]); err == nil {
			defer device.Close()

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
