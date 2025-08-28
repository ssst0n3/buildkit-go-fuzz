package containerimage

import (
	"encoding/json"
	"testing"
	"time"

	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

func FuzzPatchImageConfig(f *testing.F) {
	// --- 种子语料库 (Seed Corpus) ---
	// 提供一些初始的、有效的输入，以引导 Fuzzer。

	// 场景 1: 一个相对完整的、有效的输入
	seedTime := time.Now().Add(-1 * time.Hour)
	seedDescs, _ := json.Marshal([]ocispecs.Descriptor{
		{
			MediaType: ocispecs.MediaTypeImageLayer,
			Digest:    "sha256:f7e399895698c7365fca52839e3a3c9444451694f28148b84f33b3edd94a8b66",
			Size:      1234,
			Annotations: map[string]string{
				"containerd.io/uncompressed": "sha256:d8122247402e21ded185127fe89aa33a5b30a78a01f35d55a4e4c23173253689",
			},
		},
	})
	seedHistory, _ := json.Marshal([]ocispecs.History{
		{
			Created:   &seedTime,
			CreatedBy: "/bin/sh -c #(nop) ADD file:abc in /",
		},
	})
	f.Add(
		[]byte(`{"os":"linux", "created":"2025-01-01T00:00:00Z"}`), // dt
		seedDescs,                                                  // descs
		seedHistory,                                                // history
		[]byte("cache_data"),                                       // cache
		seedTime.Unix(),                                            // epochUnix
		false,                                                      // epochIsNil
	)

	// 场景 2: 输入包含空值或 nil
	f.Add(
		[]byte(`{}`), // dt
		[]byte("[]"), // descs
		[]byte("[]"), // history
		[]byte(nil),  // cache
		int64(0),     // epochUnix
		true,         // epochIsNil
	)

	// --- Fuzzing 逻辑 ---
	f.Fuzz(func(t *testing.T, dt []byte, descsBytes []byte, historyBytes []byte, cache []byte, epochUnix int64, epochIsNil bool) {
		// 将 Fuzzer 生成的原始数据转换为函数所需的复杂类型
		var descs []ocispecs.Descriptor
		// 如果 json.Unmarshal 失败，descs 将为空切片，这是一种有效的模糊测试情况
		_ = json.Unmarshal(descsBytes, &descs)

		var history []ocispecs.History
		// 同样，如果 Unmarshal 失败，history 也是空切片
		_ = json.Unmarshal(historyBytes, &history)

		var epoch *time.Time
		if !epochIsNil {
			// 从 fuzzed int64 创建时间
			tm := time.Unix(epochUnix, 0)
			epoch = &tm
		}

		// 调用目标函数。
		// Fuzz 测试的目标是发现导致程序崩溃（panic）的输入。
		// 因此，我们不需要断言返回值或错误。如果函数出现 panic，Fuzzer 会自动报告失败。
		// 函数返回的 error 是预期的行为（例如，对于无效的 JSON），而不是崩溃。
		_, _ = patchImageConfig(dt, descs, history, cache, epoch)
	})
}
