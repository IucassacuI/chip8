package cpu
import (
	"math/rand"
	"time"
)

type CPU struct {
	RAM  *[4096]uint8
	VRAM *[32][64]uint8

	stack []uint16

	registers [16]uint8
	index uint16
	pc uint16

	delay_timer uint8
	sound_timer uint8
}

type Instruction struct {
	opcode uint16
	nnn uint16
	nn uint8
	n uint8
	x uint8
	y uint8
}

func (cpu *CPU) Init(RAM *[4096]uint8, VRAM *[32][64]uint8){
	cpu.pc = 0x200
	cpu.RAM = RAM
	cpu.VRAM = VRAM

	rand.Seed(time.Now().UnixNano())
}

func (cpu *CPU) Push(addr uint16){
	cpu.stack = append(cpu.stack, addr)
}

func (cpu *CPU) Top() uint16 {
	return cpu.stack[len(cpu.stack)-1]
}

func (cpu *CPU) Pop() uint16 {
	addr := cpu.stack[len(cpu.stack)-1]
	cpu.stack = cpu.stack[:len(cpu.stack)-1]

	return addr
}

func (cpu *CPU) reg_dump(vx uint8, i uint16) {
	for offset := 0; uint8(offset) < vx+1; offset++ {
		(*cpu.RAM)[cpu.index+uint16(offset)] = cpu.registers[offset]
	}
}

func (cpu *CPU) reg_load(vx uint8, i uint16) {
	for offset := 0; uint8(offset) < vx+1; offset++ {
		cpu.registers[offset] = (*cpu.RAM)[cpu.index+uint16(offset)]
	}
}

func (cpu *CPU) Fetch() uint16 {
	var opcode uint16
	opcode = uint16((*cpu.RAM)[cpu.pc]) << 8 | uint16((*cpu.RAM)[cpu.pc + 1])

	return opcode
}

func (cpu *CPU) Decode(opcode uint16) Instruction {
	return Instruction{
		opcode: opcode & 0xf000,
		nnn: opcode & 0x0fff,
		nn: uint8(opcode & 0x00ff),
		n:  uint8(opcode & 0x000f),
		x:  uint8((opcode & 0x0f00) >> 8),
		y:  uint8((opcode & 0x00f0) >> 4),
	}
}

func (cpu *CPU) Draw(vx, vy, n uint8){
}

func (cpu *CPU) Execute(inst Instruction){
	var next bool = true
	var keyp uint8

	switch(inst.opcode){
		case 0:
		if inst.nn == 0xe0  {
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					(*cpu.VRAM)[y][x] = 0
				}
			}
		} else if inst.nn == 0xee {
			cpu.pc = cpu.Pop()
		} else {
			panic("operação ilegal")
		}

		case 0x1000:
		cpu.pc = inst.nnn
		next = false

		case 0x2000:
		cpu.Push(cpu.pc)
		cpu.pc = inst.nnn

		next = false

		case 0x3000:
		if cpu.registers[inst.x] == inst.nn {
			cpu.pc += 2
		}

		case 0x4000:
		if cpu.registers[inst.x] != inst.nn {
			cpu.pc += 2
		}

		case 0x5000:
		if cpu.registers[inst.x] == cpu.registers[inst.y]{
			cpu.pc += 2
		}

		case 0x6000:
		cpu.registers[inst.x] = inst.nn

		case 0x7000:
		cpu.registers[inst.x] += inst.nn

		case 0x8000:
		switch inst.n {
		case 0:
			cpu.registers[inst.x] = cpu.registers[inst.y]

		case 1:
			cpu.registers[inst.x] |= cpu.registers[inst.y]

		case 2:
			cpu.registers[inst.x] &= cpu.registers[inst.y]

		case 3:
			cpu.registers[inst.x] ^= cpu.registers[inst.y]

		case 4:
			if cpu.registers[inst.y] > 0xff - cpu.registers[inst.y] {
				cpu.registers[15] = 1
			} else {
				cpu.registers[15] = 0
			}

			cpu.registers[inst.x] += cpu.registers[inst.y]

		case 5:
			if cpu.registers[inst.y] > cpu.registers[inst.x] {
				cpu.registers[15] = 0
			} else {
				cpu.registers[15] = 1
			}

			cpu.registers[inst.x] -= cpu.registers[inst.y]

		case 6:
			cpu.registers[15] = cpu.registers[inst.x] & 1
			cpu.registers[inst.x] >>= 1

		case 7:
			if cpu.registers[inst.x] > cpu.registers[inst.y] {
				cpu.registers[15] = 0
			} else {
				cpu.registers[15] = 1
			}

			cpu.registers[inst.x] = cpu.registers[inst.y] - cpu.registers[inst.x]

		case 0xe:
			cpu.registers[15] = cpu.registers[inst.x] >> 7
			cpu.registers[inst.x] <<= 1
		default:
			panic("operação ilegal")
		}

		case 0x9000:
		if cpu.registers[inst.x] != cpu.registers[inst.y] {
			cpu.pc += 2
		}

		case 0xa000:
		cpu.index = inst.nnn

		case 0xb000:
		cpu.pc = uint16(cpu.registers[0]) + inst.nnn
		next = false

		case 0xc000:
		cpu.registers[inst.x] = uint8(rand.Intn(255)) & inst.nn

		case 0xd000:
		cpu.registers[0xF] = 0

		var y uint16 = 0
		var x uint16 = 0
		for y = 0; y < uint16(inst.n); y++ {
			pixel := (*cpu.RAM)[cpu.index+y]

			for x = 0; x < 8; x++ {
				if (pixel & (0x80 >> x)) != 0 {
					if (*cpu.VRAM)[(cpu.registers[inst.y] + uint8(y))%32][(cpu.registers[inst.x]+uint8(x))%64] == 1 {
						cpu.registers[0xF] = 1
					}
					(*cpu.VRAM)[(cpu.registers[inst.y]  + uint8(y))%32][(cpu.registers[inst.x]+uint8(x))%64] ^= 1
				}
			}
		}
		case 0xe000:
			switch inst.n {
			case 0xE:
					if keyp == cpu.registers[inst.x] {
						cpu.pc += 2
					}
			case 1:
					if keyp != cpu.registers[inst.x] {
						cpu.pc += 2
					}
			default:
				panic("Operação ilegal.")
			}

		case 0xf000:
			switch inst.nn {
			case 7:
				cpu.registers[inst.x] = cpu.delay_timer
			case 0xA:
					for {
						if keyp != 0 {
							break
						}
					}
					cpu.registers[inst.x] = keyp
			case 0x15:
				cpu.delay_timer = cpu.registers[inst.x]
			case 0x18:
				cpu.sound_timer = cpu.registers[inst.x]
			case 0x1E:
				cpu.index += uint16(cpu.registers[inst.x])
			case 0x29:
				cpu.index = uint16(cpu.registers[inst.x] * 5)
			case 0x33:
				n := cpu.registers[inst.x]
				(*cpu.RAM)[cpu.index] = uint8(n / 100)

				n -= (*cpu.RAM)[cpu.index] * 100
				(*cpu.RAM)[cpu.index+1] = n / 10

				n -= (*cpu.RAM)[cpu.index+1] * 10
				(*cpu.RAM)[cpu.index+2] = n
			case 0x55:
				cpu.reg_dump(inst.x, cpu.index)
			case 0x65:
				cpu.reg_load(inst.x, cpu.index)
			default:
				panic("operação ilegal")
			}

		default:
		panic("operação ilegal")
	}

		if next {cpu.pc += 2}

		if cpu.delay_timer > 0 {cpu.delay_timer--}

		if cpu.sound_timer > 2 {
			// sound
		} else {
			cpu.sound_timer--
		}
}
