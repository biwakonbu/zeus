package analysis

// itoa は int を文字列に変換するヘルパー関数
// strconv.Itoa と同等だが、パッケージ内で統一的に使用するため定義
// 使用箇所: wbs.go, bottleneck.go, stale.go
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	s := ""
	for n > 0 {
		s = string('0'+byte(n%10)) + s
		n /= 10
	}
	return s
}
