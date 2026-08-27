package service

import (
	"runtime"
	"testing"
)

func TestCompareAssets(t *testing.T) {
	cases := []struct {
		name string
		a, b parsedAsset
		want bool // a 是否比 b 新
	}{
		{
			name: "高版本号更新",
			a:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: stableBetaRank},
			b:    parsedAsset{version: [3]int{0, 0, 1}, betaNum: stableBetaRank},
			want: true,
		},
		{
			name: "相同版本 beta 编号大的更新",
			a:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: 3},
			b:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: 1},
			want: true,
		},
		{
			name: "相同版本正式版高于任何 beta",
			a:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: stableBetaRank},
			b:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: 9},
			want: true,
		},
		{
			name: "高版本 beta 高于低版本正式版",
			a:    parsedAsset{version: [3]int{0, 0, 3}, betaNum: 1},
			b:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: stableBetaRank},
			want: true,
		},
		{
			name: "低版本更新",
			a:    parsedAsset{version: [3]int{0, 0, 1}, betaNum: stableBetaRank},
			b:    parsedAsset{version: [3]int{0, 0, 2}, betaNum: stableBetaRank},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compareAssets(&c.a, &c.b); got != c.want {
				t.Errorf("compareAssets(%+v, %+v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestParseAssetName(t *testing.T) {
	osArch := runtime.GOOS + "-" + runtime.GOARCH

	cases := []struct {
		name       string
		asset      string
		wantParsed bool
		wantBeta   int
	}{
		{"正式版", "MCServer-v0.0.2-" + osArch + ".zip", true, stableBetaRank},
		{"预览版", "MCServer-v0.0.3-beta2-" + osArch + ".zip", true, 2},
		{"不匹配平台", "MCServer-v0.0.2-windows-arm64.zip", false, 0},
		{"非压缩包", "MCServer-v0.0.2-" + osArch + ".exe", false, 0},
		{"不含版本号", "MCServer-" + osArch + ".zip", false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parseAssetName(c.asset, "http://example.com/" + c.asset)
			if !c.wantParsed {
				if p != nil {
					t.Errorf("parseAssetName(%q) = %+v, want nil", c.asset, p)
				}
				return
			}
			if p == nil {
				t.Fatalf("parseAssetName(%q) = nil, want parsed", c.asset)
			}
			if p.betaNum != c.wantBeta {
				t.Errorf("parseAssetName(%q).betaNum = %d, want %d", c.asset, p.betaNum, c.wantBeta)
			}
		})
	}
}
