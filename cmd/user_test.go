package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadInputExit(t *testing.T) {
	// 模拟用户输入 "hello"，"world"，然后 "exit" 退出循环
	input := "hello\nworld\nexit\n"
	var out bytes.Buffer
	readInput(strings.NewReader(input), &out)

	expectedOutput := "请输入内容: 您输入了: hello\n" +
		"请输入内容: 您输入了: world\n" +
		"请输入内容: 退出程序\n"

	if out.String() != expectedOutput {
		t.Errorf("预期输出:\n%q\n实际输出:\n%q", expectedOutput, out.String())
	}
}
