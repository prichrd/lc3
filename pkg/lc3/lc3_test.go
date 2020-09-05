package lc3_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/prichrd/lc3/pkg/lc3"
)

func TestStep_ADD(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R0, 10,
		lc3.R_R1, 25,
		lc3.R_R3, 0,
		lc3.R_R4, 0,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_ADD<<12|lc3.R_R2<<9|lc3.R_R0<<6|lc3.R_R1,
		0x3001, lc3.OP_ADD<<12|lc3.R_R2<<9|lc3.R_R0<<6|1<<5|0b11111,
		0x3002, lc3.OP_ADD<<12|lc3.R_R2<<9|lc3.R_R4<<6|lc3.R_R3,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R2, 35)
	m.Step()
	assertReg(t, m, lc3.R_R2, 9)
	m.Step()
	assertReg(t, m, lc3.R_R2, 0)
}

func TestStep_AND(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R0, 0b00111,
		lc3.R_R1, 0b01011,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_AND<<12|lc3.R_R2<<9|lc3.R_R0<<6|lc3.R_R1,
		0x3001, lc3.OP_AND<<12|lc3.R_R2<<9|lc3.R_R0<<6|1<<5|0b01110,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R2, 3)
	m.Step()
	assertReg(t, m, lc3.R_R2, 6)
}

func TestStep_NOT(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R0, 0b000000000000111,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_NOT<<12|lc3.R_R2<<9|lc3.R_R0<<6,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R2, 0b1111111111111000)
}

func TestStep_BR(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_PC, 0x3000,
		lc3.R_COND, 0b111,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_BR<<12|7<<9|0x10,
	))

	m.Step()
	assertReg(t, m, lc3.R_PC, 0x3011)
}

func TestStep_JMP(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R0, 0x3005,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_JMP<<12|lc3.R_R0<<6,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_PC, 0x3005)
}

func TestStep_JSR(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R5, 0x300A,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_JSR<<12|1<<11|0x5,
		0x3006, lc3.OP_JSR<<12|lc3.R_R5<<6,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R7, 0x3001)
	assertReg(t, m, lc3.R_PC, 0x3006)
	m.Step()
	assertReg(t, m, lc3.R_R7, 0x3007)
	assertReg(t, m, lc3.R_PC, 0x300A)
}

func TestStep_LD(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3006, 0x5,
		0x3000, lc3.OP_LD<<12|lc3.R_R1<<9|0x5,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R1, 0x5)
}

func TestStep_LDI(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3007, 0x5,
		0x3006, 0x3007,
		0x3000, lc3.OP_LDI<<12|lc3.R_R1<<9|0x5,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R1, 0x5)
}

func TestStep_LDR(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R4, 0x3005,
	))
	m.LoadMemory(initMem(
		0x3006, 0x300F,
		0x3000, lc3.OP_LDR<<12|lc3.R_R1<<9|lc3.R_R4<<6|0x1,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R1, 0x300F)
}

func TestStep_LEA(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3000, lc3.OP_LEA<<12|lc3.R_R1<<9|0x1,
	))

	m.Reset()
	m.Step()
	assertReg(t, m, lc3.R_R1, 0x3002)
}

func TestStep_ST(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R1, 0x2,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_ST<<12|lc3.R_R1<<9|0x1,
	))

	m.Reset()
	m.Step()
	assertMem(t, m, 0x3002, 0x2)
}

func TestStep_STI(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R1, 0x2,
	))
	m.LoadMemory(initMem(
		0x3002, 0x3004,
		0x3000, lc3.OP_STI<<12|lc3.R_R1<<9|0x1,
	))

	m.Reset()
	m.Step()
	assertMem(t, m, 0x3004, 0x2)
}

func TestStep_STR(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadRegisters(initReg(
		lc3.R_R1, 0x3000,
		lc3.R_R2, 0x3001,
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_STR<<12|lc3.R_R1<<9|lc3.R_R2<<6|0x1,
	))

	m.Reset()
	m.Step()
	assertMem(t, m, 0x3002, 0x3000)
}

func TestStep_TRAP_GETC(t *testing.T) {
	in := make(chan rune)
	m := lc3.NewMachine()
	m.Reset()
	m.SetStdin(in)
	m.LoadMemory(initMem(
		lc3.MR_KBSR, 0,
		0x3000, lc3.OP_TRAP<<12|lc3.TRAP_GETC,
		0x3001, lc3.OP_TRAP<<12|lc3.TRAP_HALT,
	))

	sig := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		m.Start(sig, time.Nanosecond)
		wg.Done()
	}()
	in <- 'a'
	wg.Wait()
	assertReg(t, m, lc3.R_R0, uint16('a'))
}

func TestStep_TRAP_OUT(t *testing.T) {
	out := make(chan rune)
	m := lc3.NewMachine()
	m.SetStdout(out)
	m.LoadRegisters(initReg(
		lc3.R_R0, uint16('a'),
	))
	m.LoadMemory(initMem(
		0x3000, lc3.OP_TRAP<<12|lc3.TRAP_OUT,
	))

	m.Reset()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		m.Step()
		wg.Done()
	}()

	var r rune
	go func() {
		r = <-out
		wg.Done()
	}()
	wg.Wait()

	if r != 'a' {
		t.Errorf("Out should be %v, but got %v", 'a', r)
	}
}

func TestStep_TRAP_PUTS(t *testing.T) {
	out := make(chan rune)
	m := lc3.NewMachine()
	m.SetStdout(out)

	m.LoadRegisters(initReg(
		lc3.R_R0, 0x5000,
		lc3.R_PC, 0x3000,
	))

	m.LoadMemory(initMem(
		0x5000, 'f',
		0x5001, 'o',
		0x5002, 'o',
		0x3000, lc3.OP_TRAP<<12|lc3.TRAP_PUTS,
	))

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		m.Step()
		close(out)
		wg.Done()
	}()

	o := ""
	go func() {
		for r := range out {
			o += string(r)
		}
		wg.Done()
	}()
	wg.Wait()

	if o != "foo" {
		t.Errorf("Out should be %v, but got %v", "foo", o)
	}
}

func TestStep_TRAP_HALT(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3000, lc3.OP_TRAP<<12|lc3.TRAP_HALT,
	))

	m.Reset()

	sig := make(chan struct{})
	m.Start(sig, time.Nanosecond)
	if m.State() != lc3.MS_STOPPED {
		t.Errorf("Machine should be stopped, bug got %v", m.State())
	}
}

func TestStep_unhandled_op_code(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3000, 0b1101<<12,
	))

	m.Reset()
	err := m.Step()
	expErr := errors.New("op code '0xd000' is not implemented")
	if err.Error() != expErr.Error() {
		t.Errorf("expected err \"%v\", but got: \"%v\"", expErr, err)
	}
}

func TestStep_unhandled_trap_code(t *testing.T) {
	m := lc3.NewMachine()
	m.LoadMemory(initMem(
		0x3000, lc3.OP_TRAP<<12|0b0001<<9,
	))

	m.Reset()
	err := m.Step()
	expErr := errors.New("trap code '0xf200' is not implemented")
	if err.Error() != expErr.Error() {
		t.Errorf("expected err \"%v\", but got: \"%v\"", expErr, err)
	}
}

func initReg(addrvalues ...uint16) [lc3.R_COUNT]uint16 {
	reg := [lc3.R_COUNT]uint16{}
	for i := 0; i < len(addrvalues); i++ {
		reg[addrvalues[i]] = addrvalues[i+1]
		i++
	}
	return reg
}

func initMem(addrinstructions ...uint16) [65536]uint16 {
	mem := [65536]uint16{}
	for i := 0; i < len(addrinstructions); i++ {
		mem[addrinstructions[i]] = addrinstructions[i+1]
		i++
	}
	return mem
}

func assertReg(t *testing.T, m *lc3.Machine, addr uint16, val uint16) {
	if m.ReadRegister(addr) != val {
		t.Errorf("Register[%#x] should be %v, but got %v", addr, val, m.ReadRegister(addr))
	}
}

func assertMem(t *testing.T, m *lc3.Machine, addr uint16, val uint16) {
	if m.ReadMemory(addr) != val {
		t.Errorf("Memory[%#x] should be %v, but got %v", addr, val, m.ReadMemory(addr))
	}
}
