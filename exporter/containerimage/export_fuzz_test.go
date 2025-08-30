package containerimage

import (
	"context"
	"strings"
	"testing"

	"github.com/moby/buildkit/session"
)

func FuzzResolve(f *testing.F) {
	// 添加种子语料，为fuzzer提供一些初始输入。
	// 这些种子应该覆盖一些典型的和边缘的场景。

	// 场景1: 一个有效的、常见的用例
	f.Add(1, "name=my-registry/my-image:latest;push=true")
	// 场景2: 缺少必需的 "name" 属性
	f.Add(123, "push=true")
	// 场景3: 一个空的选项字符串
	f.Add(456, "")
	// 场景4: 值格式错误 (例如，push应该是布尔值)
	f.Add(789, "name=my-image;push=not-a-boolean")
	// 场景5: 包含空键或空值
	f.Add(0, "name=;=empty-key")

	// f.Fuzz 定义了fuzzing的目标函数。
	// fuzzer会生成 id 和 optStr 的随机值。
	f.Fuzz(func(t *testing.T, id int, optStr string) {
		// 1. 准备环境: 创建 imageExporter 实例
		// 在真实的buildkit测试环境中，通常会使用伪造的(fake)或模拟的(mock)依赖。
		sm, err := session.NewManager()
		if err != nil {
			t.Fatalf("无法创建 fake SourceManager: %v", err)
		}
		// 创建 imageExporter 实例，并注入依赖
		exporter := &imageExporter{
			opt: Opt{
				SessionManager: sm,
			},
		}

		// 2. 解析输入: 将fuzzer生成的字符串转换为 map[string]string
		// 我们定义一个简单的格式： "key1=value1;key2=value2"
		opts := make(map[string]string)
		pairs := strings.Split(optStr, ";")
		for _, pair := range pairs {
			// 使用 SplitN 确保值中可以包含 "="
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 && kv[0] != "" {
				opts[kv[0]] = kv[1]
			}
		}

		// 3. 调用目标函数
		// Fuzzing引擎会自动检测并报告任何由此调用引起的panic。
		_, _ = exporter.Resolve(context.Background(), id, opts)
	})
}
