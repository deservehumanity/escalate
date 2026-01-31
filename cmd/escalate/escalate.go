package main

import (
	"escalate/internal/client"
	"escalate/internal/wallet"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var verboseFlag bool = false
var qrFlag bool = false

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

		c, _ := client.NewHttpClient()
		w, err := wallet.New(&c)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(w.Mnemo)

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

		c, _ := client.NewHttpClient()
		w, err := wallet.NewFromData(data, &c)
		if err != nil {
			panic(err)
		}

		address, err := w.DeriveAddress()
		if err != nil {
			panic(err)
		}

		fmt.Println(address.Addr)
		if qrFlag {
			qr, err := qrcode.New(address.Addr, qrcode.Medium)
			if err != nil {
				panic(err)
			}

			fmt.Println(qr.ToString(false))
		}

		data, err = w.Serialize()
		if err != nil {
			panic(err)
		}

		os.WriteFile(walletFilePath, data, 0644)
	},
}

var lsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all used addresses",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		w, err := loadWallet()
		if err != nil {
			panic(err)
		}

		if verboseFlag {
			for _, addr := range w.Addresses {
				outString := fmt.Sprintf("%s [+%d sat | -%d sat]",
					addr.Addr,
					addr.Chain.FundedTxoSum,
					addr.Chain.SpentTxoSum,
				)

				fmt.Println(outString)
			}
		} else {
			for _, addr := range w.Addresses {
				fmt.Println(addr.Addr)
			}
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
			fmt.Println(w.Addresses[len(w.Addresses)-1].Addr)
		}

		if qrFlag && len(w.Addresses) > 0 {
			addr := w.Addresses[len(w.Addresses)-1].Addr
			qr, err := qrcode.New(addr, qrcode.Medium)
			if err != nil {
				panic(err)
			}

			fmt.Println(qr.ToString(false))
			// GENERATE A address qr code here
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

		data, err := w.Serialize()
		if err != nil {
			panic(err)
		}

		_, walletFilePath, err := walletPaths()
		if err != nil {
			panic(err)
		}

		os.WriteFile(walletFilePath, data, 0644)

		fmt.Println(balance)
	},
}

var pollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Polls last address for transactions",
	Run: func(cmd *cobra.Command, args []string) {
		w, err := loadWallet()
		if err != nil {
			panic(err)
		}

		var currentStats *client.AddressStats

		for {
			s, err := w.GetClient().GetAddressStats(w.Addresses[len(w.Addresses)-1].Addr)
			if err != nil {
				fmt.Println(err)
			}

			if currentStats != nil &&
				s.ChainStats.TxCount > currentStats.ChainStats.TxCount {
				fmt.Println("New transaction spotted")
				fmt.Println("Before: ", currentStats.ChainStats)
				fmt.Println("After: ", s.ChainStats)
			}

			currentStats = &s

			time.Sleep(time.Minute)
		}
	},
}

func main() {
	lsCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "verbose output")
	receiveCmd.Flags().BoolVar(&qrFlag, "qr", false, "generate a qr code")
	caCmd.Flags().BoolVar(&qrFlag, "qr", false, "generate a qr code")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(receiveCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(caCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(pollCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func loadWallet() (*wallet.Wallet, error) {
	_, walletFilePath, err := walletPaths()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(walletFilePath)
	if err != nil {
		return nil, err

	}
	c, _ := client.NewHttpClient()
	w, err := wallet.NewFromData(data, &c)
	if err != nil {
		return nil, err
	}

	return &w, nil
}
