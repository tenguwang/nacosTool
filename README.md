# Nacos CLI 工具

一个用于管理 Nacos 配置、用户和工作空间的命令行工具。

## 功能特性

- 配置管理：增删改查 Nacos 配置
- 配置导入导出：批量备份和恢复配置
- 用户管理：管理登录凭据
- 工作空间管理：切换不同的命名空间

## 安装

```bash
go build -o nacos-cli
```

## 配置

首先设置 Nacos 服务器和用户凭据：

```bash
# 设置服务器地址
./nacos-cli user server http://localhost:8848

# 设置用户凭据
./nacos-cli user set admin admin123

# 登录并保存token
./nacos-cli user login

# 或者使用命令行参数
./nacos-cli --server http://localhost:8848 --username admin --password admin123 config list
```

配置文件默认保存在 `~/.nacos-cli.yaml`

登录后，token 会被保存到配置文件中，后续命令无需重新登录，直到 token 过期。

## 使用方法

### 用户管理

```bash
# 设置服务器地址
./nacos-cli user server <server-url>

# 设置用户凭据
./nacos-cli user set <username> <password>

# 显示当前用户信息
./nacos-cli user show

# 测试登录
./nacos-cli user login

# 退出登录
./nacos-cli user logout
```

### 工作空间管理

```bash
# 设置工作空间（命名空间）
./nacos-cli workspace set <namespace-id>

# 显示当前工作空间
./nacos-cli workspace show

# 清除工作空间设置（使用默认）
./nacos-cli workspace clear

# 列出所有命名空间
./nacos-cli workspace list

# 创建命名空间
./nacos-cli workspace create <namespace-id> <name> <description>

# 删除命名空间
./nacos-cli workspace delete <namespace-id>
```

### 配置管理

```bash
# 获取配置
./nacos-cli config get <dataId> <group>

# 设置配置（直接提供内容）
./nacos-cli config set <dataId> <group> <content>

# 设置配置（从文件读取内容）
./nacos-cli config set <dataId> <group> --file <file-path>

# 设置配置并指定类型
./nacos-cli config set <dataId> <group> <content> --type yaml

# 删除配置
./nacos-cli config delete <dataId> <group>

# 列出配置
./nacos-cli config list

# 分页列出配置
./nacos-cli config list --page 1 --size 10
```

### 配置导入导出

```bash
# 导出所有配置到目录
./nacos-cli config export ./backup

# 导出指定配置
./nacos-cli config export ./backup --dataId application.yml --group DEFAULT_GROUP

# 从目录导入所有配置
./nacos-cli config import ./backup

# 导入指定配置文件
./nacos-cli config import ./backup --file DEFAULT_GROUP@application.yml
```

导出的配置文件使用 `<group>@<dataId>` 格式命名，保留原始文件扩展名。
文件开头会包含注释形式的元数据信息，方便查看和管理。

这种格式使得配置内容更易于阅读和编辑，特别是对于 YAML、JSON 等结构化配置。

## 示例

```bash
# 完整的使用流程
# 方式1：使用配置命令设置
./nacos-cli user server http://localhost:8848
./nacos-cli user set admin admin123
./nacos-cli workspace set dev
# 直接提供配置内容
./nacos-cli config set application.yml DEFAULT_GROUP "server:\n  port: 8080" --type yaml
# 或者从文件读取配置内容
./nacos-cli config set application.yml DEFAULT_GROUP --file ./config.yml --type yaml
./nacos-cli config list
./nacos-cli config export ./backup

# 方式2：使用命令行参数
./nacos-cli --server http://localhost:8848 --username admin --password admin123 config list
```

## 配置文件格式

配置文件 `~/.nacos-cli.yaml` 示例：

```yaml
server: http://localhost:8848
username: admin
password: admin123
namespace: dev
```
