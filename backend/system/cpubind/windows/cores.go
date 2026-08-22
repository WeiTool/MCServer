//go:build windows

// Package windows 提供 Windows 平台的 CPU 核心绑定实现
// 负责识别性能核心(P-Core)、查找负载最低的核心、绑定进程
package windows

import (
	"strings"
)

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// 采用四级策略（从高到低准确率），任一步命中即返回：
//  1. 调用 GetLogicalProcessorInformationEx API（通过 PowerShell + C# 互操作）
//     读取每颗逻辑核的 EfficiencyClass：0=P 1=E
//  2. 注册表 EfficientClass（兼容老系统/部分机型）
//  3. 基于 CPU 型号硬编码 + 已知消费级/服务器级 CPU 家族推断
//  4. 最终回退 - 全部逻辑核心
func GetPerformanceCores() ([]int, error) {
	// 方法1: 通过 kernel32 API 读取 EfficiencyClass（最准确，Win10 1607+）
	cores, err := getCoresFromProcessorInfo()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法2: 尝试从注册表读取 EfficientClass
	cores, err = getCoresFromRegistry()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法3: 基于 CPU 型号推断 (最可靠)
	cores, err = getCoresFallback()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法4: 最终回退 - 全部核心
	total := GetTotalCores()
	cores = make([]int, total)
	for i := 0; i < total; i++ {
		cores[i] = i
	}
	return cores, nil
}

// getCoresFromProcessorInfo 通过 PowerShell + 嵌入式 C# 调用
// kernel32!GetLogicalProcessorInformationEx，读取 RelationProcessorCore
// 结构中的 EfficiencyClass 来区分 P/E 核。
//
// EfficiencyClass 语义：数值越大代表性能/效率等级越高（由 Windows 报告），
// 典型 Intel ADL/RPL：EffClass=1 是 P-Core、EffClass=0 是 E-Core。
// AMD/服务器/纯 P 核平台：所有核 EffClass 相同，整体视为全 P。
// 做法：取所有核里出现的最大 EffClass，标记对应逻辑核为 P-Core。
func getCoresFromProcessorInfo() ([]int, error) {
	psCmd := `
	$csSource = @'
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;

public static class CoreDetector {
	private const int RelationProcessorCore = 0;

	// 显式布局：基于实测单组处理器下 SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX
	// 大小=48 字节（Relationship=ProcessorCore, GroupCount=1）。
	//   [0..3]   Relationship (int)
	//   [4..7]   Size (uint)
	//   [8]      Flags (byte)
	//   [9]      EfficiencyClass (byte)  — 数值越大代表越"高性能"
	//   [10..29] Reserved (20 字节)
	//   [30..31] GroupCount (ushort)
	//   [32..39] GroupAffinity.Mask (KAFFINITY, ulong)
	//   [40..47] GroupAffinity.Group + 保留
	[StructLayout(LayoutKind.Explicit, Size = 48)]
	private struct ENTRY {
		[FieldOffset(0)]  public int Relationship;
		[FieldOffset(4)]  public uint Size;
		[FieldOffset(9)]  public byte EfficiencyClass;
		[FieldOffset(30)] public ushort GroupCount;
		[FieldOffset(32)] public ulong Mask;
	}

	[DllImport("kernel32.dll", SetLastError = true)]
	private static extern bool GetLogicalProcessorInformationEx(
		int RelationshipType, IntPtr Buffer, ref uint ReturnedLength);

	public static int[] GetPerformanceLogicalProcessors() {
		uint len = 0;
		GetLogicalProcessorInformationEx(RelationProcessorCore, IntPtr.Zero, ref len);
		if (len == 0) return new int[0];

		IntPtr buf = Marshal.AllocHGlobal((int)len);
		try {
			if (!GetLogicalProcessorInformationEx(RelationProcessorCore, buf, ref len)) {
				return new int[0];
			}

			// 第一步：收集所有物理核条目，记录其 EffClass + 归属的逻辑 CPU
			// Key = 逻辑 CPU 编号，Value = 其所属的 EffClass
			var cpuEffClass = new Dictionary<int, byte>();
			byte maxEffClass = 0;

			long offset = buf.ToInt64();
			long end    = offset + len;
			while (offset < end) {
				IntPtr cur = new IntPtr(offset);
				var info = (ENTRY)Marshal.PtrToStructure(cur, typeof(ENTRY));

				if (info.Relationship == RelationProcessorCore) {
					if (info.EfficiencyClass > maxEffClass) maxEffClass = info.EfficiencyClass;
					ulong mask = info.Mask;
					for (int i = 0; i < 64; i++) {
						if ((mask & (1UL << i)) != 0) cpuEffClass[i] = info.EfficiencyClass;
					}
				}

				if (info.Size == 0) break;
				offset += info.Size;
			}

			// 第二步：选择 EffClass 等于最大等级的逻辑 CPU 作为 P-Core
			var pCores = new List<int>();
			foreach (var kv in cpuEffClass) {
				if (kv.Value == maxEffClass) pCores.Add(kv.Key);
			}
			pCores.Sort();
			return pCores.ToArray();
		} finally {
			Marshal.FreeHGlobal(buf);
		}
	}
}
'@

	Add-Type -TypeDefinition $csSource -Language CSharp -ErrorAction Stop
	[CoreDetector]::GetPerformanceLogicalProcessors() | ConvertTo-Json
`
	return runJSONList(psCmd)
}

// getCoresFromRegistry 通过 PowerShell 读取注册表 EfficientClass 识别 P-Core
// 与 GetLogicalProcessorInformationEx 保持一致：取 EfficientClass 最大值的核为 P-Core
func getCoresFromRegistry() ([]int, error) {
	psCmd := `
		$coresMap = @{}
		$index = 0
		$maxClass = 0
		while ($true) {
			$path = "HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\$index"
			try {
				$class = (Get-ItemProperty -Path $path -Name "EfficientClass" -ErrorAction Stop).EfficientClass
				$coresMap[$index] = [int]$class
				if ([int]$class -gt $maxClass) { $maxClass = [int]$class }
			} catch {
				break
			}
			$index++
		}

		$pCores = @()
		foreach ($key in $coresMap.Keys) {
			if ([int]$coresMap[$key] -eq $maxClass) { $pCores += [int]$key }
		}
		if ($pCores.Count -gt 0) {
			$pCores = ($pCores | Sort-Object)
			$pCores | ConvertTo-Json
		} else {
			ConvertTo-Json @()
		}
	`
	return runJSONList(psCmd)
}

// getCoresFallback 基于已知 CPU 型号推断性能核心
// 当 API/注册表法失败时作为回退方案
func getCoresFallback() ([]int, error) {
	total := GetTotalCores()
	cpuName := getCPUName()

	// ============================================
	// 服务器级 / 无 E 核 CPU → 全部都是 P-Core
	// ============================================

	// Intel Xeon（全系列可扩展 / E-xxx / W-xxx）→ 无 E 核
	if strings.Contains(cpuName, "Xeon") {
		return allCores(total), nil
	}

	// AMD 全系列（Ryzen / ThreadRipper / EPYC / Athlon 等）→ 传统无 E 核
	if strings.Contains(cpuName, "AMD") {
		return allCores(total), nil
	}

	// ============================================
	// Intel 12/13/14 代消费级（Alder/Raptor/Meteor Lake）
	// ============================================

	// Intel 12代 移动端 (Alder Lake)
	if strings.Contains(cpuName, "12th Gen Intel") {
		// i5-12450H: 4 P-Core(8T) + 4 E-Core(4T) = 12T，P=0..7
		if strings.Contains(cpuName, "i5-12450H") {
			return pcoresHT(total, 4), nil
		}
		// i7-1260P / i7-1270P: 4 P-Core(8T) + 8 E-Core(8T) = 16T，P=0..7
		if strings.Contains(cpuName, "i7-1260P") || strings.Contains(cpuName, "i7-1270P") {
			return pcoresHT(total, 4), nil
		}
		// i7-12700H / i7-12800H / i9-12900H: 6 P-Core(12T) + 8 E-Core(8T) = 20T，P=0..11
		if strings.Contains(cpuName, "i7-12700H") || strings.Contains(cpuName, "i7-12800H") || strings.Contains(cpuName, "i9-12900H") {
			return pcoresHT(total, 6), nil
		}
		// i5-12500H / i5-12600H: 4 P-Core(8T) + 8 E-Core(8T) = 16T，P=0..7
		if strings.Contains(cpuName, "i5-12500H") || strings.Contains(cpuName, "i5-12600H") {
			return pcoresHT(total, 4), nil
		}
		// i3-1215U / i3-1210U: 2 P-Core(4T) + 4/6 E-Core = 10T/12T，P=0..3
		if strings.Contains(cpuName, "i3-12") {
			return pcoresHT(total, 2), nil
		}
		// 其他 12代 型号 → 用 P:E ≈ 2:3 的经验比例（P 段占前 2/5）
		return intelPECoresByRatio(total), nil
	}

	// Intel 13代 移动端/桌面端 (Raptor Lake)
	if strings.Contains(cpuName, "13th Gen Intel") {
		// i9-13900K/KS/H/HX: 8P(16T)+16E = 32T
		if strings.Contains(cpuName, "i9-139") {
			return pcoresHT(total, 8), nil
		}
		// i7-13700K/H/HX: 8P(16T)+8/12E
		if strings.Contains(cpuName, "i7-137") {
			return pcoresHT(total, 8), nil
		}
		// i5-13500/H/HX: 6P(12T)+4/8E
		if strings.Contains(cpuName, "i5-135") {
			return pcoresHT(total, 6), nil
		}
		// i5-13400: 6P(12T)+4E
		if strings.Contains(cpuName, "i5-134") {
			return pcoresHT(total, 6), nil
		}
		// i3-1315U 等 2P
		if strings.Contains(cpuName, "i3-13") {
			return pcoresHT(total, 2), nil
		}
		return intelPECoresByRatio(total), nil
	}

	// Intel 14代 移动端/桌面端 (Meteor Lake / Raptor Lake Refresh)
	if strings.Contains(cpuName, "14th Gen Intel") {
		// i9-14900K/KS: 8P(16T) + 16E
		if strings.Contains(cpuName, "i9-149") {
			return pcoresHT(total, 8), nil
		}
		// i7-14700K: 8P + 12E；i7-14700H: 8P + 8E
		if strings.Contains(cpuName, "i7-147") {
			return pcoresHT(total, 8), nil
		}
		// i5-14600K: 6P + 8E；i5-14500/H/HX: 6P + 4/8E
		if strings.Contains(cpuName, "i5-14") {
			return pcoresHT(total, 6), nil
		}
		// i3-14100/U 无 E 核（4 非 SMT P 核）
		if strings.Contains(cpuName, "i3-14") {
			return pcoresHT(total, 4), nil
		}
		return intelPECoresByRatio(total), nil
	}

	// Intel 15代 Ultra (Lunar Lake)：2P + 4LP (E) 6T 逻辑核
	if strings.Contains(cpuName, "15th Gen Intel") || strings.Contains(cpuName, "Intel Core Ultra") {
		return pcoresHT(total, 2), nil
	}

	// ============================================
	// 通用兜底推断
	//   - 核心数 > 8 且带 "Core i5/i7/i9" → 按 Intel P/E 比例估
	//   - 否则 → 全部作为 P-Core（老机器 / 服务器 / 其他厂商）
	// ============================================
	if total > 8 && (strings.Contains(cpuName, "Core i5") || strings.Contains(cpuName, "Core i7") || strings.Contains(cpuName, "Core i9")) {
		return intelPECoresByRatio(total), nil
	}
	return allCores(total), nil
}

// pcoresHT 返回前 n 颗物理 P 核展开后的所有逻辑处理器编号
// 假设 P 核在前且启用 SMT：每颗物理 P 核对应 2 颗逻辑 CPU，按序排列
// 例：total=12, n=4 (4P × 2T) → [0..7]
func pcoresHT(total int, n int) []int {
	// 每颗 P 核 × SMT(2T) = 2n 颗逻辑 CPU，取编号 0..2n-1
	want := n * 2
	if want > total {
		want = total
	}
	return allCores(want)
}

// intelPECoresByRatio 针对 Intel 大小核架构估算 P-Core 数量
// 经验比例：P : E 逻辑核 ≈ 2 : 3（即 P 约占 40%，Raptor Lake 典型 8P+8E=48%、10P+16E=40%）
// 返回值向上取整为偶数（SMT 保证成对），最少 2 颗逻辑 CPU
func intelPECoresByRatio(total int) []int {
	if total <= 8 {
		// 核心数少（如 U 系 4+4=10T 以下），视为全 P（P-Core 占用率会高，但避免误杀）
		return allCores(total)
	}
	n := total * 2 / 5
	if n < 2 {
		n = 2
	}
	// 向上取偶数（确保 P 核成对，匹配 SMT 物理核布局）
	if n%2 != 0 {
		n++
	}
	return allCores(n)
}

// allCores 返回从 0 到 count-1 的核心编号列表
func allCores(count int) []int {
	if count <= 0 {
		count = 1
	}
	cores := make([]int, count)
	for i := 0; i < count; i++ {
		cores[i] = i
	}
	return cores
}
