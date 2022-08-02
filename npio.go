package npio

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"syscall"
	"unsafe"

	"github.com/juicechu/go-npio/driver"
)

type Mode uint8
type Pin uint8
type State uint8
type Pull uint8

// Memory offsets for gpio, see the spec for more details
const (
	SUNXI_GPIO_BASE = 0x01C20800
	GPIO_BASE_BP    = 0x01C20000
	CLOCK_BASE_BP   = 0x00101000
	GPIO_PWM_BP     = 0x01C21000

	BLOCK_SIZE = 6 * 1024
	MAP_SIZE   = 4096 * 2
	MAP_MASK   = (MAP_SIZE - 1)
)

// Pin mode, a pin can be set in Input or Output, Clock or Pwm mode
const (
	Input Mode = iota
	Output
	Clock
	Pwm
	Spi
	Alt0
	Alt1
	Alt2
	Alt3
	Alt4
	Alt5
)

// State of pin, High / Low
const (
	Low State = iota
	High
)

// Pull Up / Down / Off
const (
	PullOff Pull = iota
	PullDown
	PullUp
	PullNone
)

// Arrays for 8 / 32 bit access to memory and a semaphore for write locking
var (
	memlock  sync.Mutex
	gpioMem  []uint32
	clkMem   []uint32
	pwmMem   []uint32
	gpioMem8 []uint8
	clkMem8  []uint8
	pwmMem8  []uint8
)

var dri = driver.NewM1()

// Input: Set pin as Input
func (pin Pin) Input() {
	PinMode(pin, Input)
}

// Output: Set pin as Output
func (pin Pin) Output() {
	PinMode(pin, Output)
}

// Clock: Set pin as Clock
func (pin Pin) Clock() {
	PinMode(pin, Clock)
}

// Pwm: Set pin as Pwm
func (pin Pin) Pwm() {
	PinMode(pin, Pwm)
}

// High: Set pin High
func (pin Pin) High() {
	WritePin(pin, High)
}

// Low: Set pin Low
func (pin Pin) Low() {
	WritePin(pin, Low)
}

// Toggle pin state
func (pin Pin) Toggle() {
	TogglePin(pin)
}

// Mode: Set pin Mode
func (pin Pin) Mode(mode Mode) {
	PinMode(pin, mode)
}

// Write: Set pin state (high/low)
func (pin Pin) Write(state State) {
	WritePin(pin, state)
}

// Read pin state (high/low)
func (pin Pin) Read() State {
	return ReadPin(pin)
}

// Pull: Set a given pull up/down mode
func (pin Pin) Pull(pull Pull) {
	PullMode(pin, pull)
}

// PullUp: Pull up pin
func (pin Pin) PullUp() {
	PullMode(pin, PullUp)
}

// PullDown: Pull down pin
func (pin Pin) PullDown() {
	PullMode(pin, PullDown)
}

// PullOff: Disable pullup/down on pin
func (pin Pin) PullOff() {
	PullMode(pin, PullOff)
}

func readl(addr uint32) uint32 {
	var val uint32 = 0
	var mmap_base uint32 = (addr & ^uint32(MAP_MASK))
	var mmap_seek uint32 = ((addr - mmap_base) >> 2)
	val = gpioMem[mmap_seek]
	return val
}

func writel(val, addr uint32) {
	var mmap_base uint32 = (addr & ^uint32(MAP_MASK))
	var mmap_seek uint32 = ((addr - mmap_base) >> 2)
	gpioMem[mmap_seek] = val
}

// PinMode sets the mode of a given pin (Input, Output, Clock, Pwm or Spi)
func PinMode(pin Pin, mode Mode) {
	if pin < 0 || pin > driver.MAX_PIN_COUNT {
		return
	}
	pv := dri.PinToGpio(uint(pin))
	if -1 == pv {
		return
	}

	var (
		regval  uint32 = 0
		phyaddr uint32 = 0
	)
	var (
		bank   int = pv >> 5
		index  int = pv - (bank << 5)
		offset int = ((index - ((index >> 3) << 3)) << 2)
	)
	if driver.BP_PIN_MASK[bank][index] == -1 {
		panic(fmt.Sprintf("BP_PIN_MASK pin(%d) number error\n", pin))
		return
	}
	phyaddr = uint32(SUNXI_GPIO_BASE + (bank * 36) + ((index >> 3) << 2))
	memlock.Lock()
	defer memlock.Unlock()
	regval = readl(phyaddr)
	switch mode {
	case Input:
		regval &= ^(7 << offset)
		writel(regval, phyaddr)
	case Output:
		regval &= ^(7 << offset)
		regval |= (1 << offset)
		writel(regval, phyaddr)
	case Clock:
		//todo
	case Pwm:
		//todo
	}
}

// WritePin sets a given pin High or Low
// by setting the clear or set registers respectively
func WritePin(pin Pin, state State) {
	if pin < 0 || pin > driver.MAX_PIN_COUNT {
		return
	}
	pv := dri.PinToGpio(uint(pin))
	if -1 == pv {
		return
	}
	var (
		regval  uint32 = 0
		phyaddr uint32 = 0
	)
	var (
		bank  int = pv >> 5
		index int = pv - (bank << 5)
	)
	if driver.BP_PIN_MASK[bank][index] == -1 {
		panic(fmt.Sprintf("BP_PIN_MASK pin(%d) number error\n", pin))
		return
	}
	phyaddr = uint32(SUNXI_GPIO_BASE + (bank * 36) + 0x10) // +0x10 -> data reg
	//fmt.Printf("pin=%d pv=%02d bank=%d index=%d phyaddr=%X state=%d\n", pin, pv, bank, index, phyaddr, state)
	memlock.Lock()
	regval = readl(phyaddr)
	if state == Low {
		regval &= ^(1 << index)
		writel(regval, phyaddr)
	} else {
		regval |= (1 << index)
		writel(regval, phyaddr)
	}
	memlock.Unlock() // not deferring saves ~600ns
}

// ReadPin reads the state of a pin
func ReadPin(pin Pin) State {
	if pin < 0 || pin > driver.MAX_PIN_COUNT {
		return Low
	}
	pv := dri.PinToGpio(uint(pin))
	if -1 == pv {
		return Low
	}
	var (
		regval  uint32 = 0
		phyaddr uint32 = 0
	)
	var (
		bank  int = pv >> 5
		index int = pv - (bank << 5)
	)
	if driver.BP_PIN_MASK[bank][index] == -1 {
		panic(fmt.Sprintf("BP_PIN_MASK pin(%d) number error\n", pin))
		return Low
	}
	phyaddr = uint32(SUNXI_GPIO_BASE + (bank * 36) + 0x10) // +0x10 -> data reg
	regval = readl(phyaddr)
	regval = regval >> index
	regval &= 1
	if regval > 0 {
		return High
	}
	return Low
}

// TogglePin: Toggle a pin state (high -> low -> high)
func TogglePin(pin Pin) {
	state := ReadPin(pin)
	switch state {
	case Low:
		state = High
	case High:
		state = Low
	default:
		return
	}
	WritePin(pin, state)
}

func PullMode(pin Pin, pull Pull) {
	if pin < 0 || pin > driver.MAX_PIN_COUNT {
		return
	}
	pv := dri.PinToGpio(uint(pin))
	if -1 == pv {
		return
	}
	var (
		regval  uint32 = 0
		phyaddr uint32 = 0
	)
	var (
		bank      int = pv >> 5
		index     int = pv - (bank << 5)
		sub       int = index >> 4
		sub_index int = index - 16*sub
	)
	if driver.BP_PIN_MASK[bank][index] == -1 {
		panic(fmt.Sprintf("BP_PIN_MASK pin(%d) number error\n", pin))
		return
	}
	var pud uint32
	switch pull {
	case PullOff:
		pud = 7
	case PullUp:
		pud = 7
	case PullDown:
		pud = 5
	}
	pud &= 3
	phyaddr = uint32(SUNXI_GPIO_BASE + (bank * 36) + 0x1c + sub*4) // +0x10 -> pullUpDn reg
	regval = readl(phyaddr)
	regval &= ^(3 << (sub_index << 1))
	regval |= (pud << (sub_index << 1))
	writel(regval, phyaddr)
}

// Open and memory map GPIO memory range from /dev/mem .
// Some reflection magic is used to convert it to a unsafe []uint32 pointer
func Open() (err error) {
	var file *os.File

	// Open fd for rw mem access; try dev/mem first (need root)
	file, err = os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, os.ModePerm)
	if err != nil {
		return
	}
	// FD can be closed after memory mapping
	defer file.Close()

	memlock.Lock()
	defer memlock.Unlock()

	// Memory map GPIO registers to slice
	gpioMem, gpioMem8, err = memMap(file.Fd(), BLOCK_SIZE*10, GPIO_BASE_BP)
	if err != nil {
		return
	}

	// Memory map clock registers to slice
	clkMem, clkMem8, err = memMap(file.Fd(), BLOCK_SIZE, CLOCK_BASE_BP)
	if err != nil {
		return
	}

	// Memory map pwm registers to slice
	pwmMem, pwmMem8, err = memMap(file.Fd(), BLOCK_SIZE, GPIO_PWM_BP)
	if err != nil {
		return
	}

	return nil
}

func memMap(fd uintptr, length int, base int64) (mem []uint32, mem8 []byte, err error) {
	mem8, err = syscall.Mmap(
		int(fd),
		base,
		length,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return
	}
	// Convert mapped byte memory to unsafe []uint32 pointer, adjust length as needed
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&mem8))
	header.Len /= (32 / 8) // (32 bit = 4 bytes)
	header.Cap /= (32 / 8)
	mem = *(*[]uint32)(unsafe.Pointer(&header))
	return
}

// Close unmaps GPIO memory
func Close() error {
	memlock.Lock()
	defer memlock.Unlock()
	for _, mem8 := range [][]uint8{gpioMem8, clkMem8, pwmMem8} {
		if err := syscall.Munmap(mem8); err != nil {
			return err
		}
	}
	return nil
}
