// Copyright 2021 Sumo Logic, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import "go.opentelemetry.io/collector/pdata/pmetric"

// translateTelegrafMetric translates telegraf metrics names to align with
// Sumo conventions expected in Sumo host related apps, for example:
// * `procstat_num_threads` => `Proc_Threads`
// * `cpu_usage_irq` => `CPU_Irq`
func translateTelegrafMetric(m pmetric.Metric) {
	var newName string
	switch m.Name() {

	// CPU metrics
	case "cpu_usage_active":
		newName = "CPU_Total"
	case "cpu_usage_idle":
		newName = "CPU_Idle"
	case "cpu_usage_iowait":
		newName = "CPU_IOWait"
	case "cpu_usage_irq":
		newName = "CPU_Irq"
	case "cpu_usage_nice":
		newName = "CPU_Nice"
	case "cpu_usage_softirq":
		newName = "CPU_SoftIrq"
	case "cpu_usage_steal":
		newName = "CPU_Stolen"
	case "cpu_usage_System":
		newName = "CPU_Sys"
	case "cpu_usage_user":
		newName = "CPU_User"
	case "system_load1":
		newName = "CPU_LoadAvg_1min"
	case "system_load5":
		newName = "CPU_LoadAvg_5min"
	case "system_load15":
		newName = "CPU_LoadAvg_15min"

	// Disk metrics
	case "disk_used":
		newName = "Disk_Used"
	case "disk_used_percent":
		newName = "Disk_UsedPercent"
	case "disk_inodes_free":
		newName = "Disk_InodesAvailable"

	// Disk IO metrics
	case "diskio_reads":
		newName = "Disk_Reads"
	case "diskio_read_bytes":
		newName = "Disk_ReadBytes"
	case "diskio_writes":
		newName = "Disk_Writes"
	case "diskio_write_bytes":
		newName = "Disk_WriteBytes"

	// Memory metrics
	case "mem_total":
		newName = "Mem_Total"
	case "mem_free":
		newName = "Mem_free"
	case "mem_available":
		newName = "Mem_ActualFree"
	case "mem_used":
		newName = "Mem_ActualUsed"
	case "mem_used_percent":
		newName = "Mem_UsedPercent"
	case "mem_available_percent":
		newName = "Mem_FreePercent"

	// Procstat metrics
	case "procstat_num_threads":
		newName = "Proc_Threads"
	case "procstat_memory_vms":
		newName = "Proc_VMSize"
	case "procstat_memory_rss":
		newName = "Proc_RSSize"
	case "procstat_cpu_usage":
		newName = "Proc_CPU"
	case "procstat_major_faults":
		newName = "Proc_MajorFaults"
	case "procstat_minor_faults":
		newName = "Proc_MinorFaults"

	// Net metrics
	case "net_bytes_sent":
		newName = "Net_OutBytes"
	case "net_bytes_recv":
		newName = "Net_InBytes"
	case "net_packets_sent":
		newName = "Net_OutPackets"
	case "net_packets_recv":
		newName = "Net_InPackets"

	// Netstat metrics
	case "netstat_tcp_close":
		newName = "TCP_Close"
	case "netstat_tcp_close_wait":
		newName = "TCP_CloseWait"
	case "netstat_tcp_closing":
		newName = "TCP_Closing"
	case "netstat_tcp_established":
		newName = "TCP_Established"
	case "netstat_tcp_listen":
		newName = "TCP_Listen"
	case "netstat_tcp_time_wait":
		newName = "TCP_TimeWait"

	default:
		return
	}

	m.SetName(newName)
}
