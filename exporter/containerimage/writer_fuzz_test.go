package containerimage

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/pkg/labels"
	"github.com/containerd/containerd/v2/plugins/content/local"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/moby/buildkit/exporter"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/snapshot"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/solver"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	digest "github.com/opencontainers/go-digest"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

// FuzzCommitDistributionManifest 為 ImageWriter 的 commitDistributionManifest 方法提供了一個 Fuzz 測試。
//
// 測試策略：
// 1. 使用 go test -fuzz 引擎生成隨機輸入數據。
// 2. 主要模糊化 `config` 位元組切片，因為它是從外部來源（元數據）傳入的，最容易出現格式問題。
// 3. 同時模糊化控制函數邏輯的關鍵參數，如 `opts` 中的布林值、`epoch` 時間和 `baseImg` 的內容。
// 4. 對於像 `cache.ImmutableRef` 這樣的複雜接口，使用 `nil` 進行簡化，因為函數內部有對 `nil` 的處理邏輯。
// 5. 使用內存中的 content.Store 來隔離測試，避免對文件系統的依賴。
//
// 這個測試的目的是確保無論輸入如何組合，`commitDistributionManifest` 都不會引發 panic。
func FuzzCommitDistributionManifest(f *testing.F) {
	// 添加一些有意義的種子數據，以指導 Fuzz 引擎
	// 種子 1: 一個有效的、最小化的配置
	now := time.Now().UTC()
	validConfig, _ := json.Marshal(map[string]interface{}{
		"os":           "linux",
		"architecture": "amd64",
		"history":      []ocispecs.History{{Created: &now, CreatedBy: "test"}},
	})
	f.Add(validConfig, true, true, true, int64(1234567890), []byte(`[{"created_by": "base"}]`), []byte("sha256:baselayer"))

	// 種子 2: 一個空的配置
	f.Add([]byte{}, false, false, false, int64(0), []byte{}, []byte{})

	// 種子 3: 無效的 JSON
	f.Add([]byte("{not json"), true, false, true, int64(1), []byte("invalid json"), []byte("invalid digest data"))

	// 種子 4: 合法 JSON，但結構不符
	f.Add([]byte(`{"foo": "bar"}`), false, true, false, int64(999999), []byte(`[]`), []byte{})

	// Fuzz 引擎將基於種子數據生成更多樣化的輸入
	f.Fuzz(func(
		t *testing.T,
		config []byte,
		ociTypes bool,
		rewriteTimestamp bool,
		hasEpoch bool,
		epochNanos int64,
		baseImgHistoryData []byte,
		baseImgDiffIDData []byte,
	) {
		// --- 測試設置 ---

		// **修正點**: 使用 t.TempDir() 為每次 Fuzz 迭代創建一個獨立的臨時目錄。
		// 這可以防止並行執行時發生文件系統衝突，確保測試的隔離性和健壯性。
		tempDir := t.TempDir()
		store, err := local.NewStore(tempDir)
		if err != nil {
			t.Fatalf("failed to create local store in temp dir %s: %v", tempDir, err)
		}

		cs := containerdsnapshot.NewContentStore(store, "buildkit")
		// 創建 ImageWriter 實例
		ic, err := NewImageWriter(WriterOpt{
			ContentStore: cs,
			// 對於此函數，Snapshotter、Applier 和 Differ 不是必需的，可以為 nil
		})
		if err != nil {
			t.Fatalf("無法創建 ImageWriter: %v", err)
		}

		ctx := context.Background()
		sg := session.NewGroup("")

		// --- 根據模糊化輸入構造複雜參數 ---
		// 1. 構造 ImageCommitOpts
		var testEpoch *time.Time
		if hasEpoch {
			tm := time.Unix(0, epochNanos)
			testEpoch = &tm
		}
		opts := &ImageCommitOpts{
			OCITypes:         ociTypes,
			RewriteTimestamp: rewriteTimestamp, // 雖然此函數不直接使用，但相關邏輯依賴 epoch
			Annotations:      nil,              // 使用空的 annotations
		}

		// 2. 構造 solver.Remote
		// 創建一個帶有單個描述符的 remote，該描述符指向一個已寫入 content store 的 blob
		layerData := []byte("fuzz layer data")
		layerDigest := digest.FromBytes(layerData)
		layerDesc := ocispecs.Descriptor{
			MediaType: ocispecs.MediaTypeImageLayer,
			Digest:    layerDigest,
			Size:      int64(len(layerData)),
			Annotations: map[string]string{
				labels.LabelUncompressed: layerDigest.String(),
			},
		}
		if err := content.WriteBlob(ctx, cs, layerDigest.String(), bytes.NewReader(layerData), layerDesc); err != nil {
			t.Fatalf("寫入 blob 失敗: %v", err)
		}
		remote := &solver.Remote{
			Provider:    cs,
			Descriptors: []ocispecs.Descriptor{layerDesc},
		}

		// 3. 構造 baseImg
		var baseImg *dockerspec.DockerOCIImage
		if len(baseImgHistoryData) > 0 || len(baseImgDiffIDData) > 0 {
			baseImg = &dockerspec.DockerOCIImage{}
			// 嘗試從模糊化數據中解析 history，即使失敗也沒關係，這也是一種測試場景
			_ = json.Unmarshal(baseImgHistoryData, &baseImg.History)
			// 從模糊化數據中構造 DiffIDs
			if len(baseImgDiffIDData) > 0 {
				baseImg.RootFS.DiffIDs = []digest.Digest{digest.FromBytes(baseImgDiffIDData)}
			}
		}

		// --- 調用目標函數 ---
		// 我們預期函數可能會返回錯誤，但絕不能 panic。
		// Fuzz 測試框架會自動捕獲 panic。
		_, _, _ = ic.commitDistributionManifest(
			ctx,
			opts,
			nil, // 使用 nil 作為 cache.ImmutableRef，函數應能處理
			config,
			remote,
			&Annotations{}, // 使用空的 annotations
			nil,            // 使用 nil 作為 InlineCacheEntry
			testEpoch,
			sg,
			baseImg,
		)
	})
}

// FuzzCommitAttestationsManifest 是针对 commitAttestationsManifest 函数的模糊测试。
// 它的目标是发现任何可能导致程序崩溃（panic）的输入组合。
func FuzzCommitAttestationsManifest(f *testing.F) {
	// --- 1. 设置种子语料库 (Seed Corpus) ---
	// 这些种子为模糊测试引擎提供了有效的初始输入，以引导其进行变异。

	// 种子 1: 一个基本的有效证明，采用非 OCI Artifact 风格。
	seedTarget1 := ocispecs.Descriptor{
		MediaType: ocispecs.MediaTypeImageManifest,
		Digest:    "sha256:f54a5821fe5da5435027544002392690d8e40c59292a24e47c32888497a89e83",
		Size:      708,
	}
	seedStatements1 := []intoto.Statement{
		{
			StatementHeader: intoto.StatementHeader{
				Type:          intoto.StatementInTotoV01,
				PredicateType: "https://example.com/my-predicate/v1",
				Subject: []intoto.Subject{
					{
						Name: "pkg:docker/hello-world@sha256:f54a5821fe5da5435027544002392690d8e40c59292a24e47c32888497a89e83",
						Digest: map[string]string{
							"sha256": "f54a5821fe5da5435027544002392690d8e40c59292a24e47c32888497a89e83",
						},
					},
				},
			},
			Predicate: json.RawMessage(`{"key":"value"}`),
		},
	}
	targetJSON1, _ := json.Marshal(seedTarget1)
	statementsJSON1, _ := json.Marshal(seedStatements1)
	f.Add(targetJSON1, statementsJSON1, false, true) // ociArtifact=false, ociTypes=true

	// 种子 2: 包含两个证明，采用 OCI Artifact 风格。
	seedStatements2 := append(seedStatements1, intoto.Statement{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: "https://example.com/another-predicate/v1",
		},
		Predicate: json.RawMessage(`{"foo":"bar"}`),
	})
	statementsJSON2, _ := json.Marshal(seedStatements2)
	f.Add(targetJSON1, statementsJSON2, true, true) // ociArtifact=true, ociTypes=true

	// 种子 3: 证明列表为空数组 `[]`。
	f.Add(targetJSON1, []byte("[]"), false, false)

	// 种子 4: 证明列表为 `null`。
	f.Add(targetJSON1, []byte("null"), true, false)

	// --- 2. 定义 Fuzzing 目标函数 ---
	// 这是模糊测试的核心逻辑，它会使用引擎生成的随机数据来执行测试。
	f.Fuzz(func(t *testing.T, targetJSON []byte, statementsJSON []byte, ociArtifact bool, ociTypes bool) {
		tempDir := t.TempDir()
		store, err := local.NewStore(tempDir)
		if err != nil {
			t.Fatalf("failed to create local store in temp dir %s: %v", tempDir, err)
		}
		iw, err := NewImageWriter(WriterOpt{
			ContentStore: store,
			// Snapshotter, Applier, 和 Differ 在 commitAttestationsManifest 中未使用，
			// 因此可以安全地将它们设置为 nil。
		})
		if err != nil {
			t.Fatalf("创建 ImageWriter 失败: %v", err)
		}

		// b. 反序列化 Fuzzing 引擎提供的随机数据。
		// 如果数据不是有效的 JSON，则跳过本次迭代。我们的目标是测试函数逻辑，而非 JSON 解析器。
		var target ocispecs.Descriptor
		if err := json.Unmarshal(targetJSON, &target); err != nil {
			t.Skip()
		}

		var statements []intoto.Statement
		if err := json.Unmarshal(statementsJSON, &statements); err != nil {
			t.Skip()
		}

		// c. 根据 Fuzzing 输入设置选项。
		opts := &ImageCommitOpts{
			OCITypes: ociTypes,
		}

		// d. 调用被测试的函数。
		// Fuzzing 框架会自动捕获任何由此调用引发的 panic。
		// 我们不需要检查返回的描述符或错误，因为测试的目标是发现崩溃。
		_, _ = iw.commitAttestationsManifest(context.Background(), opts, target, statements, ociArtifact)
	})
}

type nopSnapshotter struct{ snapshot.Snapshotter }
type nopApplier struct{}

// --- 2. 随机工具 ------------------------------------------------------------

func randBool(r *rand.Rand) bool { return r.Intn(2) == 0 }

// 返回 0~n-1
func randN(r *rand.Rand, n int) int { return r.Intn(n) }

// 生成若干随机平台标识
func randomPlatforms(r *rand.Rand) []string {
	all := []string{
		"linux/amd64", "linux/arm64/v8", "windows/amd64", "linux/s390x",
	}
	// 随机取 1~len(all) 个
	num := r.Intn(len(all)) + 1
	r.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	return all[:num]
}

func FuzzCommit(f *testing.F) {
	// 多样 seed，可直接触发一些错误返回
	f.Add([]byte(`{}`))                    // 空 config
	f.Add([]byte(`invalid-json`))          // 非法 config
	f.Add(bytes.Repeat([]byte{'a'}, 9000)) // 大尺寸 config
	f.Add([]byte(`{"history":[{"empty_layer":true}]}`))

	f.Fuzz(func(t *testing.T, cfg []byte) {
		// 确保 fuzz 崩溃不会杀掉整个进程
		defer func() {
			if v := recover(); v != nil {
				t.Errorf("panic: %v", v)
			}
		}()

		r := rand.New(rand.NewSource(int64(time.Now().UnixNano())))

		// 1. content store
		tempDir := t.TempDir()
		store, err := local.NewStore(tempDir)
		if err != nil {
			t.Fatalf("failed to create local store in temp dir %s: %v", tempDir, err)
		}

		// 2. ImageWriter
		iw, _ := NewImageWriter(WriterOpt{
			ContentStore: store,
			Snapshotter:  &nopSnapshotter{},
		})

		// 3. 随机 platforms / refs
		plats := randomPlatforms(r)
		platMap := map[string]exptypes.Platforms{}
		refs := map[string]struct{}{} // 这里只需要 map key 占位
		for _, p := range plats {
			platMap[p] = exptypes.Platforms{}
			refs[p] = struct{}{}
		}
		platBytes, _ := json.Marshal(map[string]any{"platforms": platMap})

		// 4. exporter.Source
		src := &exporter.Source{
			Metadata: map[string][]byte{
				exptypes.ExporterImageConfigKey: cfg,
				exptypes.ExporterPlatformsKey:   platBytes,
				"random":                        []byte("value"),
			},
			Refs: nil, // 我们不真正构造 cache.ImmutableRef，走空 fast-path
		}

		// 5. ImageCommitOpts（随机开关）
		opts := &ImageCommitOpts{
			OCITypes:                randBool(r),
			RewriteTimestamp:        randBool(r),
			ForceInlineAttestations: randBool(r),
		}
		if randBool(r) {
			epoch := time.Unix(int64(r.Intn(1<<30)), 0).UTC()
			opts.Epoch = &epoch
		}

		// 6. 执行
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, _ = iw.Commit(ctx, src, "", nil, opts)
	})
}
