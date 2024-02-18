package utils

import "strings"

// "https://www.chajiuqqq.cn/video/20mins/stream.mpd" -> "stream"
func ExtractMpdFileName(mpdUrl string) string {
	// 使用最后一个斜杠（/）的索引来提取文件名
	index := strings.LastIndex(mpdUrl, "/")
	if index == -1 {
		return "" // 如果URL中没有斜杠，则返回空字符串
	}

	// 提取文件名
	fileName := mpdUrl[index+1 : len(mpdUrl)-4]

	return fileName
}
