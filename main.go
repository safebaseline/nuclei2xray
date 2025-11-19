package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nuclei2xray/converter"
	"nuclei2xray/poc"

	"gopkg.in/yaml.v3"
)

const failureLogFile = "conversion_failures.log"

func main() {
	var (
		inputFile  = flag.String("i", "", "输入文件路径（Nuclei POC文件）")
		outputFile = flag.String("o", "", "输出文件路径（Xray POC文件，默认为与输入文件同名）")
		dir        = flag.String("d", "", "批量转换目录（转换目录下所有.yaml文件）")
		help       = flag.Bool("h", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		printUsage()
		return
	}

	if *dir != "" {
		// 批量转换模式
		outputDir := ""
		if *outputFile != "" {
			// 如果指定了-o参数，在批量模式下作为输出目录使用
			outputDir = *outputFile
		}
		if err := batchConvert(*dir, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "批量转换失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "错误: 请指定输入文件或使用 -d 参数指定目录\n")
		printUsage()
		os.Exit(1)
	}

	// 单文件转换模式
	if err := convertFile(*inputFile, *outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "转换失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("转换完成!")
}

func printUsage() {
	fmt.Println("Nuclei POC 转 Xray POC 工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  nuclei2xray -i <输入文件> [-o <输出文件>]")
	fmt.Println("  nuclei2xray -d <目录> [-o <输出目录>]")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  -i string    输入文件路径（Nuclei POC文件）")
	fmt.Println("  -o string    输出文件路径（单文件模式）或输出目录（批量模式，可选）")
	fmt.Println("  -d string    批量转换目录（转换目录下所有.yaml文件）")
	fmt.Println("  -h           显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  nuclei2xray -i poc.yaml -o xray_poc.yml")
	fmt.Println("  nuclei2xray -d ./nuclei_pocs/")
	fmt.Println("  nuclei2xray -d ./nuclei_pocs/ -o ./xray_poc/")
}

func convertFile(inputFile, outputFile string) error {
	err := convertFileToDir(inputFile, outputFile, "")
	if err != nil {
		recordConversionFailure(inputFile, err)
	}
	return err
}

func convertFileToDir(inputFile, outputFile, outputDir string) error {
	// 读取输入文件
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析Nuclei POC
	var nucleiPOC poc.NucleiPOC
	if err := yaml.Unmarshal(data, &nucleiPOC); err != nil {
		return fmt.Errorf("解析YAML失败: %w", err)
	}

	// 转换为Xray POC
	xrayPOC, err := converter.ConvertNucleiToXray(&nucleiPOC, inputFile)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	// 生成输出文件名
	if outputFile == "" {
		baseName := filepath.Base(inputFile)
		baseName = ensureYMLExtension(baseName)

		if outputDir != "" {
			// 确保输出目录存在
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("创建输出目录失败: %w", err)
			}
			outputFile = filepath.Join(outputDir, baseName)
		} else {
			inputDir := filepath.Dir(inputFile)
			if inputDir == "." || inputDir == "" {
				outputFile = baseName
			} else {
				outputFile = filepath.Join(inputDir, baseName)
			}
		}
	}

	// 确保输出文件使用 .yml 后缀
	outputFile = ensureYMLExtension(outputFile)

	// 写入输出文件
	outputData, err := yaml.Marshal(xrayPOC)
	if err != nil {
		return fmt.Errorf("生成YAML失败: %w", err)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("成功转换: %s -> %s\n", inputFile, outputFile)
	return nil
}

func batchConvert(dir, outputDir string) error {
	// 检查目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("目录不存在: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("不是目录: %s", dir)
	}

	// 读取目录下所有.yaml文件（递归）
	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过xray_poc目录（已经是Xray格式，不需要转换）
		if info.IsDir() && filepath.Base(path) == "xray_poc" {
			return filepath.SkipDir
		}
		// 跳过已经是转换后的文件
		if !info.IsDir() {
			baseName := filepath.Base(path)
			if strings.HasSuffix(baseName, "_xray.yaml") ||
				strings.HasSuffix(baseName, "_xray.yml") ||
				strings.HasSuffix(baseName, ".xray.yaml") ||
				strings.HasSuffix(baseName, ".xray.yml") {
				return nil
			}
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("目录 %s 中没有找到.yaml或.yml文件\n", dir)
		return nil
	}

	// 如果指定了输出目录，确保目录存在
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
		fmt.Printf("输出目录: %s\n", outputDir)
	}

	fmt.Printf("找到 %d 个文件，开始批量转换...\n", len(files))

	successCount := 0
	failCount := 0
	var failedFiles []string

	for _, file := range files {
		fmt.Printf("\n处理文件: %s\n", file)

		// 如果指定了输出目录，需要计算相对路径以保持目录结构
		var targetOutputDir string
		if outputDir != "" {
			// 计算文件相对于输入目录的路径
			relPath, err := filepath.Rel(dir, filepath.Dir(file))
			if err != nil {
				// 如果无法计算相对路径，直接使用输出目录
				targetOutputDir = outputDir
			} else {
				// 保持相对目录结构
				if relPath == "." {
					targetOutputDir = outputDir
				} else {
					targetOutputDir = filepath.Join(outputDir, relPath)
				}
			}
		}

		if err := convertFileToDir(file, "", targetOutputDir); err != nil {
			fmt.Fprintf(os.Stderr, "  错误: %v\n", err)
			failCount++
			failedFiles = append(failedFiles, file)
			recordConversionFailure(file, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n批量转换完成: 成功 %d 个, 失败 %d 个\n", successCount, failCount)
	if failCount > 0 {
		fmt.Println("\n以下文件转换失败（详细信息已记录在 conversion_failures.log ）:")
		for _, f := range failedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
	return nil
}

func recordConversionFailure(file string, err error) {
	if file == "" || err == nil {
		return
	}
	entry := fmt.Sprintf("%s\t%s\t%v\n", time.Now().Format(time.RFC3339), file, err)
	f, openErr := os.OpenFile(failureLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "无法记录失败文件 %s: %v\n", file, openErr)
		return
	}
	defer f.Close()
	if _, writeErr := f.WriteString(entry); writeErr != nil {
		fmt.Fprintf(os.Stderr, "无法写入失败记录 %s: %v\n", file, writeErr)
	}
}

func ensureYMLExtension(path string) string {
	if path == "" {
		return ""
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return path + ".yml"
	}
	if strings.EqualFold(ext, ".yml") {
		return path[:len(path)-len(ext)] + ".yml"
	}
	return path[:len(path)-len(ext)] + ".yml"
}
