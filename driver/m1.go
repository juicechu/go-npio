package driver

type M1 struct {
	pintoGpio [MAX_PIN_COUNT]int
}

func (d *M1) PinToGpio(p uint) int {
	return d.pintoGpio[p]
}

func NewM1() Driver {
	// WiringPiNr. gegeben .. -> Array GPIOx orange pi guenter neu
	// A ab 0x00, B ab 0x20, C ab 0x40, D ab 0x50 ......
	// 00 - 31 = PA00-PA31
	// 32 - 63 = PB00-PB31
	// 64 - 95 = PC00-PC31
	// 96 - 127 = PD00-PD31
	// 128 - 159 = PE00-PE31
	// 160 - 191 = PF00-PF31
	// 192 - 223 = PG00-PG31
	// nanopi m1 done
	return &M1{
		pintoGpio: [MAX_PIN_COUNT]int{
			0, 6, //  0,  1
			2, 3, //  2,  3
			200, 201, //  4,  5
			1, 203, //  6,  7
			12, 11, //  8,  9
			67, 17, // 10, 11
			64, 65, // 12, 13
			66, 198, // 14, 15
			199, -1, // 16, 17
			-1, -1, // 18, 19
			-1, 20, // 20, 21
			21, 8, // 22, 23
			13, 9, // 24, 25
			7, 16, // 26, 27
			15, 14, // 28, 29
			19, 18, // 30, 31
			4, 5, // 32, 33 Debug UART pins
			-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, // ... 47
			-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, // ... 63
			/* 64~73 */
			-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		},
	}
}
