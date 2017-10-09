// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"github.com/platinasystems/go/elib"
)

// SFP Ids from eeprom.
type SfpId uint8

const (
	IdUnknown SfpId = iota
	IdGbic
	IdOnMotherboard
	IdSfp
	IdXbi
	IdXenpak
	IdXfp
	IdXff
	IdXfpE
	IdXpak
	IdX2
	IdDwdmSfp
	IdQsfp
	IdQsfpPlus
	IdCxp
	IdShieldedMiniMultilaneHD4X
	IdShieldedMiniMultilaneHD8X
	IdQsfp28
	IdCxp2
	IdCdfpStyle12
	IdShieldedMiniMultilaneHD4XFanout
	IdShieldedMiniMultilaneHD8XFanoutCable
	IdCdfpStyle3
	IdMicroQsfp
	IdQsfpDD
	// - 0x7F: Reserved
	// 0x80 - 0xFF: Vendor Specific
)

var sfpIdNames = [...]string{
	0x00: "Unknown or unspecified",
	0x01: "GBIC",
	0x02: "Module/connector soldered to motherboard",
	0x03: "SFP/SFP+/SFP28",
	0x04: "300 pin XBI",
	0x05: "XENPAK",
	0x06: "XFP",
	0x07: "XFF",
	0x08: "XFP-E",
	0x09: "XPAK",
	0x0A: "X2",
	0x0B: "DWDM-SFP/SFP+",
	0x0C: "QSFP",
	0x0D: "QSFP+",
	0x0E: "CXP",
	0x0F: "Shielded Mini Multilane HD 4X",
	0x10: "Shielded Mini Multilane HD 8X",
	0x11: "QSFP28",
	0x12: "CXP2/CXP28",
	0x13: "CDFP (Style 1/Style2)",
	0x14: "Shielded Mini Multilane HD 4X Fanout Cable",
	0x15: "Shielded Mini Multilane HD 8X Fanout Cable",
	0x16: "CDFP (Style 3)",
	0x17: "Micro QSFP",
	0x18: "QSFP-DD",
}

func (i SfpId) String() string { return elib.Stringer(sfpIdNames[:], int(i)) }

type SfpConnectorType uint8

const (
	SfpConnectorUnknown SfpConnectorType = iota
	SfpConnectorSubscriber
	SfpConnectorFibreChannelStyle1
	SfpConnectorFibreChannelStyle2
	SfpConnectorBNCTNC
	SfpConnectorFibreChannelCoax
	SfpConnectorFiberJack
	SfpConnectorLucent
	SfpConnectorMTRJ
	SfpConnectorMU
	SfpConnectorSG
	SfpConnectorOpticalPigtail
	SfpConnectorMPO1x12
	SfpConnectorMPO2x16
	SfpConnectorHSSDC2               SfpConnectorType = 0x20
	SfpConnectorCopperPigtail        SfpConnectorType = 0x21
	SfpConnectorRJ45                 SfpConnectorType = 0x22
	SfpConnectorNoSeparableConnector SfpConnectorType = 0x23
	SfpConnectorMXC2x16              SfpConnectorType = 0x24
)

var sfpConnectorTypeNames = [...]string{
	0x00: "Unknown or unspecified",
	0x01: "SC (Subscriber Connector)",
	0x02: "Fibre Channel Style 1 copper connector",
	0x03: "Fibre Channel Style 2 copper connector",
	0x04: "BNC/TNC (Bayonet/Threaded Neill-Concelman)",
	0x05: "Fibre Channel coax headers",
	0x06: "Fiber Jack",
	0x07: "LC (Lucent Connector)",
	0x08: "MT-RJ (Mechanical Transfer - Registered Jack)",
	0x09: "MU (Multiple Optical)",
	0x0A: "SG",
	0x0B: "Optical Pigtail",
	0x0C: "MPO 1x12 (Multifiber Parallel Optic)",
	0x0D: "MPO 2x16",
	0x20: "HSSDC II (High Speed Serial Data Connector)",
	0x21: "Copper pigtail",
	0x22: "RJ45 (Registered Jack)",
	0x23: "No separable connector",
	0x24: "MXC 2x16",
}

func (i SfpConnectorType) String() string { return elib.Stringer(sfpConnectorTypeNames[:], int(i)) }

type SfpCompliance byte

const (
	Log2SfpCompliance40GXLPPI, SfpCompliance40GXLPPI = iota, 1 << iota
	Log2SfpCompliance40G_LR, SfpCompliance40G_LR
	Log2SfpCompliance40G_SR, SfpCompliance40G_SR
	Log2SfpCompliance40G_CR, SfpCompliance40G_CR
	Log2SfpCompliance10G_SR, SfpCompliance10G_SR
	Log2SfpCompliance10G_LR, SfpCompliance10G_LR
	Log2SfpCompliance10G_LRM, SfpCompliance10G_LRM
	Log2SfpComplianceExtendedValid, SfpComplianceExtendedValid
)

func (i SfpCompliance) String() string {
	var t = [...]string{
		Log2SfpCompliance40GXLPPI:      "40G XLPPI",
		Log2SfpCompliance40G_LR:        "40G LR",
		Log2SfpCompliance40G_SR:        "40G SR",
		Log2SfpCompliance40G_CR:        "40G CR",
		Log2SfpCompliance10G_SR:        "10G SR",
		Log2SfpCompliance10G_LR:        "10G LR",
		Log2SfpCompliance10G_LRM:       "10G LRM",
		Log2SfpComplianceExtendedValid: "extended",
	}
	return elib.FlagStringer(t[:], elib.Word(i))
}

type SfpExtendedCompliance byte

const (
	SfpExtendedComplianceUnspecified       = iota
	SfpExtendedCompliance100G_AOC_BER_5e5  // 01h 100G_AOC (Active Optical Cable) or 25GAUI C2M AOC. Providing a worst BER of 5 × 10^(-5)
	SfpExtendedCompliance100G_SR           // 02h 100GBASE-SR4 or 25GBASE-SR
	SfpExtendedCompliance100G_LR           // 03h 100GBASE-LR4 or 25GBASE-LR
	SfpExtendedCompliance100G_ER           // 04h 100GBASE-ER4 or 25GBASE-ER
	SfpExtendedCompliance100G_SR10         // 05h 100GBASE-SR10
	SfpExtendedCompliance100G_CWDM4        // 06h 100G CWDM4
	SfpExtendedCompliance100G_PSM4         // 07h 100G PSM4 Parallel SMF
	SfpExtendedCompliance100G_ACC_BER_5e5  // 08h 100G ACC (Active Copper Cable) or 25GAUI C2M ACC. Providing a worst BER of 5 × 10^(-5)
	_                                      // 09h Obsolete (assigned before 100G CWDM4 MSA required FEC)
	_                                      // 0Ah Reserved
	SfpExtendedCompliance100G_CR           // 0Bh 100GBASE-CR4 or 25GBASE-CR CA-L
	SfpExtendedCompliance25G_CR_CA_S       // 0Ch 25GBASE-CR CA-S
	SfpExtendedCompliance25G_CR_CA_N       // 0Dh 25GBASE-CR CA-N
	_                                      // 0Eh Reserved
	_                                      // 0Fh Reserved
	SfpExtendedCompliance40G_ER            // 10h 40GBASE-ER4
	SfpExtendedCompliance4x10G_SR          // 11h 4 x 10GBASE-SR
	SfpExtendedCompliance40G_PSM4          // 12h 40G PSM4 Parallel SMF
	SfpExtendedComplianceG959_1_P1I1_2D1   // 13h G959.1 profile P1I1-2D1 (10709 MBd, 2km, 1310nm SM)
	SfpExtendedComplianceG959_1_P1S1_2D2   // 14h G959.1 profile P1S1-2D2 (10709 MBd, 40km, 1550nm SM)
	SfpExtendedComplianceG959_1_P1L1_2D2   // 15h G959.1 profile P1L1-2D2 (10709 MBd, 80km, 1550nm SM)
	SfpExtendedCompliance10GBASE_T         // 16h 10GBASE-T with SFI electrical interface
	SfpExtendedCompliance100G_CLR4         // 17h 100G CLR4
	SfpExtendedCompliance100G_AOC_BER_1e12 // 18h 100G AOC or 25GAUI C2M AOC. Providing a worst BER of 10^(-12) or below
	SfpExtendedCompliance100G_ACC_BER_1e12 // 19h 100G ACC or 25GAUI C2M ACC. Providing a worst BER of 10^(-12) or below
	SfpExtendedCompliance100GE_DWDM2       // 1Ah 100GE-DWDM2 (DWDM transceiver using 2 wavelengths on a 1550nm DWDM grid with a reach up to 80km)
)

func (i SfpExtendedCompliance) String() string {
	var t = [...]string{
		SfpExtendedComplianceUnspecified:       "unspecified",
		SfpExtendedCompliance100G_AOC_BER_5e5:  "100G AOC BER < 5e-5",
		SfpExtendedCompliance100G_SR:           "100GBASE-SR4",
		SfpExtendedCompliance100G_LR:           "100GBASE-LR4",
		SfpExtendedCompliance100G_ER:           "100GBASE-ER4",
		SfpExtendedCompliance100G_SR10:         "100GBASE-SR10",
		SfpExtendedCompliance100G_CWDM4:        "100G CWDM4",
		SfpExtendedCompliance100G_PSM4:         "100G PSM4 Parallel SMF",
		SfpExtendedCompliance100G_ACC_BER_5e5:  "100G ACC BER < 5e-5",
		SfpExtendedCompliance100G_CR:           "100GBASE-CR4 or 25GBASE-CR CA-L",
		SfpExtendedCompliance25G_CR_CA_S:       "25GBASE-CR CA-S",
		SfpExtendedCompliance25G_CR_CA_N:       "25GBASE-CR CA-N",
		SfpExtendedCompliance40G_ER:            "40GBASE-ER4",
		SfpExtendedCompliance4x10G_SR:          "4 x 10GBASE-SR",
		SfpExtendedCompliance40G_PSM4:          "40G PSM4",
		SfpExtendedComplianceG959_1_P1I1_2D1:   "G959.1 profile P1I1-2D1 (10709 MBd, 2km, 1310nm SM)",
		SfpExtendedComplianceG959_1_P1S1_2D2:   "G959.1 profile P1S1-2D2 (10709 MBd, 40km, 1550nm SM)",
		SfpExtendedComplianceG959_1_P1L1_2D2:   "G959.1 profile P1L1-2D2 (10709 MBd, 80km, 1550nm SM)",
		SfpExtendedCompliance10GBASE_T:         "10GBASE-T with SFI electrical interface",
		SfpExtendedCompliance100G_CLR4:         "100G CLR4",
		SfpExtendedCompliance100G_AOC_BER_1e12: "100G AOC BER < 1e-12",
		SfpExtendedCompliance100G_ACC_BER_1e12: "100G ACC BER < 1e-12",
		SfpExtendedCompliance100GE_DWDM2:       "100GE-DWDM2",
	}
	return elib.Stringer(t[:], int(i))
}
