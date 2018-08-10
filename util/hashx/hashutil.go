// hashx
package hashx

import (
	"crypto/md5"
	"encoding/hex"
)

//获取MD5加密串
func MD5(source string) string {
	h := md5.New()
	h.Write([]byte(source))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
