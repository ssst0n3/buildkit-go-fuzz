package containerimage

import (
	"bytes"
	"context"
	"testing"

	"github.com/moby/buildkit/exporter/containerimage/exptypes"
)

func FuzzLoad(f *testing.F) {
	// 添加种子语料库，包含一些有意义的初始输入。
	// 格式是一个由空字节分隔的字符串序列：key1\x00val1\x00key2\x00val2\x00...
	f.Add([]byte(exptypes.OptKeyName + "\x00testimage\x00"))
	f.Add([]byte(exptypes.OptKeyOCITypes + "\x00true\x00"))
	f.Add([]byte(exptypes.OptKeyOCITypes + "\x00not-a-bool\x00"))
	f.Add([]byte("compression\x00gzip\x00compression-level\x005\x00"))
	f.Add([]byte("source-date-epoch\x001672531200\x00"))
	f.Add([]byte("invalid-key\x00some-value\x00"))
	f.Add([]byte(""))         // 空输入
	f.Add([]byte("\x00\x00")) // 空的键和值

	// Fuzz引擎将在此处注入随机生成的字节数据
	f.Fuzz(func(t *testing.T, data []byte) {
		// 从Fuzz引擎提供的原始字节数据创建map。
		// 我们将数据解释为由空字节分隔的键值对序列。
		// 这种方法允许Fuzz引擎生成各种各样的map结构。
		parts := bytes.Split(data, []byte{0})
		opt := make(map[string]string)

		// 我们遍历切片，每次取两个部分组成一个键值对。
		// 如果部分数量为奇数，最后一个将被忽略。
		for i := 0; i+1 < len(parts); i += 2 {
			key := string(parts[i])
			value := string(parts[i+1])
			opt[key] = value
		}

		c := &ImageCommitOpts{}
		ctx := context.Background()

		// 调用Load函数。Fuzz测试的目标是找到导致程序恐慌（panic）的输入。
		// 该函数对于格式错误的输入应该返回error，所以我们不需要检查返回的错误。
		// 如果发生恐慌，Fuzz引擎会自动报告失败并保存导致问题的输入。
		_, _ = c.Load(ctx, opt)
	})
}
