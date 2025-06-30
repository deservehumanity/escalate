package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func walletPaths() (walletPath, walletFilePath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	walletPath = filepath.Join(home, ".escalate")
	walletFilePath = filepath.Join(walletPath, "wallet.json")
	return
}

var rootCmd = &cobra.Command{
	Use:   "escalate",
	Short: "A CLI Bitcoin Wallet",
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new bitcoin wallet",
	Run: func(cmd *cobra.Command, args []string) {
		walletPath, walletFilePath, err := walletPaths()
		if err != nil {
			panic(err)
		}

		if err := os.MkdirAll(walletPath, 0755); err != nil {
			panic(err)
		}

		w, err := New()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Seed phrase:")
		fmt.Println(w.Mnemonic)

		data, err := w.Serialize()
		if err != nil {
			fmt.Println(err)
			return
		}

		os.WriteFile(walletFilePath, data, 0644)
	},
}

var receiveCmd = &cobra.Command{
	Use:     "receive",
	Short:   "Generate new receive address",
	Aliases: []string{"r", "rec"},
	Run: func(cmd *cobra.Command, args []string) {
		_, walletFilePath, err := walletPaths()
		if err != nil {
			panic(err)
		}

		data, err := os.ReadFile(walletFilePath)
		if err != nil {
			panic(err)
		}

		wallet, err := Deserialize(data)
		if err != nil {
			panic(err)
		}

		addr, err := wallet.DeriveAddress()
		if err != nil {
			panic(err)
		}

		fmt.Println(addr)

		data, err = wallet.Serialize()
		if err != nil {
			panic(err)
		}

		os.WriteFile(walletFilePath, data, 0644)
	},
}

var lsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all used addresses",
	Aliases: []string{"ls", "l"},
	Run: func(cmd *cobra.Command, args []string) {
		_, walletFilePath, err := walletPaths()
		if err != nil {
			panic(err)
		}

		data, err := os.ReadFile(walletFilePath)
		if err != nil {
			panic(err)
		}

		w, err := Deserialize(data)
		if err != nil {
			panic(err)
		}

		for _, addr := range w.Addresses {
			fmt.Println(addr)
		}
	},
}

var caCmd = &cobra.Command{
	Use:     "current",
	Short:   "Get latest generated address",
	Aliases: []string{"ca", "c"},
	Run: func(cmd *cobra.Command, args []string) {
		_, walletFilePath, err := walletPaths()
		if err != nil {
			panic(err)
		}

		data, err := os.ReadFile(walletFilePath)
		if err != nil {
			panic(err)
		}

		w, err := Deserialize(data)
		if err != nil {
			panic(err)
		}

		if len(w.Addresses) > 0 {
			fmt.Println(w.Addresses[len(w.Addresses)-1])
		}
	},
}

func main() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(receiveCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(caCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
