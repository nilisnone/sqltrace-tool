package tools

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// ExtractLine 从文件中提取行
func ExtractLine(fd *os.File, start int) (len_str string, next_pos int, ret_err error) {
	_, err := fd.Seek(int64(start), io.SeekStart)
	if err != nil {
		return "", 0, err
	}
	bs := ""
	bn := 0
	t_bn := 0
	// 找到下一行位置，和一整行内容
	offset_bn := 0
	for {
		b := make([]byte, 256)
		bn, err = fd.Read(b)
		// logE("read %d", bn)
		if err != nil && err != io.EOF {
			return "", 0, err
		}
		t_bn += bn
		bst := string(b[:])
		bst2 := string(bytes.TrimLeft(b, "\n")[:])
		// logE("got str: %s, total_bn:%d, start:%d", bst, t_bn, start)
		br := strings.Index(bst2, "\n")
		if bn > 0 && br == -1 {
			bs += bst2[:]
			continue
		}
		if br != -1 {
			bs += bst2[:br]
			// 包含换行，还需要去掉额外的换行之后的长度
			offset_bn = bn - len(bst[:br])
			// logE("offset bn: %d", offset_bn)
			break
		}
		if err == io.EOF {
			break
		}
	}
	if t_bn == 0 {
		return bs, start, nil
	}
	if offset_bn < 0 {
		offset_bn = 0
	}
	return bs, t_bn - offset_bn + start + 1, nil
}
