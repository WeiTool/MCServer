// Package stream 提供进程日志流的读取与编码转换
package stream

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// StreamLines 逐行读取管道内容并交给回调处理
// 对读取的原始字节做编码识别：若为合法 UTF-8 则直接使用；
// 否则按 GBK 解码（Windows 中文系统上 Java 默认输出 GBK），统一转换为 UTF-8 避免中文乱码
// onLine 每收到一行（已去除行尾换行符）即调用，用于日志推送与类型检测
func StreamLines(reader io.Reader, onLine func(line string)) {
	bufReader := bufio.NewReaderSize(reader, 64*1024)
	for {
		// 逐行读取原始字节（含换行符）
		line, err := bufReader.ReadBytes('\n')
		if len(line) > 0 {
			text := DecodeToUTF8(line)
			cleanLine := strings.TrimRight(text, "\r\n")
			onLine(cleanLine)
		}
		if err != nil {
			// 读到管道末尾（io.EOF）即结束；其余错误也视为结束
			break
		}
	}
}

// DecodeToUTF8 将可能为 GBK 的字节转换为 UTF-8 字符串
// 若输入本身是合法 UTF-8，则原样返回；否则用 GBK 解码
// 参数 rawBytes 为从进程管道读到的原始字节（可能含换行符）
func DecodeToUTF8(rawBytes []byte) string {
	// 去除末尾换行符后再校验，避免截断的多字节 UTF-8 字符被误判
	trimmed := rawBytes
	// 若末尾是换行符 \n，则去掉它
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\n' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	// 若末尾是回车符 \r，则再去掉它（处理 Windows 的 \r\n 行尾）
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\r' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	// 若去除行尾后的字节是合法 UTF-8，则直接按 UTF-8 转成字符串返回
	if utf8.Valid(trimmed) {
		return string(rawBytes)
	}
	// 非法 UTF-8（很可能是 GBK），按 GBK 解码
	utf8Str, err := simplifiedchinese.GBK.NewDecoder().String(string(trimmed))
	if err != nil {
		// 解码失败则退回原始字节，避免完全丢失内容
		return string(rawBytes)
	}
	return utf8Str
}
