package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "用户管理",
	Long:  `管理Nacos用户和密码`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录验证",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		// 使用下划线忽略不需要的返回值
		_, err := client.Login()
		if err != nil {
			return fmt.Errorf("登录失败: %w", err)
		}

		// 保存token到配置文件
		viper.Set("token", client.Token)
		viper.Set("tokenExpiry", client.TokenExpiry)

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configFile = home + "/.nacos-cli.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("保存token失败: %w", err)
		}

		// 计算token过期时间
		expiryTime := time.Unix(client.TokenExpiry, 0)
		expiryFormatted := expiryTime.Format("2006-01-02 15:04:05")

		fmt.Println("登录成功!")
		fmt.Printf("Token将在 %s 过期\n", expiryFormatted)
		return nil
	},
}

var setUserCmd = &cobra.Command{
	Use:   "set [username] [password]",
	Short: "设置用户凭据",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("username", args[0])
		viper.Set("password", args[1])

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configFile = home + "/.nacos-cli.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("用户凭据已保存到 %s\n", configFile)
		return nil
	},
}

var showUserCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		username := viper.GetString("username")
		password := viper.GetString("password")
		server := viper.GetString("server")
		namespace := viper.GetString("namespace")

		// 创建密码掩码
		maskedPassword := ""
		if len(password) > 0 {
			maskedPassword = password[:1] + "****"
		}

		fmt.Printf("服务器: %s\n", server)
		fmt.Printf("用户名: %s\n", username)
		fmt.Printf("密码: %s\n", maskedPassword)
		fmt.Printf("命名空间: %s\n", namespace)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 清除保存的token
		viper.Set("token", "")
		viper.Set("tokenExpiry", 0)

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configFile = home + "/.nacos-cli.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("清除token失败: %w", err)
		}

		fmt.Println("已退出登录")
		return nil
	},
}

var setServerCmd = &cobra.Command{
	Use:   "server [server-url]",
	Short: "设置服务器地址",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("server", args[0])

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configFile = home + "/.nacos-cli.yaml"
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("服务器地址已设置为: %s\n", args[0])
		fmt.Printf("配置已保存到: %s\n", configFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	userCmd.AddCommand(loginCmd)
	userCmd.AddCommand(logoutCmd)
	userCmd.AddCommand(setUserCmd)
	userCmd.AddCommand(showUserCmd)
	userCmd.AddCommand(setServerCmd)
}
