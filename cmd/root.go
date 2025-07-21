package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "nacos-cli",
	Short: "Nacos命令行工具",
	Long:  `一个用于管理Nacos配置、用户和工作空间的命令行工具`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: $HOME/.nacos-cli.yaml)")
	rootCmd.PersistentFlags().String("server", "", "Nacos服务器地址")
	rootCmd.PersistentFlags().String("username", "", "用户名")
	rootCmd.PersistentFlags().String("password", "", "密码")
	rootCmd.PersistentFlags().String("namespace", "", "命名空间")

	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nacos-cli")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}
