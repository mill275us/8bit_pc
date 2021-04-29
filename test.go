package main

import "fmt"

func main() {
	fmt.Println("    Starting     ")

	// Masks represent inline and gates to control via control wires
	// What is read in and outputed from the various components
	// These are also used to generate the microcode word
	var pc_out_mask = 1
	var pc_inc_mask = 2
	var mar_in_mask = 4
	var ram_out_mask = 8
	var acc_load_mask = 16
	var acc_add_mask = 32
	var acc_out_mask = 64
	var ir_in_mask = 128
	var step_reset_mask = 256
	var halt_exec_mask = 512
	var acc_inc_mask = 1024

	// Components
	// pc = program counter
	// step = step counter
	// bus = represents the 8 bit bus that is used for data and addressing
	// mar = the Memory Access register
	// ir = the Instruction register
	// acc = contents of the accummmulator
	var pc = 0
	var step = 0
	var bus = 0
	var mar = 0
	var ir = 0
	var acc = 0
	var ram = make(map[int]int)

	// All instructions have the same first three microcodes
	//
	//		T0: Program counter out and MAR in
	//		T1: MAR out and IR int
	//		T2: Increment Program Counter
	//
	// These are stores with a key = step as first 8 bits and instruction code as second
	// Value output is the control word
	var decoder = make(map[int]int)
	var controlWord int

	// LDA from Direct Memmory Address - Instruction 0000
	decoder[0*256+0x00] = pc_out_mask + mar_in_mask
	decoder[1*256+0x00] = ram_out_mask + ir_in_mask
	decoder[2*256+0x00] = pc_inc_mask
	decoder[3*256+0x00] = pc_out_mask + mar_in_mask
	decoder[4*256+0x00] = ram_out_mask + acc_load_mask
	decoder[5*256+0x00] = step_reset_mask + pc_inc_mask

	// ADC from Direct Memmory Address - Instruction 0001
	decoder[0*256+0x01] = pc_out_mask + mar_in_mask
	decoder[1*256+0x01] = ram_out_mask + ir_in_mask
	decoder[2*256+0x01] = pc_inc_mask
	decoder[3*256+0x01] = pc_out_mask + mar_in_mask
	decoder[4*256+0x01] = ram_out_mask + acc_add_mask
	decoder[5*256+0x01] = step_reset_mask + pc_inc_mask

	// STA to Direct Memory Address - Instruction 0002

	// Inc Accum  - Instruction 0003
	decoder[0*256+0x03] = pc_out_mask + mar_in_mask
	decoder[1*256+0x03] = ram_out_mask + ir_in_mask
	decoder[2*256+0x03] = pc_inc_mask
	decoder[3*256+0x03] = acc_inc_mask
	decoder[4*256+0x03] = step_reset_mask

	// JMP to Address

	// BNE to Address

	// Halt - Instruction 15
	decoder[0*256+0x0f] = pc_out_mask + mar_in_mask
	decoder[1*256+0x0f] = ram_out_mask + ir_in_mask
	decoder[2*256+0x0f] = pc_inc_mask
	decoder[3*256+0x0f] = halt_exec_mask

	// Test Program
	ram[0] = 0x00 // LDA
	ram[1] = 0x0f // Data for LDA
	ram[2] = 0x01 // ADC
	ram[3] = 0x01 // Date for ADC
	ram[4] = 0x03 // Inc the Accum
	ram[5] = 0x0f // Halt

	for i := 0; i < 34; i++ {

		controlWord = decoder[step*256+ir]
		fmt.Printf("Step #%d, PC is %d, IR is %08b, Bus value is %08b, acc is %d ---", step, pc, ir, bus, acc)
		fmt.Printf(" Executing word %012b\n", controlWord)

		if controlWord&pc_out_mask == pc_out_mask {
			bus = pc
		}

		if controlWord&pc_inc_mask == pc_inc_mask {
			pc++
		}

		if controlWord&mar_in_mask == mar_in_mask {
			mar = bus
		}

		if controlWord&ram_out_mask == ram_out_mask {
			bus = ram[mar]
		}

		if controlWord&acc_load_mask == acc_load_mask {
			acc = bus
		}

		if controlWord&acc_add_mask == acc_add_mask {
			acc += bus
		}

		if controlWord&acc_out_mask == acc_out_mask {
			bus = acc
		}

		if controlWord&ir_in_mask == ir_in_mask {
			ir = bus
		}

		if controlWord&halt_exec_mask == halt_exec_mask {
			fmt.Println("Halt command found")
			i = 999999999
		}

		if controlWord&acc_inc_mask == acc_inc_mask {
			acc++
		}

		if controlWord&step_reset_mask == step_reset_mask {
			step = 0
		} else {
			step++
		}

	}

}
