//go:build !darwin

package folderdialog

import "errors"

// FinderSelection 非 macOS 不支持
func FinderSelection() (string, error) {
	return "", errors.New("仅 macOS 支持读取访达选中项")
}
