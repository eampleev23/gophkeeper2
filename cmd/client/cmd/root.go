package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper2",
	Short: "gophkeeper2 client",
	Long:  `Выберите действие: регистрация или выход.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Добро пожаловать, выберите действие:")
		fmt.Println("1. Зарегистрироваться")
		fmt.Println("2. Выйти")

		var choice int
		fmt.Print("> ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			registerUser()
		case 2:
			fmt.Println("Выход...")
			os.Exit(0)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
