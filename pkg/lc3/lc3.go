package lc3

import (
	"fmt"
	"sync"
	"time"
)

const (
	// R_R0 represents the index of the general purpose register 0.
	R_R0 uint16 = iota
	// R_R1 represents the index of the general purpose register 1.
	R_R1
	// R_R2 represents the index of the general purpose register 2.
	R_R2
	// R_R3 represents the index of the general purpose register 3.
	R_R3
	// R_R4 represents the index of the general purpose register 4.
	R_R4
	// R_R5 represents the index of the general purpose register 5.
	R_R5
	// R_R6 represents the index of the general purpose register 6.
	R_R6
	// R_R7 represents the index of the general purpose register 7.
	R_R7
	// R_PC represents the index of the program counter register.
	R_PC
	// R_COND represents the index of the condition flags register.
	R_COND
	// R_COUNT represents the number of registers.
	R_COUNT
)

const (
	// OP_BR represents the branching operation.
	OP_BR uint16 = iota
	// OP_ADD represents the add operation.
	OP_ADD
	// OP_LD represents the load operation.
	OP_LD
	// OP_ST represents the store operation.
	OP_ST
	// OP_JSR represents the jump register operation.
	OP_JSR
	// OP_AND represents the bitwise AND operation.
	OP_AND
	// OP_LDR represents the load register operation.
	OP_LDR
	// OP_STR represents the store register operation.
	OP_STR
	// OP_RTI represents a supervisor operation.
	OP_RTI
	// OP_NOT represents the bitwise NOT operation.
	OP_NOT
	// OP_LDI represents the load indirect operation.
	OP_LDI
	// OP_STI represents the store indirect operation.
	OP_STI
	// OP_JMP represents the jump operation.
	OP_JMP
	// OP_RES represents the reserve operation.
	OP_RES
	// OP_LEA represents the load effective address operation.
	OP_LEA
	// OP_TRAP represents the execute trap operation.
	OP_TRAP
)

const (
	// FL_POS represents a positive value in the R_COND register.
	FL_POS uint16 = 1 << iota
	// FL_ZRO represents a zero value in the R_COND register.
	FL_ZRO
	// FL_NEG represents a negative value in the R_COND register.
	FL_NEG
)

const (
	// MR_KBSR represents the memory address of the keyboard status.
	MR_KBSR uint16 = 0xFE00
	// MR_KBDR represents the memory address of the keyboard data.
	MR_KBDR uint16 = 0xFE02
)

const (
	// TRAP_GETC represents the TRAP action of reading a char from the keyboard (not echoed).
	TRAP_GETC uint16 = 0x20 + iota
	// TRAP_OUT represents the TRAP action of outputing a character.
	TRAP_OUT
	// TRAP_PUTS represents the TRAP action of outputing a string of characters.
	TRAP_PUTS
	// TRAP_IN represents the TRAP action of getting a char from the keyboard (echoed).
	TRAP_IN
	// TRAP_PUTSP represents the TRAP action of outputing a byte string.
	TRAP_PUTSP
	// TRAP_HALT represents the TRAP action that stops the program loop.
	TRAP_HALT
)

// MachineState represents the states of the LC-3 Virtual Machine.
type MachineState uint8

const (
	// MS_STOPPED represents the state of a stopped machine.
	MS_STOPPED MachineState = iota
	// MS_RUNNING represents the state of a machine processing instructions.
	MS_RUNNING
)

// Machine implements an LC-3 Virtual Machine.
type Machine struct {
	mem   [65536]uint16
	reg   [R_COUNT]uint16
	state MachineState

	sig         chan struct{}
	stdin       chan rune
	stdout      chan rune
	inputBuffer []rune
}

// NewMachine returns a new instance of the LC-3 Virtual Machine.
func NewMachine() *Machine {
	return &Machine{}
}

// LoadMemory loads the provided data into memory.
func (m *Machine) LoadMemory(mem [65536]uint16) {
	m.mem = mem
}

// ReadMemory returns the memory value for the provided address.
func (m *Machine) ReadMemory(addr uint16) uint16 {
	return m.mem[addr]
}

// LoadRegisters loads the provided data into the registers.
func (m *Machine) LoadRegisters(reg [R_COUNT]uint16) {
	m.reg = reg
}

// ReadRegister returns the register value for the provided address.
func (m *Machine) ReadRegister(addr uint16) uint16 {
	return m.reg[addr]
}

// SetStdin sets the standard input stream of the VM.
func (m *Machine) SetStdin(in chan rune) {
	m.stdin = in
}

// SetStdout sets the standard output stream of the VM.
func (m *Machine) SetStdout(out chan rune) {
	m.stdout = out
}

// State returns the state of the machine.
func (m *Machine) State() MachineState {
	return m.state
}

// Reset brings the machine back to it's inital PC and clears the condition register.
func (m *Machine) Reset() {
	m.reg[R_PC] = 0x3000
	m.reg[R_COND] = 0
}

// Start starts the instruction loop.
func (m *Machine) Start(sig chan struct{}, clockSpeed time.Duration) {
	m.sig = sig
	m.state = MS_RUNNING

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-m.sig:
				return
			case in := <-m.stdin:
				m.inputBuffer = append(m.inputBuffer, in)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-m.sig:
				return
			default:
				m.processInput()
				m.Step()
			}
		}
	}()

	wg.Wait()
}

// Stop stops the instruction loop.
func (m *Machine) Stop() {
	m.state = MS_STOPPED
	close(m.sig)
}

// Step reads an instruction and executes it.
func (m *Machine) Step() error {
	pc := m.reg[R_PC] + 1
	instr := m.mem[m.reg[R_PC]]

	switch instr >> 12 {
	case OP_ADD:
		dr := (instr >> 9) & 0x7
		sr1 := (instr >> 6) & 0x7
		immFlag := (instr >> 5) & 0x1
		if immFlag == 1 {
			imm5 := signExtend(instr&0x1F, 5)
			m.reg[dr] = m.reg[sr1] + imm5
		} else {
			sr2 := instr & 0x7
			m.reg[dr] = m.reg[sr1] + m.reg[sr2]
		}
		m.updateFlags(dr)

	case OP_AND:
		dr := (instr >> 9) & 0x7
		sr1 := (instr >> 6) & 0x7
		immFlag := (instr >> 5) & 0x1
		if immFlag == 1 {
			m.reg[dr] = m.reg[sr1] & signExtend(instr&0x1F, 5)
		} else {
			sr2 := instr & 0x7
			m.reg[dr] = m.reg[sr1] & m.reg[sr2]
		}
		m.updateFlags(dr)

	case OP_NOT:
		dr := (instr >> 9) & 0x7
		sr := (instr >> 6) & 0x7
		m.reg[dr] = ^m.reg[sr]
		m.updateFlags(dr)

	case OP_BR:
		nzp := (instr >> 9) & 0x7
		pcOffset := signExtend(instr&0x1FF, 9)
		if nzp&m.reg[R_COND] != 0 {
			pc += pcOffset
		}

	case OP_JMP:
		baseR := (instr >> 6) & 0x7
		pc = m.reg[baseR]

	case OP_JSR:
		m.reg[R_R7] = pc
		pcOffsetFlag := (instr >> 11) & 0x1
		if pcOffsetFlag == 1 {
			pc += signExtend(instr&0x7FF, 11)
		} else {
			pc = m.reg[(instr>>6)&0x7]
		}

	case OP_LD:
		dr := (instr >> 9) & 0x7
		m.reg[dr] = m.mem[pc+(instr&0xFF)]
		m.updateFlags(dr)

	case OP_LDI:
		dr := (instr >> 9) & 0x7
		m.reg[dr] = m.mem[m.mem[pc+(instr&0xFF)]]
		m.updateFlags(dr)

	case OP_LDR:
		dr := (instr >> 9) & 0x7
		baseR := (instr >> 6) & 0x7
		m.reg[dr] = m.mem[m.reg[baseR]+signExtend(instr&0x3F, 6)]
		m.updateFlags(dr)

	case OP_LEA:
		dr := (instr >> 9) & 0x7
		m.reg[dr] = pc + signExtend(instr&0x1FF, 9)
		m.updateFlags(dr)

	case OP_ST:
		sr := (instr >> 9) & 0x7
		m.mem[pc+signExtend(instr&0x1FF, 9)] = m.reg[sr]

	case OP_STI:
		sr := (instr >> 9) & 0x7
		m.mem[m.mem[pc+signExtend(instr&0x1FF, 9)]] = m.reg[sr]

	case OP_STR:
		sr := (instr >> 9) & 0x7
		baseR := (instr >> 6) & 0x7
		m.mem[m.reg[baseR]+signExtend(instr&0x3F, 6)] = m.reg[sr]

	case OP_TRAP:
		switch instr & 0xFF {
		case TRAP_GETC:
		loop:
			for {
				select {
				case <-m.sig:
					return nil
				default:
					if len(m.inputBuffer) > 0 {
						break loop
					}
				}
			}
			m.reg[R_R0], m.inputBuffer = uint16(m.inputBuffer[0]), m.inputBuffer[1:]

		case TRAP_OUT:
			m.stdout <- rune(m.reg[R_R0])

		case TRAP_PUTS:
			addr := m.reg[R_R0]
			var i uint16
			for {
				r := rune(m.mem[addr+i] & 0xFFFF)
				if r == rune(0) {
					break
				}
				m.stdout <- r
				i++
			}

		case TRAP_HALT:
			m.Stop()

		default:
			return fmt.Errorf("trap code '0x%04x' is not implemented", instr)
		}

	default:
		return fmt.Errorf("op code '0x%x' is not implemented", instr)
	}

	m.reg[R_PC] = pc
	return nil
}

func (m *Machine) processInput() {
	kbsrVal := m.mem[MR_KBSR]
	kbsrReady := ((kbsrVal & 0x8000) == 0)
	if kbsrReady && len(m.inputBuffer) > 0 {
		m.mem[MR_KBSR] = kbsrVal | 0x8000
		m.mem[MR_KBDR] = uint16(m.inputBuffer[0])
	}
}

func signExtend(x uint16, bitCount int) uint16 {
	if (x>>(bitCount-1))&1 == 1 {
		x |= (0xFFFF << bitCount)
	}
	return x
}

func (m *Machine) updateFlags(r uint16) {
	if m.reg[r] == 0 {
		m.reg[R_COND] = uint16(FL_ZRO)
	} else if (m.reg[r] >> 15) == 1 {
		m.reg[R_COND] = uint16(FL_NEG)
	} else {
		m.reg[R_COND] = uint16(FL_POS)
	}
}
