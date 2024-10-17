package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var (
	count int
)

var AddressGenCmd = &cobra.Command{
	Use:   "generate-addresses",
	Short: "Generate Ethereum addresses",
	Run:   runAddressGenerator,
}

func init() {
	AddressGenCmd.Flags().IntVarP(&count, "n", "n", 1000000, "number of addresses to generate")
	AddressGenCmd.Flags().StringVarP(&outputFile, "output", "o", "addresses.txt", "output file for the addresses")
}

func runAddressGenerator(_ *cobra.Command, _ []string) {
	// Open the output file
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	for i := 0; i < count; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}
		address := crypto.PubkeyToAddress(key.PublicKey).Hex() // Get the hex representation of the address

		// Write the address to the file
		if _, err := file.WriteString(address + "\n"); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}
	fmt.Printf("Generated %d Ethereum addresses\n", count)
}
