// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ledgpio

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/cmd/w83795"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "ledgpio"

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

const (
	maxFanTrays = 4
	maxPsu      = 2
)

var (
	lastFanStatus [maxFanTrays]string
	lastPsuStatus [maxPsu]string

	psuLed       = []uint8{0x8, 0x10}
	psuLedYellow = []uint8{0x8, 0x10}
	psuLedOff    = []uint8{0x04, 0x01}

	sysLed       byte = 0x1
	sysLedGreen  byte = 0x1
	sysLedYellow byte = 0xc
	sysLedOff    byte = 0x80

	fanLed       byte = 0x6
	fanLedGreen  byte = 0x2
	fanLedYellow byte = 0x6
	fanLedOff    byte = 0x0

	deviceVer     byte
	saveFanSpeed  string
	forceFanSpeed bool

	once sync.Once
	Init = func() {}

	first int

	Vdev I2cDev

	VpageByKey map[string]uint8
)

type cmd struct {
	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t
	var err error
	first = 1

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]float64)
	cmd.lasts = make(map[string]string)
	cmd.lastu = make(map[string]uint16)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}
	//if err = cmd.update(); err != nil {
	//      close(cmd.stop)
	//      return err
	//}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if Vdev.Addr != 0 {
				if err = cmd.update(); err != nil {
					//close(cmd.stop)
					return err
				}
			}
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	close(cmd.stop)
	return nil
}

func (cmd *cmd) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}

	if first == 1 {
		err := Vdev.LedFpInit()
		if err != nil {
			return err
		}
		first = 0
	}
	err := Vdev.LedStatus()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) LedFpInit() error {
	var d byte

	pin, found := gpio.Pins["SYSTEM_LED_RST_L"]
	if found {
		pin.SetValue(true)
	}

	deviceVer, _ = readVer()
	// save initial fan speed
	saveFanSpeed, _ = redis.Hget(redis.DefaultHash, "fan_tray.speed")
	forceFanSpeed = false

	r := getRegs()
	r.Output[0].get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	o := s[1].D[0]

	//on bmc boot up set front panel SYS led to green, FAN led to yellow, let PSU drive PSU LEDs
	d = 0xff ^ (sysLed | fanLed)
	o &= d
	o |= sysLedGreen | fanLedYellow

	r.Output[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	r.Config[0].get(h)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	o = s[1].D[0]
	o |= psuLed[0] | psuLed[1]
	o &= (sysLed | fanLed) ^ 0xff

	r.Config[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) LedFpReinit() error {

	deviceVer, _ = readVer()
	r := getRegs()

	r.Config[0].get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	o := s[1].D[0]
	o |= psuLed[0] | psuLed[1]
	o &= (sysLed | fanLed) ^ 0xff

	r.Config[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) LedStatus() error {
	r := getRegs()
	var o, c uint8
	var d byte

	if deviceVer == 0xff || deviceVer == 0x00 {
		psuLed = []uint8{0x0c, 0x03}
		psuLedYellow = []uint8{0x00, 0x00}
		psuLedOff = []uint8{0x04, 0x01}
		sysLed = 0xc0
		sysLedGreen = 0x0
		sysLedYellow = 0xc
		sysLedOff = 0x80
		fanLed = 0x30
		fanLedGreen = 0x10
		fanLedYellow = 0x20
		fanLedOff = 0x30
	}

	allFanGood := true
	fanStatChange := false
	for j := 0; j < maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j+1))+".status")
		if !strings.Contains(p, "ok") {
			allFanGood = false
		}
		if lastFanStatus[j] != p {
			fanStatChange = true
			//if any fan tray is failed or not installed, set front panel FAN led to yellow
			if strings.Contains(p, "warning") && !strings.Contains(lastFanStatus[j], "not installed") {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("warning: fan tray ", j+1, " failure")
				if !forceFanSpeed {
					w83795.Vdev.SetFanSpeed("high")
					forceFanSpeed = true
				}
			} else if strings.Contains(p, "not installed") {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("warning: fan tray ", j+1, " not installed")
				if !forceFanSpeed {
					w83795.Vdev.SetFanSpeed("high")
					forceFanSpeed = true
				}
			} else if strings.Contains(lastFanStatus[j], "not installed") && (strings.Contains(p, "warning") || strings.Contains(p, "ok")) {
				log.Print("notice: fan tray ", j+1, " installed")
			}
		}
		lastFanStatus[j] = p
	}

	if allFanGood && !forceFanSpeed {
		saveFanSpeed, _ = redis.Hget(redis.DefaultHash, "fan_tray.speed")
	}

	//if any fan tray is failed or not installed, set front panel FAN led to yellow
	if fanStatChange {
		if allFanGood {
			// if all fan trays have "ok" status, set front panel FAN led to green
			allStat := true
			for i := range lastFanStatus {
				if lastFanStatus[i] == "" {
					allStat = false
				}
			}
			if allStat {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedGreen
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("notice: all fan trays up")
				if w83795.Vdev.Addr != 0 {
					fanspeed, _ := w83795.Vdev.GetFanSpeed()
					if fanspeed != saveFanSpeed {
						w83795.Vdev.SetFanSpeed(saveFanSpeed)
					}
				}
				forceFanSpeed = false
			}
		}

	}

	for j := 0; j < maxPsu; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "psu"+strconv.Itoa(j+1)+".status")
		if lastPsuStatus[j] != p {
			r.Output[0].get(h)
			r.Config[0].get(h)
			closeMux(h)
			err := DoI2cRpc()
			if err != nil {
				return err
			}
			o = s[1].D[0]
			c = s[3].D[0]
			//if PSU is not installed or installed and powered on, set front panel PSU led to off or green (PSU drives)
			if strings.Contains(p, "not_installed") || strings.Contains(p, "powered_on") {
				c |= psuLed[j]
			} else if strings.Contains(p, "powered_off") {
				//if PSU is installed but powered off, set front panel PSU led to yellow
				d = 0xff ^ psuLed[j]
				o &= d
				o |= psuLedYellow[j]
				c &= (psuLed[j]) ^ 0xff
			}
			r.Output[0].set(h, o)
			r.Config[0].set(h, c)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			lastPsuStatus[j] = p
			if p != "" {
				log.Print("notice: psu", j+1, " ", p)
			}
		}
	}
	return nil
}

func readVer() (v byte, err error) {
	f, err := os.Open("/tmp/ver")
	if err != nil {
		return 0, err
	}
	b1 := make([]byte, 5)
	_, err = f.Read(b1)
	if err != nil {
		return 0, err
	}
	f.Close()
	return b1[0], nil
}