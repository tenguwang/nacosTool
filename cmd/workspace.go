package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "工作空间管理",
	Long:  `管理Nacos命名空间/工作空间`,
}

var setWorkspaceCmd = &cobra.Command{
	Use:   "set [namespace]",
	Short: "设置当前工作空间",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("namespace", args[0])

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

		fmt.Printf("工作空间已设置为: %s\n", args[0])
		return nil
	},
}

var showWorkspaceCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前工作空间",
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := viper.GetString("namespace")
		if namespace == "" {
			fmt.Println("当前工作空间: public (默认)")
		} else {
			fmt.Printf("当前工作空间: %s\n", namespace)
		}
		return nil
	},
}

var clearWorkspaceCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除工作空间设置",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("namespace", "")

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

		fmt.Println("工作空间已清除，使用默认命名空间")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)

	workspaceCmd.AddCommand(setWorkspaceCmd)
	workspaceCmd.AddCommand(showWorkspaceCmd)
	workspaceCmd.AddCommand(clearWorkspaceCmd)

	// 添加命名空间管理命令
	var listNamespaceCmd = &cobra.Command{
		Use:   "list",
		Short: "列出所有命名空间",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := createClient()
			_, err := client.Login()
			if err != nil {
				return fmt.Errorf("登录失败: %w", err)
			}

			namespaces, err := client.ListNamespaces()
			if err != nil {
				return fmt.Errorf("获取命名空间列表失败: %w", err)
			}

			fmt.Printf("%-20s %-30s %-40s\n", "命名空间ID", "名称", "描述")
			fmt.Println(strings.Repeat("-", 90))
			for _, ns := range namespaces {
				fmt.Printf("%-20s %-30s %-40s\n", ns.Namespace, ns.NamespaceShowName, ns.NamespaceDesc)
			}
			return nil
		},
	}

	var createNamespaceCmd = &cobra.Command{
		Use:   "create [namespace-id] [name] [description]",
		Short: "创建命名空间",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := createClient()
			_, err := client.Login()
			if err != nil {
				return fmt.Errorf("登录失败: %w", err)
			}

			namespaceID := args[0]
			namespaceName := args[1]
			namespaceDesc := args[2]

			if err := client.CreateNamespace(namespaceID, namespaceName, namespaceDesc); err != nil {
				return fmt.Errorf("创建命名空间失败: %w", err)
			}

			fmt.Printf("命名空间 %s (%s) 创建成功\n", namespaceID, namespaceName)
			return nil
		},
	}

	var deleteNamespaceCmd = &cobra.Command{
		Use:   "delete [namespace-id]",
		Short: "删除命名空间",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := createClient()
			_, err := client.Login()
			if err != nil {
				return fmt.Errorf("登录失败: %w", err)
			}

			namespaceID := args[0]

			// 确认删除
			fmt.Printf("确定要删除命名空间 %s 吗？此操作不可恢复！(y/N): ", namespaceID)
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				fmt.Println("操作已取消")
				return nil
			}

			if err := client.DeleteNamespace(namespaceID); err != nil {
				return fmt.Errorf("删除命名空间失败: %w", err)
			}

			fmt.Printf("命名空间 %s 删除成功\n", namespaceID)
			return nil
		},
	}

	workspaceCmd.AddCommand(listNamespaceCmd)
	workspaceCmd.AddCommand(createNamespaceCmd)
	workspaceCmd.AddCommand(deleteNamespaceCmd)
}
