package containerimage

import (
	"bytes"
	"testing"

	"github.com/containerd/platforms"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

func FuzzParseAnnotations(f *testing.F) {
	// 1. 添加种子语料库 (Seed Corpus)
	// 我们使用 strings.Join 和 bytes.Join 来方便地构建遵循我们自定义格式的种子。
	pairSep := []byte{0, 0} // \x00\x00
	kvSep := []byte{0}      // \x00

	// 场景: 空的输入
	f.Add([]byte{})

	// 场景: 包含一个有效的索引注解
	f.Add(bytes.Join([][]byte{
		[]byte("buildkit.annotation.index.version"),
		[]byte("1.0"),
	}, kvSep))

	// 场景: 包含一个有效的、带平台的清单注解
	f.Add(bytes.Join([][]byte{
		[]byte("buildkit.annotation.manifest.author/linux/amd64"),
		[]byte("test-user"),
	}, kvSep))

	// 场景: 包含一个非注解的键，应被放入"rest" map中
	f.Add(bytes.Join([][]byte{
		[]byte("custom.key"),
		[]byte("custom-value"),
	}, kvSep))

	// 场景: 混合了多个键值对
	f.Add(bytes.Join([][]byte{
		bytes.Join([][]byte{[]byte("buildkit.annotation.index.foo"), []byte("bar")}, kvSep),
		bytes.Join([][]byte{[]byte("buildkit.annotation.manifest.foo/linux/amd64"), []byte("baz")}, kvSep),
		bytes.Join([][]byte{[]byte("some.other.key"), []byte("value")}, kvSep),
	}, pairSep))

	// 场景: 包含格式错误的键 (对于ParseAnnotations)
	f.Add(bytes.Join([][]byte{
		[]byte("buildkit.annotation."),
		[]byte("malformed key"),
	}, kvSep))

	// 场景: 包含一个没有值的键 (测试我们的解析逻辑)
	f.Add([]byte("key-with-no-value"))

	// 场景: 值中包含分隔符 (我们的解析逻辑应该能正确处理)
	f.Add([]byte("key\x00value\x00with\x00separator"))

	// 2. 定义Fuzz目标函数
	f.Fuzz(func(t *testing.T, data []byte) {
		// --- 测试驱动代码 (Test Harness) ---
		// 将fuzzer提供的原始 []byte 解析成 map[string][]byte
		inputMap := make(map[string][]byte)

		// 如果输入为空，直接跳到函数调用
		if len(data) == 0 {
			// 继续执行以测试目标函数对空map的处理
		} else {
			// 使用双空字节分割成键值对
			pairs := bytes.Split(data, pairSep)
			for _, pair := range pairs {
				// 使用单个空字节将键和值分开。
				// SplitN确保我们只分割一次，这样值本身就可以包含空字节。
				parts := bytes.SplitN(pair, kvSep, 2)
				if len(parts) == 2 {
					// 只有当键和值都存在时才添加到map中
					key := string(parts[0])
					value := parts[1]
					// 为防止fuzzer生成空字符串键（在map中是有效的，但可能不是我们想要的），
					// 我们可以选择忽略它们。
					if key != "" {
						inputMap[key] = value
					}
				}
			}
		}
		// --- 驱动代码结束 ---

		// --- 目标函数调用 ---
		// 现在我们有了一个安全的、构造好的map，可以放心地调用目标函数了。
		ag, rest, err := ParseAnnotations(inputMap)
		// --- 目标函数调用结束 ---

		// 3. 对结果进行断言和一致性检查
		if err != nil {
			if ag != nil || rest != nil {
				t.Errorf("当返回错误时，期望AnnotationsGroup和rest map为nil，但实际并非如此")
			}
			return
		}

		if ag == nil || rest == nil {
			t.Errorf("当没有错误时，期望AnnotationsGroup和rest map为非nil，但实际为nil")
		}

		outputKeyCount := len(rest)
		for _, annotations := range ag {
			outputKeyCount += len(annotations.Index)
			outputKeyCount += len(annotations.IndexDescriptor)
			outputKeyCount += len(annotations.Manifest)
			outputKeyCount += len(annotations.ManifestDescriptor)
		}

		if outputKeyCount != len(inputMap) {
			t.Errorf("输入和输出的键数量不匹配：输入有 %d 个键，但处理后得到 %d 个", len(inputMap), outputKeyCount)
		}
	})
}

// encodeMap 仅用于为 Fuzz 函数提供种子数据。
func encodeMap(m map[string][]byte) []byte {
	var buf bytes.Buffer
	for k, v := range m {
		buf.WriteString(k)
		buf.WriteByte(0)
		buf.Write(v)
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

// decodeMap 把模糊输入拆成键值对：key\x00value\x00key\x00value...
// 如果最后剩下一个孤立的 key，则丢弃。
func decodeMap(b []byte) map[string][]byte {
	out := make(map[string][]byte)
	parts := bytes.Split(b, []byte{0})
	for i := 0; i+1 < len(parts); i += 2 {
		out[string(parts[i])] = parts[i+1]
	}
	return out
}

func FuzzAnnotationsGroupPlatform(f *testing.F) {
	// ---- Seed corpus ----
	f.Add(encodeMap(map[string][]byte{
		// 尽量放一个看起来能被 exptypes.ParseAnnotationKey 识别的 key，增加有效覆盖
		"containerimage.manifest": []byte(`{"dummy":"value"}`),
	}))
	f.Add(encodeMap(map[string][]byte{
		"totally-random-key": []byte("abcdef"),
	}))

	// ---- Fuzz function ----
	f.Fuzz(func(t *testing.T, raw []byte) {
		// 1. 把随机输入解码成 map[string][]byte
		data := decodeMap(raw)

		// 2. 通过 ParseAnnotations 获取 AnnotationsGroup
		ag, _, err := ParseAnnotations(data)
		if err != nil {
			// 即使解析失败，也只是返回 error，不应 panic
			return
		}

		// 3. 为更多覆盖准备 platform 列表
		var platformsToTest []*ocispecs.Platform
		platformsToTest = append(platformsToTest, nil) // 必测 nil

		// 尝试把剩余字节解析成一个平台字符串
		if p, err := platforms.Parse(string(raw)); err == nil {
			platformsToTest = append(platformsToTest, &p)
		}

		// 4. 调用目标函数，确保不会 panic
		for _, p := range platformsToTest {
			_ = ag.Platform(p)
		}
	})
}

// FuzzAnnotationsGroupMerge 是针对 AnnotationsGroup 的 Merge 方法编写的模糊测试用例。
//
// 测试逻辑如下：
//  1. Fuzzer 生成两个独立的字节切片 `data1` 和 `data2`。
//  2. 我们将这两个字节切片通过一个辅助函数 `bytesToMap` 转换为 `map[string][]byte` 的格式。
//     这个转换过程模拟了从文本输入（例如 key=value 格式）生成 map 的过程。
//  3. 这两个 map 分别被传入 `ParseAnnotations` 函数，从而生成两个 `AnnotationsGroup` 实例，即 `ag` 和 `other`。
//     这一步确保了我们提供给 `Merge` 函数的输入是“合法”且符合预期的，因为它们是通过应用的解析逻辑生成的。
//  4. 最后，调用 `ag.Merge(other)`。Fuzz 测试的主要目标是发现任何可能导致程序崩溃 (panic) 的边界情况。
//     如果 `Merge` 函数在处理任何由 fuzzer 生成的输入时发生 panic，测试将自动失败并报告问题。
func FuzzAnnotationsGroupMerge(f *testing.F) {
	// 为 fuzzer 提供一些初始种子（seed corpus），以引导其发现更有趣的输入。
	// 这些种子覆盖了多种场景，例如：
	// - 两个 AnnotationsGroup 都有数据，且部分 key 重叠。
	// - 其中一个 AnnotationsGroup 为空。
	// - 两个都为空。
	// - 包含无效的 annotation key 格式。
	f.Add(
		[]byte("buildkit.annotation.index.key1=val1\nbuildkit.annotation.manifest.linux/amd64.key2=val2"),
		[]byte("buildkit.annotation.index.descriptor.key3=val3\nbuildkit.annotation.manifest.linux/amd64.key2=new-val"),
	)
	f.Add([]byte(""), []byte("buildkit.annotation.index.key1=val1"))
	f.Add([]byte("buildkit.annotation.index.key1=val1"), []byte(""))
	f.Add([]byte{}, []byte{})
	f.Add([]byte("invalid-key-format-no-equals"), []byte("buildkit.annotation.index.key1=val1"))

	// bytesToMap 是一个辅助函数，用于将 fuzzer 提供的原始字节切片转换为 `map[string][]byte`。
	// 这是 `ParseAnnotations` 函数所需的输入格式。
	// 该函数通过换行符 `\n` 分割键值对，再通过等号 `=` 分割键和值。
	// 这种方法可以从非结构化的 fuzzer 输入中有效地生成结构化的 map 数据。
	bytesToMap := func(data []byte) map[string][]byte {
		m := make(map[string][]byte)
		pairs := bytes.Split(data, []byte("\n"))
		for _, pair := range pairs {
			if key, value, found := bytes.Cut(pair, []byte("=")); found {
				m[string(key)] = value
			}
		}
		return m
	}

	// Fuzz 主逻辑
	f.Fuzz(func(t *testing.T, data1 []byte, data2 []byte) {
		// 1. 从 fuzzer 的输入 `data1` 生成第一个 AnnotationsGroup `ag`。
		map1 := bytesToMap(data1)
		ag, _, err := ParseAnnotations(map1)
		if err != nil {
			// 如果 `ParseAnnotations` 返回错误，说明 fuzzer 生成的输入对于解析逻辑是无效的。
			// 我们应该跳过（return）这次测试，因为我们的目标是测试 `Merge` 函数，而不是 `ParseAnnotations`。
			return
		}

		// 2. 从 fuzzer 的输入 `data2` 生成第二个 AnnotationsGroup `other`。
		map2 := bytesToMap(data2)
		other, _, err := ParseAnnotations(map2)
		if err != nil {
			// 同上，跳过无效输入。
			return
		}

		// 3. 调用目标函数 `Merge`。
		// Fuzz 测试的核心在于发现导致程序崩溃 (panic) 的输入。
		// Go 的 fuzzing 引擎会自动检测并报告在此调用期间发生的任何 panic。
		// 注意：如果 `ParseAnnotations` 的输入 map 为空，它会返回一个 nil 的 `AnnotationsGroup`。
		// `Merge` 函数的设计能够正确处理接收者 `ag` 或参数 `other` 为 nil 的情况，这也是我们需要测试的场景之一。
		ag.Merge(other)
	})
}
