//go:build !darwin

package folderdialog

// FinderSelection 非 macOS：返回空（不报错，避免前端误提示「系统不支持」）
func FinderSelection() (string, error) {
	return "", nil
}
