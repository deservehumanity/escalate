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
		w, err := loadWallet()
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
		w, err := loadWallet()
		if err != nil {
			panic(err)
		}

		if len(w.Addresses) > 0 {
			fmt.Println(w.Addresses[len(w.Addresses)-1])
		}
	},
}

var balanceCmd = &cobra.Command{
	Use:     "balance",
	Short:   "Get wallet balance",
	Aliases: []string{"bc", "b"},
	Run: func(cmd *cobra.Command, args []string) {
		w, err := loadWallet()
		if err != nil {
			panic(err)
		}

		balance, err := w.GetBalance()
		if err != nil {
			panic(err)
		}

		fmt.Println(balance)
	},
}

var forgetCmd = &cobra.Command{
	Use:     "forget",
	Short:   "Erase data about previous transactions. Creating new address will advance as if no addresses have beed forgotten",
	Aliases: []string{"fg", "f"},
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

		w.ForgetAllButFirst()

		data, err = w.Serialize()
		if err != nil {
			panic(err)
		}

		os.WriteFile(walletFilePath, data, 0644)
	},
}

func main() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(receiveCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(caCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(forgetCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func loadWallet() (*Wallet, error) {
	_, walletFilePath, err := walletPaths()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(walletFilePath)
	if err != nil {
		return nil, err

	}

	w, err := Deserialize(data)
	if err != nil {
		return nil, err
	}

	return w, nil
}
