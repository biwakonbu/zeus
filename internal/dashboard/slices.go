package dashboard

// nonNilStrings は nil スライスを空スライスに正規化する。
// JSON へエンコードしたときに null ではなく [] になることを保証する。
func nonNilStrings(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}
