package compression

import "testing"

func FuzzParseAttributes(f *testing.F) {
	// === 种子语料库 ===
	// 每个种子包含属性值和决定该属性是否存在的布尔值。
	// (compression, force, level, hasCompression, hasForce, hasLevel)
	f.Add("gzip", "true", "5", true, true, true)                   // 所有键都存在
	f.Add("zstd", "", "", true, false, false)                      // 只有 compression 存在
	f.Add("", "true", "", false, true, false)                      // 只有 force-compression 存在
	f.Add("", "", "10", false, false, true)                        // 只有 compression-level 存在
	f.Add("invalid", "not-a-bool", "not-an-int", true, true, true) // 所有键都存在且值都无效

	// === Fuzzing 逻辑 ===
	f.Fuzz(func(t *testing.T, compression, force, level string, hasCompression, hasForce, hasLevel bool) {
		// 1. 根据Fuzzing引擎提供的随机值和布尔标志构造输入map
		attrs := make(map[string]string)

		if hasCompression {
			attrs[attrLayerCompression] = compression
		}
		if hasForce {
			attrs[attrForceCompression] = force
		}
		if hasLevel {
			attrs[attrCompressionLevel] = level
		}

		// 2. 调用待测试函数
		config, err := ParseAttributes(attrs)

		// 3. 健全性检查
		if err != nil {
			if config != (Config{}) {
				t.Errorf("当返回错误时，期望得到一个零值的Config对象, 但实际得到: %+v", config)
			}
		}
	})
}
