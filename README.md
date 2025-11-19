# Nuclei2Xray

一个用Go语言编写的工具，用于将Nuclei的POC（概念验证）转换为Xray的POC格式（AI写的，我只提供了概念）。

## 功能特性

- 支持将Nuclei YAML格式的POC转换为Xray YAML格式
- 支持HTTP和Network类型的POC转换
- 支持匹配器（matchers）和提取器（extractors）的转换
- 支持批量转换目录下的所有POC文件
- 自动处理请求方法、路径、headers、body等字段
- 支持raw请求格式解析

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd nuclei2xray
go mod download
go build -o nuclei2xray
```

## 使用方法

### 单文件转换

```bash
# 基本用法
nuclei2xray -i nuclei_poc.yaml -o xray_poc.yml

# 不指定输出文件（默认与输入文件同名，位于同一目录，会覆盖原POC）
nuclei2xray -i nuclei_poc.yaml

# 建议搭配 -o 指定输出目录以保留原始 Nuclei POC
```

### 批量转换

```bash
# 转换目录下所有.yaml和.yml文件（输出到原文件同目录）
nuclei2xray -d ./nuclei_pocs/

# 转换目录下所有.yaml和.yml文件到指定输出目录
nuclei2xray -d ./nuclei_pocs/ -o ./xray_poc/
```

### 查看帮助

```bash
nuclei2xray -h
```

## 转换说明

### 支持的Nuclei POC特性

- ✅ HTTP请求（GET, POST, PUT, DELETE等）
- ✅ Raw请求格式
- ✅ Headers和Body
- ✅ 匹配器（Matchers）：
  - Status匹配
  - Size匹配
  - Word匹配
  - Regex匹配
  - DSL匹配
  - Binary匹配
- ✅ 提取器（Extractors）：
  - Regex提取
  - JSON提取
  - XPath提取
  - DSL提取
- ✅ 网络请求（Network）
- ✅ 变量（Variables）
- ✅ 元数据（Info）

### 转换映射

| Nuclei字段 | Xray字段 | 说明 |
|-----------|---------|------|
| `info.name` | `name` | POC名称 |
| `info.author` | `detail.author` | 作者 |
| `info.description` | `detail.description` | 描述 |
| `info.reference` | `detail.links` | 参考链接 |
| `info.tags` | `detail.tags` | 标签 |
| `http[].method` | `rules[].method` | HTTP方法 |
| `http[].path` | `rules[].path` | 请求路径 |
| `http[].headers` | `rules[].headers` | 请求头 |
| `http[].body` | `rules[].body` | 请求体 |
| `http[].matchers` | `rules[].search` / `rules[].expression` | 匹配规则 |
| `http[].extractors` | `rules[].output` | 提取规则 |
| `variables` | `set` | 变量设置 |

## 示例

### Nuclei POC示例

```yaml
id: test-poc

info:
  name: Test Vulnerability
  author: test
  severity: high
  description: Test description

http:
  - method: GET
    path:
      - "{{BaseURL}}/test"
    matchers:
      - type: word
        words:
          - "vulnerable"
        condition: and
```

### 转换后的Xray POC

```yaml
name: Test Vulnerability
transport: http
rules:
  r0:
    request:
      cache: true
      method: GET
      path: "{{BaseURL}}/test"
      follow_redirects: true
    expression: response.status == 200 && response.body_string.contains('vulnerable')
expression: r0()
detail:
  author: test
  description: Test description
  tags: []
  vulnerability:
    level: high
```

## 注意事项

1. 某些复杂的Nuclei特性可能无法完全转换，需要手动调整
2. DSL表达式会尽量保持原样，但可能需要根据Xray的语法进行调整
3. 批量转换时，如果不指定输出目录（-o），输出文件会保存在与输入文件相同的目录下
4. 批量转换时，如果指定了输出目录（-o），会保持原有的目录结构
5. 批量转换过程中出现的失败会被追加写入项目根目录的`conversion_failures.log`文件，便于后续排查
6. 建议转换后验证POC的有效性

## 版本修改记录

### v1.1.1 (2025-XX-XX)

- **改进**: 默认输出文件名与输入文件完全一致（包括扩展名），便于在不同目录中保持命名统一；如果不指定`-o`，转换结果会覆盖原始文件

### v1.1.0 (2024-XX-XX)

- **修复**: 修复了path字段中`{{BaseURL}}`的处理问题
  - 添加了`removeBaseURL`函数，用于正确处理路径中的`{{BaseURL}}`占位符
  - 确保转换后的path字段以`/`开头，符合Xray的路径格式要求
  - 如果path仅包含`{{BaseURL}}`，转换后默认为`/`

### v1.0.0 (2024-XX-XX)

- **初始版本**: 实现完整的Nuclei到Xray的POC转换功能
  
  **核心功能**:
  - ✅ HTTP请求转换（支持GET、POST、PUT、DELETE等方法）
  - ✅ Network网络请求转换（TCP/UDP协议）
  - ✅ Raw请求格式解析和转换
  - ✅ 请求头（Headers）和请求体（Body）转换
  - ✅ 变量（Variables）转换，支持Nuclei变量表达式到Xray格式
  
  **匹配器（Matchers）支持**:
  - ✅ Status状态码匹配
  - ✅ Size响应大小匹配
  - ✅ Word关键词匹配
  - ✅ Regex正则表达式匹配
  - ✅ DSL表达式匹配
  - ✅ Binary二进制匹配
  - ✅ 支持AND/OR条件组合
  
  **提取器（Extractors）支持**:
  - ✅ Regex正则表达式提取
  - ✅ JSON路径提取
  - ✅ XPath提取
  - ✅ DSL表达式提取
  
  **元数据转换**:
  - ✅ POC名称、作者、描述转换
  - ✅ 参考链接（Reference）转换
  - ✅ 标签（Tags）转换
  - ✅ 严重程度（Severity）映射
  - ✅ CVE/CWE ID转换
  
  **批量处理**:
  - ✅ 支持单文件转换
  - ✅ 支持批量转换目录下所有YAML文件
  - ✅ 支持递归目录扫描
  - ✅ 支持保持目录结构输出
  - ✅ 自动跳过已转换文件（`.xray.yml`或旧版本生成的`_xray.yml`后缀）
  - ✅ 自动跳过`xray_poc`目录

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License


