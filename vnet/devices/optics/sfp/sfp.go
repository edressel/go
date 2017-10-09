// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"strings"
	"time"
	"unsafe"
)

type QsfpThreshold struct {
	Alarm, Warning struct{ Hi, Lo float64 }
}

type QsfpModuleConfig struct {
	TemperatureInCelsius QsfpThreshold
	SupplyVoltageInVolts QsfpThreshold
	RxPowerInWatts       QsfpThreshold
	TxBiasCurrentInAmps  QsfpThreshold
}

type Access interface {
	// Activate/deactivate module reset.
	SfpReset(active bool)
	// Enable/disable low power mode.
	SfpSetLowPowerMode(enable bool)
	SfpReadWrite(offset uint, p []uint8, isWrite bool)
}

type state struct {
	txDisable uint8
}

type QsfpModule struct {
	// Read in when module is inserted and taken out of reset.
	Eeprom

	signalValues [QsfpNSignal]bool

	Config QsfpModuleConfig

	s state
	a Access
}

func getQsfpRegs() *qsfpRegs { return (*qsfpRegs)(hw.BasePointer) }

func (r *reg8) offset() uint  { return uint(uintptr(unsafe.Pointer(r)) - hw.BaseAddress) }
func (r *reg16) offset() uint { return uint(uintptr(unsafe.Pointer(r)) - hw.BaseAddress) }

func (r *reg8) get(m *QsfpModule) uint8 {
	var b [1]uint8
	m.a.SfpReadWrite(r.offset(), b[:], false)
	return b[0]
}

func (r *reg8) set(m *QsfpModule, v uint8) {
	var b [1]uint8
	b[0] = v
	m.a.SfpReadWrite(r.offset(), b[:], true)
}

func (r *reg16) get(m *QsfpModule) (v uint16) {
	var b [2]uint8
	m.a.SfpReadWrite(r.offset(), b[:], false)
	return uint16(b[0])<<8 | uint16(b[1])
}

func (r *reg16) set(m *QsfpModule, v uint16) {
	var b [2]uint8
	b[0] = uint8(v >> 8)
	b[1] = uint8(v)
	m.a.SfpReadWrite(r.offset(), b[:], true)
}

func (r *regi16) get(m *QsfpModule) (v int16) { v = int16((*reg16)(r).get(m)); return }
func (r *regi16) set(m *QsfpModule, v int16)  { (*reg16)(r).set(m, uint16(v)) }

func (t *QsfpThreshold) get(m *QsfpModule, r *qsfpThreshold, unit float64) {
	t.Warning.Hi = float64(r.warning.hi.get(m)) * unit
	t.Warning.Lo = float64(r.warning.lo.get(m)) * unit
	t.Alarm.Hi = float64(r.alarm.hi.get(m)) * unit
	t.Alarm.Lo = float64(r.alarm.lo.get(m)) * unit
}

const (
	TemperatureToCelsius = 1 / 256.
	SupplyVoltageToVolts = 100e-6
	RxPowerToWatts       = 1e-7
	TxBiasCurrentToAmps  = 2e-6
)

func (m *QsfpModule) SetSignal(s QsfpSignal, new bool) (old bool) {
	old = m.signalValues[s]
	m.signalValues[s] = new
	if old != new {
		switch s {
		case QsfpModuleIsPresent:
			m.Present(new)
		case QsfpInterruptStatus:
			if new {
				m.Interrupt()
			}
		}
	}
	return
}
func (m *QsfpModule) GetSignal(s QsfpSignal) bool { return m.signalValues[s] }

func (m *QsfpModule) Init(a Access) { m.a = a }

func (m *QsfpModule) Interrupt() {
}

func (m *QsfpModule) Present(is bool) {
	r := getQsfpRegs()

	if !is {
		m.invalidateCache()
		return
	}

	// Wait for module to become ready.
	start := time.Now()
	for r.status.get(m)&(1<<0) != 0 {
		if time.Since(start) >= 100*time.Millisecond {
			panic("ready timeout")
		}
	}

	// Read EEPROM.
	if r.upperMemoryMapPageSelect.get(m) != 0 {
		r.upperMemoryMapPageSelect.set(m, 0)
	}
	p := (*[128 / 2]uint16)(unsafe.Pointer(&m.Eeprom))
	for i := range p {
		p[i] = r.upperMemory[i].get(m)
	}

	// Might as well select page 3 forever.
	r.upperMemoryMapPageSelect.set(m, 3)
	if false {
		tr := (*qsfpThresholdRegs)(unsafe.Pointer(&r.upperMemory[0]))
		m.Config.TemperatureInCelsius.get(m, &tr.temperature, TemperatureToCelsius)
		m.Config.SupplyVoltageInVolts.get(m, &tr.supplyVoltage, SupplyVoltageToVolts)
		m.Config.RxPowerInWatts.get(m, &tr.rxPower, RxPowerToWatts)
		m.Config.TxBiasCurrentInAmps.get(m, &tr.txBiasCurrent, TxBiasCurrentToAmps)
	}
}

func (m *QsfpModule) validateCache() {
}

func (m *QsfpModule) invalidateCache() {
}

func (m *QsfpModule) TxEnable(enableMask, laneMask uint) uint {
	r := getQsfpRegs()
	was := m.s.txDisable
	disableMask := byte(^enableMask)
	is := 0xf & ((was &^ byte(laneMask)) | disableMask)
	if is != was {
		r.txDisable.set(m, is)
		m.s.txDisable = is
	}
	return uint(was)
}

func (r *Eeprom) GetCompliance() (c SfpCompliance, x SfpExtendedCompliance) {
	c = SfpCompliance(r.Compatibility[0])
	x = SfpExtendedComplianceUnspecified
	if c&SfpComplianceExtendedValid != 0 {
		x = SfpExtendedCompliance(r.Options[0])
	}
	return
}

func trim(b []byte) string {
	// Strip trailing nulls.
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

func (r *Eeprom) String() string {
	s := fmt.Sprintf("Id: %v", r.Id)
	s += fmt.Sprintf("\n  Vendor: %s, Part Number %s, Revision 0x%x, Serial %s, Date %s",
		trim(r.VendorName[:]), trim(r.VendorPartNumber[:]), trim(r.VendorRevision[:]),
		trim(r.VendorSerialNumber[:]), trim(r.VendorDateCode[:]))
	s += fmt.Sprintf("\n  Connector Type: %v", r.ConnectorType)

	c, x := r.GetCompliance()
	s += fmt.Sprintf("\n  Compliance: %v", c)
	if x != SfpExtendedComplianceUnspecified {
		s += fmt.Sprintf(" %v", x)
	}
	return s
}

func (m *QsfpModule) String() string {
	return m.Eeprom.String()
}
