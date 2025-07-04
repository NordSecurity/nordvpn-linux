package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMachineID(t *testing.T) {
	category.Set(t, category.Unit)

	const hostname = "host"

	tests := []struct {
		name         string
		hostname     string
		filesContent map[string]string
		expectedId   func() uuid.UUID
		expectsError bool
	}{
		{
			name:         "Fails to generate system ID when hostname is empty",
			expectedId:   func() uuid.UUID { return uuid.UUID{} },
			expectsError: true,
		},
		{
			name:         "Fails for empty files",
			expectedId:   func() uuid.UUID { return uuid.UUID{} },
			hostname:     "host",
			expectsError: true,
		},
		{
			name:     "Successful for hostname + /etc/machine-id",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /var/lib/dbus/machine-id",
			hostname: hostname,
			filesContent: map[string]string{
				"/var/lib/dbus/machine-id": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /sys/class/dmi/id/product_uuid",
			hostname: hostname,
			filesContent: map[string]string{
				"/sys/class/dmi/id/product_uuid": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte(uuid.NameSpaceDNS.String()))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /etc/machine-id + /proc/cpuinfo",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id": uuid.NameSpaceDNS.String(),
				"/proc/cpuinfo":   "Serial: cpuinfo",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte("cpuinfo"))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /etc/machine-id + /sys/class/dmi/id/board_serial",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id":                uuid.NameSpaceDNS.String(),
				"/sys/class/dmi/id/board_serial": "board_serial",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte("board_serial"))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Add files are present",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id":                uuid.NameSpaceDNS.String(),
				"/var/lib/dbus/machine-id":       uuid.NameSpaceDNS.String(),
				"/sys/class/dmi/id/product_uuid": uuid.NameSpaceURL.String(),
				"/sys/class/dmi/id/board_serial": "board_serial",
				"/proc/cpuinfo":                  "Serial: cpuinfo",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				// CPU ID
				id = uuid.NewSHA1(id, []byte("cpuinfo"))
				// device product number, /sys/class/dmi/id/product_uuid
				id = uuid.NewSHA1(id, []byte(uuid.NameSpaceURL.String()))
				// board serial number
				id = uuid.NewSHA1(id, []byte("board_serial"))

				return id
			},
			expectsError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := NewMachineID(
				func(fileName string) ([]byte, error) {
					val, ok := test.filesContent[fileName]
					if !ok {
						return nil, fmt.Errorf("cannot open file")
					}
					return []byte(val), nil
				},
				func() (name string, err error) {
					if test.hostname == "" {
						return "", fmt.Errorf("failed to get hostname")
					}
					return test.hostname, nil
				},
			)

			// test internal device generator which uses the hardware info
			id, err := generator.generateID()

			assert.Equal(t, test.expectsError, err != nil)
			assert.Equal(t, test.expectedId(), id)

			// Test if MachineID UUID contains hostname string
			assert.False(t, strings.Contains(id.String(), hostname))

			// Test if MachineID UUID contains hostname bytes
			byteStringSice := []byte(hostname)
			for index, hexVal := range byteStringSice {
				if !strings.Contains(id.String(), fmt.Sprintf("%x", int(hexVal))) {
					break
				}
				assert.NotEqual(t, index, len(byteStringSice)-1, "Machine ID contains hostname bytes")
			}

			// generate second time to be sure the same result is obtained
			secondId, err := generator.generateID()
			assert.Equal(t, test.expectsError, err != nil)
			assert.Equal(t, test.expectedId(), secondId)

			// check that the public function always returns UUID,
			// even if the application is not able to get system & hardware information
			machineUUID := generator.GetMachineID()
			parsedUUID, err := uuid.Parse(machineUUID.String())
			assert.Nil(t, err)
			assert.Equal(t, machineUUID, parsedUUID)
		})
	}
}

func TestFallbackGenerateUUID(t *testing.T) {
	generator := MachineID{}

	// check that a UUID is returned and that is is not all 0s
	id, err := uuid.Parse(generator.fallbackGenerateUUID().String())
	assert.Nil(t, err)
	assert.NotEqual(t, id, uuid.UUID{})
}

func TestMachineID_GetArchitectureVariantName(t *testing.T) {
	tests := []struct {
		name       string
		archFamily string
		content    string
		want       string
		wantErr    bool
	}{
		{
			name:       "Detects valid ARMv7",
			archFamily: "armhf",
			content: `Linux Everest 3.2.40 #7321 SMP Wed Mar 23 11:47:17 CST 2016 armv7l GNU/Linux synology_armada375_ds115
Processor : ARMv7 Processor rev 1 (v7l)
processor : 0
BogoMIPS : 1594.16

processor : 1
BogoMIPS : 1594.16

Features : swp half thumb fastmult vfp edsp neon vfpv3 tls
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x4
CPU part : 0xc09
CPU revision : 1

Hardware : Marvell Armada-375 Board
Revision : 0000
Serial : 0000000000000000`,
			want:    "armhfv7",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv5TE",
			archFamily: "armel",
			content: `Processor name : Feroceon 88F6281 rev 1 (v5l) @ 1.2 GHz
BogoMIPS : 1196.85
Features : swp half thumb fastmult edsp
CPU implementer : 0x56
CPU architecture: 5TE
CPU variant : 0x2
CPU part : 0x131
CPU revision : 1

Hardware : Feroceon-KW ARM
Revision : 0000
Serial : 0000000000000000`,
			want:    "armelv5te",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv7 (2)",
			archFamily: "armhf",
			content: `Processor : Marvell PJ4Bv7 Processor rev 2 (v7l)
processor : 0
BogoMIPS : 1196.85

processor : 1
BogoMIPS : 1196.85

processor : 2
BogoMIPS : 1196.85

Features : swp half thumb fastmult vfp edsp vfpv3 tls
CPU implementer : 0x56
CPU architecture: 7
CPU variant : 0x2
CPU part : 0x584
CPU revision : 2

Hardware : Marvell Armada XP Development Board
Revision : 0000
Serial : 0000000000000000`,
			want:    "armhfv7",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv8",
			archFamily: "aarch64",
			content: `Processor       : AArch64 Processor rev 4 (aarch64)
processor       : 0
processor       : 1
processor       : 2
processor       : 3
Features        : fp asimd evtstrm aes pmull sha1 sha2 crc32 wp half thumb fastmult vfp edsp neon vfpv3 tlsi vfpv4 idiva idivt
CPU implementer : 0x41
CPU architecture: 8
CPU variant     : 0x0
CPU part        : 0xd03
CPU revision    : 4

Hardware        : Amlogic
Serial          : beca79021141185f
AmLogic Serial  : 210d740041443cd1828b454f6133c990`,
			want:    "aarch64v8",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv7 (3) Multi processors",
			archFamily: "armhf",
			content: `processor : 0
model name : ARMv7 Processor rev 4 (v7l)
BogoMIPS : 38.40
Features : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm crc32
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part : 0xd03
CPU revision : 4

processor : 1
model name : ARMv7 Processor rev 4 (v7l)
BogoMIPS : 38.40
Features : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm crc32
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part : 0xd03
CPU revision : 4

processor : 2
model name : ARMv7 Processor rev 4 (v7l)
BogoMIPS : 38.40
Features : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm crc32
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part : 0xd03
CPU revision : 4

processor : 3
model name : ARMv7 Processor rev 4 (v7l)
BogoMIPS : 38.40
Features : half thumb fastmult vfp edsp neon vfpv3 tls vfpv4 idiva idivt vfpd32 lpae evtstrm crc32
CPU implementer : 0x41
CPU architecture: 7
CPU variant : 0x0
CPU part : 0xd03
CPU revision : 4

Hardware : BCM2709
Revision : a22082
Serial : 000000000db98e4e`,
			want:    "armhfv7",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv6",
			archFamily: "armel",
			content: `Processor       : ARMv6-compatible processor rev 7 (v6l)
BogoMIPS        : 697.95
Features        : half thumb fastmult vfp edsp java tls
CPU implementer : 0x41
CPU architecture: 6
CPU variant     : 0x0
CPU part        : 0xb76
CPU revision    : 7
Hardware        : BCM2708
Revision        : 0002
Serial          : 00000000abcdef01`,
			want:    "armelv6",
			wantErr: false,
		},
		{
			name:       "Detects valid ARMv5",
			archFamily: "armel",
			content: `Processor       : ARM926EJ-S rev 5 (v5l)
BogoMIPS        : 226.06
Features        : swp half thumb fastmult edsp java
CPU implementer : 0x41
CPU architecture: 5
CPU variant     : 0x0
CPU part        : 0x926
CPU revision    : 5
Hardware        : ARM-Board
Revision        : 0000
Serial          : 00000000deadbeef`,
			want:    "armelv5",
			wantErr: false,
		},
		{
			name:       "Unsupported arch - amd64 architecture",
			archFamily: "amd64",
			content: `		processor	: 0
vendor_id	: AuthenticAMD
cpu family	: 25
model		: 116
model name	: AMD Ryzen 7 PRO 7840U w/ Radeon 780M Graphics
stepping	: 1
microcode	: 0xa704107
cpu MHz		: 2162.084
cache size	: 1024 KB
physical id	: 0
siblings	: 16
core id		: 0
cpu cores	: 8
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 16
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ht syscall nx mmxext fxsr_opt pdpe1gb rdtscp lm constant_tsc rep_good amd_lbr_v2 nopl xtopology nonstop_tsc cpuid extd_apicid aperfmperf rapl pni pclmulqdq monitor ssse3 fma cx16 sse4_1 sse4_2 x2apic movbe popcnt aes xsave avx f16c rdrand lahf_lm cmp_legacy svm extapic cr8_legacy abm sse4a misalignsse 3dnowprefetch osvw ibs skinit wdt tce topoext perfctr_core perfctr_nb bpext perfctr_llc mwaitx cpb cat_l3 cdp_l3 hw_pstate ssbd mba perfmon_v2 ibrs ibpb stibp ibrs_enhanced vmmcall fsgsbase bmi1 avx2 smep bmi2 erms invpcid cqm rdt_a avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves cqm_llc cqm_occup_llc cqm_mbm_total cqm_mbm_local user_shstk avx512_bf16 clzero irperf xsaveerptr rdpru wbnoinvd cppc arat npt lbrv svm_lock nrip_save tsc_scale vmcb_clean flushbyasid decodeassists pausefilter pfthreshold vgif x2avic v_spec_ctrl vnmi avx512vbmi umip pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg avx512_vpopcntdq rdpid overflow_recov succor smca fsrm flush_l1d amd_lbr_pmc_freeze
bugs		: sysret_ss_attrs spectre_v1 spectre_v2 spec_store_bypass srso
bogomips	: 6587.11
TLB size	: 3584 4K pages
clflush size	: 64
cache_alignment	: 64
address sizes	: 48 bits physical, 48 bits virtual
power management: ts ttp tm hwpstate cpb eff_freq_ro [13] [14] [15]`,
			want:    "amd64",
			wantErr: false,
		},
		{
			name:       "Unsupported arch - x86 architecture",
			archFamily: "i386",
			content: `processor   : 0
vendor_id   : GenuineIntel
cpu family  : 6
model       : 37
model name  : Intel(R) Core(TM) i3 CPU       M 330  @ 2.13GHz
stepping    : 2
cpu MHz     : 933.000
cache size  : 3072 KB
physical id : 0
siblings    : 4
core id     : 0
cpu cores   : 2
apicid      : 0
initial apicid  : 0
fdiv_bug    : no
hlt_bug     : no
f00f_bug    : no
coma_bug    : no
fpu     : yes
fpu_exception   : yes
cpuid level : 11
wp      : yes
flags       : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe nx rdtscp lm constant_tsc arch_perfmon pebs bts xtopology nonstop_tsc aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm sse4_1 sse4_2 popcnt lahf_lm arat dts tpr_shadow vnmi flexpriority ept vpid
bogomips    : 4256.49
clflush size    : 64
cache_alignment : 64
address sizes   : 36 bits physical, 48 bits virtual
power management:`,
			want:    "i386",
			wantErr: false,
		},
		{
			name:       "Empty content",
			archFamily: "armhf",
			content:    "",
			want:       "",
			wantErr:    true,
		},
		{
			name:       "Malformed content with no architecture line",
			archFamily: "armhf",
			content: `Processor: Unknown CPU
Features: none
Something Else: else else else`,
			want:    "",
			wantErr: true,
		},
		{
			name:       "ARMv7 architecture without keyword 'CPU architecture'",
			archFamily: "armhf",
			content: `Processor : ARMv7 Processor rev 1 (v7l)
processor : 0
BogoMIPS : 1594.16

Hardware : Marvell Armada-375 Board
Revision : 0000
Serial : 0000000000000000`,
			want:    "",
			wantErr: true,
		},
		{
			name:       "ARMv5 with lowercase CPU architecture line",
			archFamily: "armel",
			content: `processor : 0
cpu implementer : 0x41
cpu architecture: 5
cpu variant : 0x0
cpu part : 0x926
cpu revision : 1
Hardware : ARM-Board`,
			want:    "",
			wantErr: true,
		},
		{
			name:       "ARMv8 AArch64 missing architecture line",
			archFamily: "aarch64",
			content: `Processor       : AArch64 Processor rev 4 (aarch64)
Features        : fp asimd evtstrm aes sha1 sha2 crc32`,
			want:    "",
			wantErr: true,
		},
		{
			name:       "Completely unsupported input format",
			archFamily: "armhf",
			content:    "<<<binary data>>> \x00\x01\x02\x03",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := NewMachineID(
				func(fileName string) ([]byte, error) {
					if fileName != "/proc/cpuinfo" {
						return nil, fmt.Errorf("incorrect file used")
					}
					return []byte(tt.content), nil
				},
				func() (name string, err error) { return "", nil },
			)

			got, err := getter.GetArchitectureVariantName(tt.archFamily)
			assert.Equal(t, err != nil, tt.wantErr,
				"MachineID.GetArchitectureVariantName() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, got, tt.want,
				"MachineID.GetArchitectureVariantName() = %v, want %v", got, tt.want)
		})
	}

	getter := NewMachineID(
		func(fileName string) ([]byte, error) {
			if fileName != "/proc/something-other" {
				return nil, fmt.Errorf("incorrect file used")
			}
			return []byte(""), nil
		},
		func() (name string, err error) { return "", nil },
	)

	got, err := getter.GetArchitectureVariantName("aarch64")
	assert.NotEqual(t, err, nil,
		"MachineID.GetArchitectureVariantName() error = %v, wantErr %v", err, true)
	assert.Equal(t, got, "",
		"MachineID.GetArchitectureVariantName() = %v, want %v", got, "")
}

func Test_normalizeArchFamily(t *testing.T) {
	type args struct {
		family string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Exact arm string",
			args: args{family: "arm"},
			want: "arm",
		},
		{
			name: "ARM uppercase",
			args: args{family: "ARM"},
			want: "arm",
		},
		{
			name: "armv7 variant",
			args: args{family: "armv7l"},
			want: "arm",
		},
		{
			name: "aarch64 lowercase",
			args: args{family: "aarch64"},
			want: "arm",
		},
		{
			name: "AARCH64 uppercase",
			args: args{family: "AARCH64"},
			want: "arm",
		},
		{
			name: "Unrelated x86_64",
			args: args{family: "x86_64"},
			want: "x86_64",
		},
		{
			name: "Mixed-case AMD64",
			args: args{family: "AmD64"},
			want: "amd64",
		},
		{
			name: "RISC-V architecture",
			args: args{family: "riscv64"},
			want: "riscv64",
		},
		{
			name: "MIPS architecture",
			args: args{family: "mips"},
			want: "mips",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeArchFamily(tt.args.family)
			assert.Equal(t, got, tt.want, "normalizeArchFamily() = %v, want %v", got, tt.want)
		})
	}
}
