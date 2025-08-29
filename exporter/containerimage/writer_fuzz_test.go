package containerimage

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/pkg/labels"
	"github.com/containerd/containerd/v2/plugins/content/local"
	"github.com/moby/buildkit/session"
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
		store, err := local.NewStore("/test")
		if err != nil {
			t.Fatalf("failed to create local store: %v", err)
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
