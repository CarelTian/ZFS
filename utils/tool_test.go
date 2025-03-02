package utils

import (
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"testing"
)

// 假设 ZFSNode 的定义如下：

func TestTreeSerialization(t *testing.T) {
	root := &ZFSNode{
		Name: "root",
		Children: []*ZFSNode{
			{
				Name: "child1",
			},
			{
				Name: "child2",
				Children: []*ZFSNode{
					{Name: "grandchild1"},
					{Name: "grandchild2"},
				},
			},
		},
	}

	// 将树结构序列化为 YAML 数据
	data, err := yaml.Marshal(root)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 将 YAML 数据写入文件（如 tree.yaml）
	err = os.WriteFile("tree.yaml", data, 0644)
	if err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}

func TestTreeDeserialization(t *testing.T) {
	data, err := os.ReadFile("tree.yaml")
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	original := &ZFSNode{
		Name: "root",
		Children: []*ZFSNode{
			{
				Name: "child1",
			},
			{
				Name: "child2",
				Children: []*ZFSNode{
					{Name: "grandchild1"},
					{Name: "grandchild2"},
				},
			},
		},
	}

	var root ZFSNode
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 使用 reflect.DeepEqual 比较两个对象是否相同
	if !reflect.DeepEqual(original, &root) {
		t.Errorf("反序列化后的对象与原始对象不相等\n原始对象: %+v\n反序列化对象: %+v", original, root)
	}
}

func TestFormat(t *testing.T) {
	ret := FormatFileSize(512)
	expected := "512B"
	if ret != expected {
		t.Fatalf("格式化错误")
	}
	ret1 := FormatFileSize(1024)
	expected1 := "1.00KB"
	if ret1 != expected1 {
		t.Fatalf("格式化错误")
	}
	ret2 := FormatFileSize(1024 * 1024)
	expected2 := "1.00MB"
	if ret2 != expected2 {
		t.Fatalf("格式化错误")
	}

	ret3 := FormatFileSize(1024 * 1024 * 1024)
	expected3 := "1.00GB"
	if ret3 != expected3 {
		t.Fatalf("格式化错误")
	}
	ret4 := FormatFileSize(1024 * 1024 * 1024 * 1024)
	expected4 := "1024.00GB"
	if ret4 != expected4 {
		t.Fatalf("格式化错误")
	}
}

func TestRename(t *testing.T) {
	name := "ZFS.txt"
	expected := "ZFS(1).txt"
	ret := Rename(name, 1)
	if ret != expected {
		t.Fatalf("重命名出错，ret=%s, expected=%s", ret, expected)
	}
}
