// This file is not a real test.
package main

import (
	"bytes"
	_ "embed"
	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
	"os"
	"testing"
)

//go:embed favicon.ico
var qs []byte

func rsrc() {
	//版本信息
	rs := winres.ResourceSet{}
	icon, _ := winres.LoadICO(bytes.NewReader(qs))
	rs.SetIcon(winres.Name("APPICON"), icon)
	vi := version.Info{
		FileVersion:    [4]uint16{1, 0, 0, 1},
		ProductVersion: [4]uint16{1, 0, 0, 1},
	}
	var lid uint16
	lid = 0x804                                               //繁中0x404 英文 0x409
	_ = vi.Set(lid, version.ProductName, "m3u8d for windows") //产品名称
	_ = vi.Set(lid, version.OriginalFilename, "m3u8d.exe")
	rs.SetVersionInfo(vi)
	rs.SetManifest(winres.AppManifest{
		ExecutionLevel:      2,
		DPIAwareness:        3,
		UseCommonControlsV6: true,
	})
	out, _ := os.Create("rsrc_amd64.syso")
	defer out.Close()
	rs.WriteObject(out, winres.ArchAMD64)
}

func TestRSRC(t *testing.T) {
	rsrc()
}
