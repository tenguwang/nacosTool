package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "命名空间管理",
	Long:  `管理Nacos命名空间的创建、删除和查看操作`,
}

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

		if len(namespaces) == 0 {
			fmt.Println("没有找到命名空间")
			return nil
		}

		fmt.Printf("%-40s %-20s %-50s\n", "命名空间ID", "命名空间名称", "描述")
		fmt.Println(strings.Repeat("-", 110))
		for _, ns := range namespaces {
			fmt.Printf("%-40s %-20s %-50s\n", ns.Namespace, ns.NamespaceShowName, ns.NamespaceDesc)
		}
		return nil
	},
}

var createNamespaceCmd = &cobra.Command{
	Use:   "create [namespace-id] [namespace-name]",
	Short: "创建命名空间",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return fmt.Errorf("登录失败: %w", err)
		}

		namespaceId := args[0]
		namespaceName := args[1]
		description, _ := cmd.Flags().GetString("desc")

		if err := client.CreateNamespace(namespaceId, namespaceName, description); err != nil {
			return fmt.Errorf("创建命名空间失败: %w", err)
		}

		fmt.Printf("成功创建命名空间: %s (%s)\n", namespaceName, namespaceId)
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

		namespaceId := args[0]

		if err := client.DeleteNamespace(namespaceId); err != nil {
			return fmt.Errorf("删除命名空间失败: %w", err)
		}

		fmt.Printf("成功删除命名空间: %s\n", namespaceId)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)

	namespaceCmd.AddCommand(listNamespaceCmd)
	namespaceCmd.AddCommand(createNamespaceCmd)
	namespaceCmd.AddCommand(deleteNamespaceCmd)

	createNamespaceCmd.Flags().StringP("desc", "d", "", "命名空间描述")
}
