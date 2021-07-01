# 常用 Golang 函数封装

## bigFile 大文件分片读写
```
package main

import (
	"fmt"

	"github.com/marknown/utils"
)

func main() {
	// 大文件读写分离（
	reader := &utils.BigFileReader{
		Path:      "读取文件路径",
		SliceSize: 4 << 20, // 分片大小 4MB
	}

	pathToWrite := "写入文件路径"

	// 分片读
	reader.Read(func(slice *utils.BigFileSlice, n int, err error) {
		// 以下代码为示例代码，自己可以做其它处理
		// 分片写
		w := utils.BigFileWriterCacheInstance.Get(slice.UniqueID, pathToWrite)
		w.Write(slice)

		fmt.Printf("\r已%.2f%% 完成 %dMB", w.CurrentPercent, w.CurrentSize>>20)
	})
}
```