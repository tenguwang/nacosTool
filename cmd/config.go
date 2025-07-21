package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nacos-cli/pkg/nacos"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  `管理Nacos配置的增删改查操作`,
}

var getConfigCmd = &cobra.Command{
	Use:   "get [dataId] [group]",
	Short: "获取配置",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return err
		}

		config, err := client.GetConfig(args[0], args[1])
		if err != nil {
			return err
		}

		fmt.Printf("DataID: %s\n", config.DataID)
		fmt.Printf("Group: %s\n", config.Group)
		fmt.Printf("Content:\n%s\n", config.Content)
		return nil
	},
}

var setConfigCmd = &cobra.Command{
	Use:   "set [dataId] [group] [content]",
	Short: "创建或更新配置",
	Long: `创建或更新Nacos配置。可以直接提供内容或从文件读取内容。
对于简单的配置，可以直接使用命令行参数；
对于复杂的配置，建议使用 --file 参数从文件读取，或使用 import 命令导入已导出的配置文件。`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return err
		}

		var content string

		// 检查是否指定了文件
		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			// 从文件读取内容
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("读取文件失败: %w", err)
			}
			content = string(fileData)
		} else if len(args) >= 3 {
			// 从命令行参数获取内容
			content = args[2]
		} else {
			return fmt.Errorf("必须提供内容参数或使用--file指定文件")
		}

		config := &nacos.Config{
			DataID:  args[0],
			Group:   args[1],
			Content: content,
		}

		configType, _ := cmd.Flags().GetString("type")
		if configType != "" {
			config.Type = configType
		} else {
			// 尝试从文件扩展名推断类型
			if strings.HasSuffix(args[0], ".yml") || strings.HasSuffix(args[0], ".yaml") {
				config.Type = "yaml"
			} else if strings.HasSuffix(args[0], ".properties") {
				config.Type = "properties"
			} else if strings.HasSuffix(args[0], ".json") {
				config.Type = "json"
			} else if strings.HasSuffix(args[0], ".xml") {
				config.Type = "xml"
			}
		}

		if err := client.PublishConfig(config); err != nil {
			return err
		}

		fmt.Printf("配置 %s@%s 设置成功\n", args[0], args[1])
		return nil
	},
}

var deleteConfigCmd = &cobra.Command{
	Use:   "delete [dataId] [group]",
	Short: "删除配置",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return err
		}

		if err := client.DeleteConfig(args[0], args[1]); err != nil {
			return err
		}

		fmt.Printf("配置 %s@%s 删除成功\n", args[0], args[1])
		return nil
	},
}

var listConfigCmd = &cobra.Command{
	Use:   "list",
	Short: "列出配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return fmt.Errorf("登录失败: %w", err)
		}

		pageNo, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("size")

		configs, err := client.ListConfigs(pageNo, pageSize)
		if err != nil {
			return fmt.Errorf("获取配置列表失败: %w", err)
		}

		if len(configs) == 0 {
			fmt.Println("没有找到配置")
			return nil
		}

		fmt.Printf("%-40s %-20s %-10s\n", "DataID", "Group", "Type")
		fmt.Println(strings.Repeat("-", 70))
		for _, config := range configs {
			configType := config.Type
			if configType == "" {
				// 尝试从文件扩展名推断类型
				if strings.HasSuffix(config.DataID, ".yml") || strings.HasSuffix(config.DataID, ".yaml") {
					configType = "yaml"
				} else if strings.HasSuffix(config.DataID, ".properties") {
					configType = "properties"
				} else if strings.HasSuffix(config.DataID, ".json") {
					configType = "json"
				} else if strings.HasSuffix(config.DataID, ".xml") {
					configType = "xml"
				} else {
					configType = "text"
				}
			}
			fmt.Printf("%-40s %-20s %-10s\n", config.DataID, config.Group, configType)
		}
		return nil
	},
}

var exportConfigCmd = &cobra.Command{
	Use:   "export [output-dir]",
	Short: "导出配置",
	Long:  `导出配置到指定目录，可以指定dataId和group导出单个配置`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return err
		}

		outputDir := args[0]
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}

		// 检查是否指定了dataId和group
		dataId, _ := cmd.Flags().GetString("dataId")
		group, _ := cmd.Flags().GetString("group")

		// 如果指定了dataId和group，则只导出单个配置
		if dataId != "" && group != "" {
			fullConfig, err := client.GetConfig(dataId, group)
			if err != nil {
				return fmt.Errorf("获取配置 %s@%s 失败: %v", dataId, group, err)
			}

			if err := exportSingleConfig(outputDir, fullConfig); err != nil {
				return err
			}

			fmt.Printf("成功导出配置: %s@%s\n", group, dataId)
			return nil
		}

		// 否则导出所有配置
		configs, err := client.ListConfigs(1, 1000)
		if err != nil {
			return err
		}

		exportCount := 0
		for _, config := range configs {
			fullConfig, err := client.GetConfig(config.DataID, config.Group)
			if err != nil {
				fmt.Printf("获取配置 %s@%s 失败: %v\n", config.DataID, config.Group, err)
				continue
			}

			if err := exportSingleConfig(outputDir, fullConfig); err != nil {
				fmt.Printf("导出配置 %s@%s 失败: %v\n", config.Group, config.DataID, err)
				continue
			}

			exportCount++
		}

		fmt.Printf("配置导出完成，共导出 %d 个配置到 %s\n", exportCount, outputDir)
		return nil
	},
}

// 导出单个配置的辅助函数
func exportSingleConfig(outputDir string, config *nacos.Config) error {

	// 确定配置类型
	configType := config.Type
	if configType == "" {
		// 尝试从文件扩展名推断类型
		if strings.HasSuffix(config.DataID, ".yml") || strings.HasSuffix(config.DataID, ".yaml") {
			configType = "yaml"
		} else if strings.HasSuffix(config.DataID, ".properties") {
			configType = "properties"
		} else if strings.HasSuffix(config.DataID, ".json") {
			configType = "json"
		} else if strings.HasSuffix(config.DataID, ".xml") {
			configType = "xml"
		} else {
			configType = "text"
		}
	}

	// 创建内容文件（直接保存配置内容）
	// 使用 group@dataId 格式命名文件，保留原始扩展名
	filename := fmt.Sprintf("%s@%s", config.Group, config.DataID)
	filepath := filepath.Join(outputDir, filename)

	// 在文件开头添加注释，包含配置的元信息
	var header string
	switch configType {
	case "yaml", "yml":
		header = fmt.Sprintf("# Nacos配置\n# DataID: %s\n# Group: %s\n# Type: %s\n\n",
			config.DataID, config.Group, configType)
	case "properties":
		header = fmt.Sprintf("# Nacos配置\n# DataID: %s\n# Group: %s\n# Type: %s\n\n",
			config.DataID, config.Group, configType)
	case "json":
		// JSON不支持注释，所以不添加头部信息
		header = ""
	case "xml":
		header = fmt.Sprintf("<!-- Nacos配置\nDataID: %s\nGroup: %s\nType: %s\n-->\n\n",
			config.DataID, config.Group, configType)
	default:
		header = fmt.Sprintf("# Nacos配置\n# DataID: %s\n# Group: %s\n# Type: %s\n\n",
			config.DataID, config.Group, configType)
	}

	// 组合头部信息和内容
	content := header + config.Content

	if err := ioutil.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

var importConfigCmd = &cobra.Command{
	Use:   "import [input-dir]",
	Short: "导入配置",
	Long:  `从指定目录导入配置，可以指定文件导入单个配置`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := createClient()
		_, err := client.Login()
		if err != nil {
			return err
		}

		inputDir := args[0]

		// 检查是否指定了文件
		file, _ := cmd.Flags().GetString("file")
		if file != "" {
			// 导入单个文件
			filePath := filepath.Join(inputDir, file)
			if err := importSingleFile(client, filePath); err != nil {
				return fmt.Errorf("导入文件 %s 失败: %w", file, err)
			}
			fmt.Printf("成功导入配置: %s\n", file)
			return nil
		}

		// 否则导入目录中的所有文件
		files, err := ioutil.ReadDir(inputDir)
		if err != nil {
			return fmt.Errorf("读取目录失败: %w", err)
		}

		successCount := 0
		for _, fileInfo := range files {
			if fileInfo.IsDir() {
				continue
			}

			filePath := filepath.Join(inputDir, fileInfo.Name())
			if err := importSingleFile(client, filePath); err != nil {
				fmt.Printf("导入文件 %s 失败: %v\n", fileInfo.Name(), err)
				continue
			}

			successCount++
		}

		fmt.Printf("配置导入完成，成功导入 %d 个配置\n", successCount)
		return nil
	},
}

// 导入单个文件的辅助函数
func importSingleFile(client *nacos.Client, filePath string) error {
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 从文件名解析group和dataId
	// 文件名格式: group@dataId
	fileName := filepath.Base(filePath)
	parts := strings.SplitN(fileName, "@", 2)
	if len(parts) != 2 {
		return fmt.Errorf("文件名格式不正确，应为 'group@dataId': %s", fileName)
	}

	group := parts[0]
	dataId := parts[1]

	// 确定配置类型
	configType := ""
	// 从dataId的扩展名推断类型
	if strings.Contains(dataId, ".") {
		ext := filepath.Ext(dataId)
		switch strings.ToLower(ext) {
		case ".yaml", ".yml":
			configType = "yaml"
		case ".properties":
			configType = "properties"
		case ".json":
			configType = "json"
		case ".xml":
			configType = "xml"
		default:
			configType = "text"
		}
	}

	// 从文件内容中移除注释头部
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")

	// 跳过注释行
	startLine := 0
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "<!--") && trimmedLine != "" {
			startLine = i
			break
		}
	}

	// 重新组合内容，跳过注释头部
	cleanContent := strings.Join(lines[startLine:], "\n")

	// 创建配置对象
	config := &nacos.Config{
		DataID:  dataId,
		Group:   group,
		Content: cleanContent,
		Type:    configType,
	}

	// 发布配置
	if err := client.PublishConfig(config); err != nil {
		return fmt.Errorf("发布配置失败: %w", err)
	}

	return nil
}

func createClient() *nacos.Client {
	server := viper.GetString("server")
	username := viper.GetString("username")
	password := viper.GetString("password")
	namespace := viper.GetString("namespace")
	token := viper.GetString("token")
	tokenExpiry := viper.GetInt64("tokenExpiry")

	client := nacos.NewClient(server, username, password, namespace)

	// 如果有保存的token且未过期，则使用它
	if token != "" && tokenExpiry > time.Now().Unix() {
		client.Token = token
		client.TokenExpiry = tokenExpiry
	}

	return client
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(getConfigCmd)
	configCmd.AddCommand(setConfigCmd)
	configCmd.AddCommand(deleteConfigCmd)
	configCmd.AddCommand(listConfigCmd)
	configCmd.AddCommand(exportConfigCmd)
	configCmd.AddCommand(importConfigCmd)

	setConfigCmd.Flags().String("type", "", "配置类型 (yaml, properties, json等)")
	setConfigCmd.Flags().StringP("file", "f", "", "从文件读取配置内容")

	exportConfigCmd.Flags().StringP("dataId", "d", "", "指定要导出的配置ID")
	exportConfigCmd.Flags().StringP("group", "g", "", "指定要导出的配置分组")

	importConfigCmd.Flags().StringP("file", "f", "", "指定要导入的配置文件")

	listConfigCmd.Flags().Int("page", 1, "页码")
	listConfigCmd.Flags().Int("size", 20, "每页大小")
}
