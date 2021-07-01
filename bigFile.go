package utils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// BigFileReader 大文件读结构
type BigFileReader struct {
	Path      string   // 读取路径
	File      *os.File // 文件操作指针
	Error     error    // 错误
	SliceSize int64    //
}

// BigFileWriter 大文件写结构
type BigFileWriter struct {
	UniqueID       string   // 源文件的唯一标识符（缓存）
	Path           string   // 写入路径
	File           *os.File // 文件操作指针
	Error          error    // 错误
	CurrentSize    int64    // 当前已写入文件大小
	CurrentPercent float64  // 当前已写入百分比
	*sync.Once              // 只创建一次空白文件
	*sync.Mutex             // 用来计算已经接收的文件数
}

// BigFileSlice 大文件切片结构体
type BigFileSlice struct {
	UniqueID string // 源文件的唯一标识符
	Total    int64  // 总文件大小
	Size     int64  // 本切片大小
	Offset   int64  // 本切片的偏移量
	Content  []byte // 本切片的内容
}

// 并发接收多个文件的多个分片时，要缓存 BigFileWriter
type BigFileWriterCache struct {
	*sync.RWMutex // 用来防止并发
	Cache         map[string]*BigFileWriter
}

// BigFileWriterCache 实例化后的缓存
var BigFileWriterCacheInstance = &BigFileWriterCache{
	RWMutex: &sync.RWMutex{},
	Cache:   map[string]*BigFileWriter{},
}

// Get 从 BigFileWriterCache 中获得缓存，如果缓存不存在，则实例化
func (bfwc *BigFileWriterCache) Get(UniqueID string, writePath string) *BigFileWriter {
	bfwc.RLock()
	if w, ok := bfwc.Cache[UniqueID]; ok {
		bfwc.RUnlock()
		return w
	}
	bfwc.RUnlock()

	bfwc.Lock()
	bfwc.Unlock()

	w := &BigFileWriter{
		UniqueID: UniqueID,
		Path:     writePath,
		Once:     &sync.Once{},
		Mutex:    &sync.Mutex{},
	}
	bfwc.Cache[UniqueID] = w
	return w
}

// Del 从 BigFileWriterCache 中删除使用完毕的缓存
func (bfwc *BigFileWriterCache) Del(UniqueID string) bool {
	bfwc.Lock()
	bfwc.Unlock()

	if _, ok := bfwc.Cache[UniqueID]; ok {
		delete(bfwc.Cache, UniqueID)
	}

	return true
}

// Read 分片读取大文件
func (bf *BigFileReader) Read(callback func(slice *BigFileSlice, n int, err error)) {
	if bf.SliceSize <= 0 {
		bf.Error = errors.New("请初始化切片大小")
		return
	}

	bf.File, bf.Error = os.Open(bf.Path)
	defer bf.File.Close()

	if nil != bf.Error {
		return
	}

	fi, _ := bf.File.Stat()
	total := fi.Size()
	if total <= 0 {
		bf.Error = errors.New("文件内容为空")
		return
	}

	// 计算文件唯一标识符
	var absPath string
	absPath, bf.Error = filepath.Abs(bf.Path)
	if nil != bf.Error {
		return
	}

	// 生成一个源文件的唯一ID
	UniqueID := SHA1(absPath)

	var offset int64 = 0

	for {
		sliceSize := bf.SliceSize

		// 小于分片大小
		if total < sliceSize {
			sliceSize = total
		}
		// 最后一天分片大小小于 bf.SliceSize
		if offset+bf.SliceSize > total {
			sliceSize = total - offset
		}

		// 表示已经取到结尾了，可以退出了
		if sliceSize <= 0 {
			break
		}

		sliceByte := make([]byte, sliceSize)
		n, err := bf.File.Read(sliceByte)
		if nil != err || io.EOF == err {
			break
		}

		slice := &BigFileSlice{
			UniqueID: UniqueID,
			Total:    total,
			Size:     sliceSize,
			Offset:   offset,
			Content:  sliceByte,
		}

		offset += sliceSize

		callback(slice, n, err)
	}
}

// CreateEmptyFile 创建一个和指定字节大小相同的空白文件，用于改写
func (bf *BigFileWriter) CreateEmptyFile(size int64) {
	if nil == bf.Once {
		bf.Error = errors.New("请初始化 LockOnce 锁")
		return
	}

	// 只能执行一次，后续都是写入
	bf.Do(func() {
		bf.File, bf.Error = os.OpenFile(bf.Path, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if nil != bf.Error {
			return
		}

		// 在指定位置写入一个字节，这样空文件就创建好了，等介于如下两行
		// file.Seek(26, 0)
		// file.Write([]byte{0})
		_, bf.Error = bf.File.WriteAt([]byte{0}, size-1)

		return
	})
}

// Write 分片写入指定文件
func (bf *BigFileWriter) Write(slice *BigFileSlice) {
	if nil == bf.Mutex {
		bf.Error = errors.New("请初始化 LockMutex 锁")
		return
	}

	// 先创建占位文件
	bf.CreateEmptyFile(slice.Total)

	// 并发统计
	bf.Lock()
	defer bf.Unlock()

	bf.CurrentSize += slice.Size
	bf.CurrentPercent = float64(bf.CurrentSize) / float64(slice.Total) * 100
	if _, bf.Error = bf.File.WriteAt(slice.Content, slice.Offset); nil != bf.Error {
		return
	}

	// 接受的文件大小和传过来的总大小一致，表示写完成，可以关毕
	if bf.CurrentSize >= slice.Total {
		// 清除缓存
		BigFileWriterCacheInstance.Del(bf.UniqueID)
		bf.File.Close()
	}
}

// BigFileCopy 大文件分片复制
func BigFileCopy(src string, dist string) error {
	reader := &BigFileReader{
		Path:      src,
		SliceSize: 4 << 20,
	}

	if nil != reader.Error {
		return reader.Error
	}

	reader.Read(func(slice *BigFileSlice, n int, err error) {
		w := BigFileWriterCacheInstance.Get(slice.UniqueID, dist)
		w.Write(slice)
	})

	if nil != reader.Error {
		return reader.Error
	}

	return nil
}
