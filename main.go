package main

import (
	"github.com/hajimehoshi/ebiten/v2"
  "github.com/IucassacuI/chip8/cpu"
	"image/color"
	"os"
)

type CHIP8 struct {
	RAM [4096]uint8
	VRAM [32][64]uint8

	CPU cpu.CPU
}

func (c *CHIP8) Update() error {
	var opcode uint16 = c.CPU.Fetch()
	var inst cpu.Instruction = c.CPU.Decode(opcode)

	c.CPU.Execute(inst)

	return nil
}

func (c *CHIP8) Draw(screen *ebiten.Image) {
	var frame *ebiten.Image = ebiten.NewImage(64, 32)

	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if c.VRAM[y][x] == 1 {
				frame.Set(x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
			} else {
				frame.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(10, 10)

	screen.DrawImage(frame, options)
}

func (c *CHIP8) Layout(outWidth, outHeight int) (int, int) {
	return outWidth, outHeight
}

func main(){
	var chip8 CHIP8 = CHIP8{
		CPU: cpu.CPU{},
	}

	chip8.CPU.Init(&chip8.RAM, &chip8.VRAM)

	ROM, err := os.ReadFile(os.Args[1])
	if err != nil {os.Exit(1)}

	for i, b := range ROM {
		chip8.RAM[0x200+i] = b
	}

	ebiten.SetWindowTitle("CHIP-8")
	ebiten.SetWindowSize(640, 320)
	ebiten.SetTPS(500)

	if err := ebiten.RunGame(&chip8); err != nil {
		panic(err)
	}
}
