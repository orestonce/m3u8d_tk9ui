//go:build !windows || !idm

package main

func IDMConfig(a *TKApp) bool {
	return false
}
