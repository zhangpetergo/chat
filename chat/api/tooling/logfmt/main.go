package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LogEntry 结构体表示日志条目
type LogEntry struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func main() {
	// 假设日志来自标准输入（例如 `tail -f log.json | go run main.go`）
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text() // 读取日志行

		var log LogEntry
		err := json.Unmarshal([]byte(line), &log)
		if err != nil {
			fmt.Println("解析失败:", err)
			continue
		}

		// 格式化日志输出
		fmt.Printf("[%s] %s\n", log.Level, log.Timestamp)
		fmt.Println("  " + strings.ReplaceAll(log.Message, "\n", "\n  ")) // 处理换行
		fmt.Println("-----------------------------------")
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Println("读取日志时出错:", err)
	}
}
