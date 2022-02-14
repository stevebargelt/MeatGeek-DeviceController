package main

import (
	"fmt"
	"log"
	"strings"

	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

func main() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
		// Make sure periph is initialized.
	state, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	// Prints the loaded driver.
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}
	// Enumerate all SPI ports available and the corresponding pins.
	fmt.Print("SPI ports available:\n")
	for _, ref := range spireg.All() {
		fmt.Printf("- %s\n", ref.Name)
		if ref.Number != -1 {
			fmt.Printf("  %d\n", ref.Number)
		}
		if len(ref.Aliases) != 0 {
			fmt.Printf("  %s\n", strings.Join(ref.Aliases, " "))
		}

		p, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v", err)
		}
		if p, ok := p.(spi.Pins); ok {
			fmt.Printf("  CLK : %s\n", p.CLK())
			fmt.Printf("  MOSI: %s\n", p.MOSI())
			fmt.Printf("  MISO: %s\n", p.MISO())
			fmt.Printf("  CS  : %s\n", p.CS())
		}
		if err := p.Close(); err != nil {
			fmt.Printf("  Failed to close: %v", err)
		}
	}
}