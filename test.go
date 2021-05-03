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
	var pc_in_mask = 2048
	var zf_clear_mask = 4096
	var zf_set_mask = 8192
	var zf_out_mask = 16384

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
	var zf = 0
	var ram = make(map[int]int)

	// All instructions have the same first three microcodes
	//
	//		T0: Program counter out and MAR in
	//		T1: MAR out and IR int
	//		T2: Increment Program Counter
	//
	// These are stores with a key = step as first 8 bits and instruction code as second
	// Key
	//		Segment 1: Step count that the action applies to
	//		Segment 2: The instruction number being executed
	// Control Word - the value returned bu the array
	//		The control lines to activate and de-activate 1/0
	var decoder = make(map[int]int)
	var controlWord int

	// LDA Literal - Value is in the next memory loc - Instruction 0000
	decoder[0*256+0x00] = pc_out_mask + mar_in_mask
	decoder[1*256+0x00] = ram_out_mask + ir_in_mask
	decoder[2*256+0x00] = pc_inc_mask
	decoder[3*256+0x00] = pc_out_mask + mar_in_mask
	decoder[4*256+0x00] = ram_out_mask + acc_load_mask
	decoder[5*256+0x00] = step_reset_mask + pc_inc_mask

	// ADC Literal - Value is in the next memory loc - Instruction 0001
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

	// Jump    - Instruction 0004
	decoder[0*256+0x04] = pc_out_mask + mar_in_mask
	decoder[1*256+0x04] = ram_out_mask + ir_in_mask
	decoder[2*256+0x04] = pc_inc_mask
	decoder[3*256+0x04] = pc_out_mask + mar_in_mask
	decoder[4*256+0x04] = ram_out_mask
	decoder[5*256+0x04] = pc_in_mask
	decoder[6*256+0x04] = step_reset_mask

	// EQZ - Equal Zero
	// If zero flag is set skip ahead 1 pc
	// Otherwise execute the next instruction
	// Destroys value in accum - Instruction 0005
	decoder[0*256+0x05] = pc_out_mask + mar_in_mask
	decoder[1*256+0x05] = ram_out_mask + ir_in_mask
	decoder[2*256+0x05] = pc_inc_mask
	decoder[3*256+0x05] = pc_out_mask + acc_load_mask
	decoder[4*256+0x05] = zf_out_mask
	decoder[5*256+0x05] = acc_add_mask
	decoder[6*256+0x05] = acc_out_mask
	decoder[7*256+0x05] = pc_in_mask
	decoder[8*256+0x05] = step_reset_mask

	// Halt - Instruction 15
	decoder[0*256+0x0f] = pc_out_mask + mar_in_mask
	decoder[1*256+0x0f] = ram_out_mask + ir_in_mask
	decoder[2*256+0x0f] = pc_inc_mask
	decoder[3*256+0x0f] = halt_exec_mask

	// Test Program
	ram[0] = 0x00  // LDA
	ram[1] = 0x0f  // Data for LDA
	ram[2] = 0x01  // ADC
	ram[3] = 0x01  // Date for ADC
	ram[4] = 0x03  // Inc the Accum
	ram[5] = 0x04  // Jmp
	ram[6] = 0x09  // Jump to Address is 9
	ram[9] = 0x00  // LDA
	ram[10] = 0x00 // Data for accum
	ram[11] = 0x05 // EQZ
	ram[12] = 0x0f // Halt if zero flag not set
	ram[13] = 0x00 // LDA
	ram[14] = 0x01 // Load a 1 for zero flag is set
	ram[15] = 0x0f // Halt

	for i := 0; i < 54; i++ {

		controlWord = decoder[step*256+ir]
		fmt.Printf("Step #%d, PC is %d, IR is %08b, Bus value is %08b, acc is %d ---", step, pc, ir, bus, acc)
		fmt.Printf(" Executing word %016b\n", controlWord)

		if controlWord&pc_out_mask == pc_out_mask {
			bus = pc
		}

		if controlWord&pc_inc_mask == pc_inc_mask {
			pc++
		}

		if controlWord&pc_in_mask == pc_in_mask {
			pc = bus
			//fmt.Println("Updating pc to ", bus)
		}

		if controlWord&mar_in_mask == mar_in_mask {
			mar = bus
		}

		if controlWord&ram_out_mask == ram_out_mask {
			bus = ram[mar]
			//fmt.Printf("Reading from addr %d and putting %d on bus\n", mar, bus)
		}

		if controlWord&acc_load_mask == acc_load_mask {
			acc = bus
			if acc == 0 {
				zf = 1
			}
		}

		if controlWord&acc_add_mask == acc_add_mask {
			acc += bus
			if acc == 0 {
				zf = 1
			}
		}

		if controlWord&acc_out_mask == acc_out_mask {
			bus = acc
		}

		if controlWord&ir_in_mask == ir_in_mask {
			ir = bus
		}

		if controlWord&zf_clear_mask == zf_clear_mask {
			zf = 0
		}

		if controlWord&zf_set_mask == zf_set_mask {
			zf = 1
		}

		if controlWord&zf_out_mask == zf_out_mask {
			bus = zf
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
