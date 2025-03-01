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
