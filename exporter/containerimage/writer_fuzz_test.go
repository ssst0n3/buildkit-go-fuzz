package containerimage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/containerd/containerd/content/local"
	"github.com/moby/buildkit/session"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/solver"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

func FuzzCommitDistributionManifest(f *testing.F) {
	// Seed corpus
	// The fuzzer will automatically run with these seed values to help guide fuzzing and create
	// a baseline for comparison.
	// 添加一些种子语料，以引导 Fuzz 引擎。
	// 所有 JSON 输入现在都是 []byte 类型，以匹配 Fuzz 目标签名。
	f.Add(
		true,
		[]byte(`{"os":"linux","architecture":"amd64"}`),
		[]byte(`[{"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 123, "digest": "sha256:f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"}]`),
		[]byte(`{"org.opencontainers.image.title":"my-image"}`),
		[]byte(`{}`),
		[]byte(`{"some":"cache"}`),
		true,
		int64(1672531200), // 2023-01-01
	)
	f.Add(
		false,
		[]byte(`not a valid json`),
		[]byte(`[]`),
		[]byte(`{}`),
		[]byte(`{}`),
		[]byte{},
		false,
		int64(0),
	)
	f.Add(
		true,
		[]byte(`{}`),
		[]byte(`[{"mediaType": "invalid"}]`),
		[]byte(`not a valid json`),
		[]byte(`not a valid json`),
		[]byte(nil),
		true,
		int64(-100), // 测试负数时间戳
	)
	f.Fuzz(func(t *testing.T,
		oci bool,
		configJSON []byte,                        // 已从 string 更改为 []byte
		remoteDescriptorsJSON []byte,             // 已从 string 更改为 []byte
		manifestAnnotationsJSON []byte,           // 已从 string 更改为 []byte
		manifestDescriptorAnnotationsJSON []byte, // 已从 string 更改为 []byte
		inlineCache []byte,
		epochValid bool,
		epochSeconds int64,
	) {
		// 准备
		store, err := local.NewStore("/test")
		if err != nil {
			t.Fatalf("failed to create local store: %v", err)
		}
		cs := containerdsnapshot.NewContentStore(store, "buildkit")
		// 函数并不直接使用Snapshotter/Applier/Differ, 但是patchImageLayers可能会用到. 由于使用 nil ref, 似乎可以避免这些.
		ic, err := NewImageWriter(WriterOpt{ContentStore: cs})
		if err != nil {
			t.Fatalf("failed to create ImageWriter: %v", err)
		}

		// 准备输入数据
		opts := &ImageCommitOpts{OCITypes: oci}
		config := []byte(configJSON)

		var remote solver.Remote
		// 我们需要设置provider 用于layer data.
		remote.Provider = cs
		// 将fuzzed descriptors 解码
		var descriptors []ocispecs.Descriptor
		if err := json.Unmarshal([]byte(remoteDescriptorsJSON), &descriptors); err == nil {
			remote.Descriptors = descriptors
		}

		var annotations Annotations
		// 将fuzzed annotations 解码
		if err := json.Unmarshal([]byte(manifestAnnotationsJSON), &annotations.Manifest); err != nil {
			// 失败是可以接受的，fuzzer 可以提供无效的JSON.
		}
		if err := json.Unmarshal([]byte(manifestDescriptorAnnotationsJSON), &annotations.ManifestDescriptor); err != nil {
			// same here
		}

		var epoch *time.Time
		if epochValid {
			// 使用非负的秒数来创建有效的时间
			if epochSeconds < 0 {
				epochSeconds = -epochSeconds
			}
			tm := time.Unix(epochSeconds, 0)
			epoch = &tm
		}

		// 让我们假设 patchImageLayers 是可用的.
		_, _, _ = ic.commitDistributionManifest(
			context.Background(),
			opts,
			nil, // ref
			config,
			&remote,
			&annotations,
			inlineCache,
			epoch,
			session.NewGroup("fuzz"),
		)
	})
}

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
