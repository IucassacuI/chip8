package main

import (
	"fmt"
	"github.com/mattn/go-tty"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

var memory []uint8 = make([]uint8, 4096)

var registers []uint8 = make([]uint8, 16)

var index uint16

var pc uint16 = 0x200

var screen [32][64]uint8

var delay_timer uint8
var sound_timer uint8

var stack []uint16 = make([]uint16, 16)

var fontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0,
	0x20, 0x60, 0x20, 0x20, 0x70,
	0xF0, 0x10, 0xF0, 0x80, 0xF0,
	0xF0, 0x10, 0xF0, 0x10, 0xF0,
	0x90, 0x90, 0xF0, 0x10, 0x10,
	0xF0, 0x80, 0xF0, 0x10, 0xF0,
	0xF0, 0x80, 0xF0, 0x90, 0xF0,
	0xF0, 0x10, 0x20, 0x40, 0x40,
	0xF0, 0x90, 0xF0, 0x90, 0xF0,
	0xF0, 0x90, 0xF0, 0x10, 0xF0,
	0xF0, 0x90, 0xF0, 0x90, 0x90,
	0xE0, 0x90, 0xE0, 0x90, 0xE0,
	0xF0, 0x80, 0x80, 0x80, 0xF0,
	0xE0, 0x90, 0x90, 0x90, 0xE0,
	0xF0, 0x80, 0xF0, 0x80, 0xF0,
	0xF0, 0x80, 0xF0, 0x80, 0x80,
}

var keypad = map[uint8]uint8{
	0: 0,
	'1': 0x1,
	'2': 0x2,
	'3': 0x3,
	'4': 0xC,
	'q': 0x4,
	'w': 0x5,
	'e': 0x6,
	'r': 0xD,
	'a': 0x7,
	's': 0x8,
	'd': 0x9,
	'f': 0xE,
	'z': 0xA,
	'x': 0x0,
	'c': 0xB,
	'v': 0xF,
}

func clearscreen() {
	for ind := 0; ind < 32; ind++ {
		for i := 0; i < 64; i++ {
			screen[ind][i] = 0
		}
	}
}

func update() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()

	for _, y := range screen {
		for _, x := range y {
			if x == 0 {
				fmt.Print(" ")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Print("\n")

	}
}

func loadfont() {
	for index, byt := range fontSet {
		memory[index] = byt
	}
}

func load_rom(game string) {
	rom, err := os.ReadFile(game)

	if err != nil {
		panic("ROM não existente")
	}

	if len(rom) > len(memory)-0x200 {
		panic("ROM maior que a memória")
	}

	for i, byt := range rom {
		memory[0x200+i] = byt
	}
}

func top() uint16 {
	return stack[len(stack)-1]
}

func pop() {
	stack = stack[:len(stack)-1]
}

func reg_dump(vx uint8, i uint16) {
	for offset := 0; uint8(offset) < vx+1; offset++ {
		memory[index+uint16(offset)] = registers[offset]
	}
}

func reg_load(vx uint8, i uint16) {
	for offset := 0; uint8(offset) < vx+1; offset++ {
		registers[offset] = memory[index+uint16(offset)]
	}
}

func getkey() uint8 {
	tty, err := tty.Open()

	if err != nil {
		log.Fatal(err)
	}

	defer tty.Close()

	key, err := tty.ReadRune()

	if err != nil {
		log.Fatal(err)
	}

	return keypad[uint8(key)]
}

func draw(vx, vy, n uint8) {
		registers[0xF] = 0
		
		var y uint16 = 0
		var x uint16 = 0
		for y = 0; y < uint16(n); y++ {
			pixel := memory[index+y]
			for x = 0; x < 8; x++ {
				if (pixel & (0x80 >> x)) != 0 {
					if screen[(vy + uint8(y))%32][(vx+uint8(x))%64] == 1 {
						registers[0xF] = 1
					}
					screen[(vy  + uint8(y))%32][(vx+uint8(x))%64] ^= 1
				}
			}
		}
	update()
}

func run() {

	var opcode uint16
	var nnn uint16
	var nn uint8
	var n uint8
	var x uint8
	var y uint8
	var keyp uint8
	var next bool = true

	go func()uint8{ for{ keyp = getkey() } }()

	for pc != 0xFFE {
		opcode = uint16(memory[pc])<<8 | uint16(memory[pc+1])

		nnn = opcode & 0x0FFF
		nn = uint8(opcode & 0x00FF)
		n = uint8(opcode & 0x000F)
		x = uint8((opcode & 0x0F00) >> 8)
		y = uint8((opcode & 0x00F0) >> 4)

		switch opcode & 0xF000 {
		case 0:
			if nn == 0xE0 {
				clearscreen()
			} else if nn == 0xEE {
				pc = top()
				pop()
			} else {
				panic("Operação ilegal.")
			}
		case 0x1000:
			pc = nnn
			next = false
		case 0x2000:
			stack = append(stack, pc)
			pc = nnn
			next = false
		case 0x3000:
			if registers[x] == nn {
				pc += 2
			}
		case 0x4000:
			if registers[x] != nn {
				pc += 2
			}
		case 0x5000:
			if registers[x] == registers[y] {
				pc += 2
			}
		case 0x6000:
			registers[x] = nn
		case 0x7000:
			registers[x] += nn
		case 0x8000:
			switch n {
			case 0:
				registers[x] = registers[y]
			case 1:
				registers[x] |= registers[y]
			case 2:
				registers[x] &= registers[y]
			case 3:
				registers[x] ^= registers[y]
			case 4:
				if registers[y] > 0xFF-registers[y] {
					registers[15] = 1
				} else {
					registers[15] = 0
				}
				registers[x] += registers[y]
			case 5:
				if registers[y] > registers[x] {
					registers[15] = 0
				} else {
					registers[15] = 1
				}
				registers[x] -= registers[y]
			case 6:
				registers[15] = registers[x] & 1
				registers[x] >>= 1
			case 7:
				if registers[x] > registers[y] {
					registers[15] = 0
				} else {
					registers[15] = 1
				}
				registers[x] = registers[y] - registers[x]
			case 0xE:
				registers[15] = registers[x] >> 7

				registers[x] <<= 1
			default:
				panic("Operação Ilegal.")
			}
		case 0x9000:
			if registers[x] != registers[y] {
				pc += 2
			}
		case 0xA000:
			index = nnn
		case 0xB000:
			pc = uint16(registers[0]) + nnn
			next = false
		case 0xC000:
			registers[x] = uint8(rand.Intn(255)) & nn
		case 0xD000:
			draw(registers[x], registers[y], n)
		case 0xE000:
			switch n {
			case 0xE:
					if keyp == registers[x] {
						pc += 2
					}
			case 1:
					if keyp != registers[x] {
						pc += 2
					}
			default:
				panic("Operação ilegal.")
			}
		case 0xF000:
			switch nn {
			case 7:
				registers[x] = delay_timer
			case 0xA:
					for {
						if keyp != 0 {
							break
						}
					}
					registers[x] = keyp
			case 0x15:
				delay_timer = registers[x]
			case 0x18:
				sound_timer = registers[x]
			case 0x1E:
				index += uint16(registers[x])
			case 0x29:
				index = uint16(registers[x] * 5)
			case 0x33:
				n := registers[x]
				memory[index] = uint8(n / 100)
				n -= memory[index] * 100
				memory[index+1] = n / 10
				n -= memory[index+1] * 10
				memory[index+2] = n
			case 0x55:
				reg_dump(x, index)
			case 0x65:
				reg_load(x, index)
			}
		default:
			panic("Operação ilegal")
		}

		if next {
			pc += 2
		}

		next = true

		if delay_timer > 0 {
			delay_timer -= 1
		}

		if sound_timer > 2 {
			fmt.Print("\a")
		} else {
			sound_timer -= 1
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) < 2 {
		fmt.Print("Digite o nome do arquivo")
	} else {
		load_rom(os.Args[1])
		run()
	}
}
