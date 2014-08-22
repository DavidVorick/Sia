# Bytecode Specification #

Scripting is an integral part of the Sia network. Every wallet has a script associated with it, and during compilation each script executes using inputs supplied to it in heartbeats. Scripts are written in a Sia-specific *bytecode*, a low-level language similar to assembly. Low-level languages are more susceptible to programming errors, so most users should stick with the scripts provided in [scripts.go](../src/delta/scripts.go). However, anyone looking to take full advantage of the scripting system should study the specification outlined here.

## Properties ##

The Sia bytecode is Turing-complete, but places a limit on the total number of instructions that can executed during any given run. This prevents infinite loops and discourages prohibitively expensive computations.

The bytecode is stack-based. Elements on the stack are arbitrary-length byte slices. Depending on the opcode used, these slices can be interpreted as integers, floats, or other types. In addition to the stack, a set of addressable registers, each containing a byte slice, are provided for convenience.

## Execution ##

Scripts only run if they are provided with an *input*. An input, just like a script, is a byte slice. Before execution begins, the input is appended to the end of the script.

Like most bytecodes, execution is facilitated by an instruction pointer, or iptr. The iptr begins at the first byte of the script and looks up its associated opcode function. If the opcode requires arguments, the bytes after the opcode byte are passed to it, and the iptr is incremented once for each argument byte. *(Note: this must be taken into account when calculating offset addresses.)*

The Sia bytecode also makes use of a separate pointer call the *data pointer*, or dptr. Unlike the iptr, the dptr can only be controlled via opcodes. It is used to read and write bytes in the script or input. Think of it as a cursor, as in a text editor. The dptr is often used to read inputs into registers. It seeks to the end of the data it reads, allowing for easy sequential reads.

## Resources ##

Scripts have access to a finite quantity of resources, including wallet balance, number of instructions, allocated memory, and more. The exact mechanics of how resources are allocated and charged are still being worked out. Exhausting any of the resources will cause the script to terminate. Support may be added later for making sets of instructions "atomic:" either all of them will execute or none of them will.

## Termination ##

A number of conditions can cause a script to stop executing. The most benign is upon encountering the `exit` bytecode `0xFF`, or upon reaching the end of the script. Another opcode, `reject`, terminates execution with a special error that indicates the script owner should not be charged for any resources used. (This is to protect scripts from malicious inputs.) Finally, there are a multitude of errors that can cause the script to terminate mid-execution, such as dividing by zero, popping an empty stack, or passing malformed data to an opcode. If a script terminates in this way, the owner will still be charged for the resources used.

After the script terminates without error, it is saved to disk. This means that any changes to the script body will be present upon the next execution of the script.

## Limitations ##

Many of the operations that people would like to do via scripting, such as resizing a sector or verifying a signature, are simply too complex to be implemented via low-level operations. As such, these operations are accessible only through specific opcodes; there are no low-level IO primitives for sending network requests or reading files from the disk. While this greatly limits the scope of operations that a script can perform, we feel that this is appropriate for the level of generality expected of the scripting system.

There is currently no support for writing new procedures (functions, methods, etc.) in the bytecode. That is, you cannot define something like a factorial function and later call it with a supplied argument. Functions are not obviously aligned with the goals of the scripting system, so it is doubtful that they will be support in the future. However, functions can be crudely approximated through the use of `goto`.

## Notes ##

The scripting system is still in its infancy, and is subject to API-breaking changes. That said, some guidelines can still be provided for people looking to write their own scripts.

Generally, you will want to protect your scripts using public key cryptography. To accomplish this, place your public key in the script body, and supply a cryptographic signature in any inputs you submit. The `verify` opcode can be used to verify cryptographic signatures. If verification fails, use `reject` (or more succinctly, `cond_reject`) to halt execution.

Most of the more complex operations, such as proposing an upload to the quorum, require many arguments. Since opcodes are limited (for now) to two arguments, the current approach is to encode multiple arguments into one byte slice, store the byte slice in a register, and reference the register in the opcode. This is not a permanent solution, but in the meantime you should expect to make heavy use of the dptr to load and store arguments.

## List of bytecodes ##

Note that some of these descriptions are insufficient to explain the format of the data to be passed as arguments or other details. For a more exact specification of the function of each opcode, consult their implementations in [instructions.go](../src/delta/instructions.go)

| Hex  | Name          | Args | Description                                                                            |
|------|---------------|------|----------------------------------------------------------------------------------------|
| 0x00 | no_op         | 0    | do nothing                                                                             |
| 0x01 | push_byte     | 1    | push a byte ($1) onto the stack                                                        |
| 0x02 | push_short    | 2    | push a short ($1$2) onto the stack                                                     |
| 0x03 | pop           | 0    | pop a stack value, discarding it                                                       |
| 0x04 | dup           | 0    | duplicate a stack value                                                                |
| 0x05 | swap          | 0    | swap two stack values                                                                  |
| 0x06 | add_int       | 0    | integer addition                                                                       |
| 0x07 | add_float     | 0    | floating point addition                                                                |
| 0x08 | sub_int       | 0    | integer subtraction                                                                    |
| 0x09 | sub_float     | 0    | floating point subtraction                                                             |
| 0x0A | mul_int       | 0    | integer multiplication                                                                 |
| 0x0B | mul_float     | 0    | floating point multiplication                                                          |
| 0x0C | div_int       | 0    | integer division                                                                       |
| 0x0D | div_float     | 0    | floating point division                                                                |
| 0x0E | mod_int       | 0    | integer modulus                                                                        |
| 0x0F | neg_int       | 0    | integer negation                                                                       |
| 0x10 | neg_float     | 0    | floating point negation                                                                |
| 0x11 | binary_or     | 0    | integer binary or                                                                      |
| 0x12 | binary_and    | 0    | integer binary and                                                                     |
| 0x13 | binary_xor    | 0    | integer binary xor                                                                     |
| 0x14 | shift_left    | 1    | shift integer left by $1 bits                                                          |
| 0x15 | shift_right   | 1    | shift integer right by $1 bits                                                         |
| 0x16 | equal         | 0    | test stack values for equality; push 1 (true) or 0 (false)                             |
| 0x17 | not_equal     | 0    | inequality                                                                             |
| 0x18 | less_int      | 0    | integer less than                                                                      |
| 0x19 | less_float    | 0    | floating point less than                                                               |
| 0x1A | greater_int   | 0    | integer greater than                                                                   |
| 0x1B | greater_float | 0    | floating point greater than                                                            |
| 0x1C | logical_not   | 0    | logical negation                                                                       |
| 0x1D | logical_or    | 0    | logical or                                                                             |
| 0x1E | logical_and   | 0    | logical and                                                                            |
| 0x1F | if_goto       | 2    | if non-zero, jump to instruction at offset formed by $1$2                              |
| 0x20 | if_move       | 2    | same as if_goto, but with a relative, rather than absolute, address                    |
| 0x21 | goto          | 2    | unconditional jump                                                                     |
| 0x22 | move          | 2    | unconditional move                                                                     |
| ---- | ----          | -    | data pointer and register opcodes                                                      |
| 0x30 | store         | 1    | pop a stack value into register $1                                                     |
| 0x31 | load          | 1    | push a stack value from register $1                                                    |
| 0x32 | data_goto     | 2    | move data pointer to address $1$2                                                      |
| 0x33 | data_move     | 2    | move data pointer by offset $1$2                                                       |
| 0x34 | data_push     | 1    | push (and move dptr) $1 bytes (zero-padded) from data pointer onto stack               |
| 0x35 | data_store    | 2    | store (and move dptr) $1 bytes (zero-padded) from data pointer in register $2          |
| 0x36 | data_copy     | 1    | copy (and move dptr) popped number of bytes (max 2^16) into register $1                |
| 0x37 | data_paste    | 1    | overwrite script with popped number of bytes (max 2^16) from register $1               |
| 0x38 | transfer      | 0    | move instruction pointer to data pointer                                               |
| ---- | ----          | -    | function opcodes                                                                       |
| 0x40 | verify        | 0    | verify a signature; pushes boolean success value                                       |
| 0x41 | add_sibling   | 0    | add sibling; pushes boolean success value                                              |
| 0x42 | add_wallet    | 0    | add a wallet with an initial balance and script                                        |
| 0x43 | send          | 0    | send siacoins from host wallet to recipient                                            |
| 0x44 | resize_sec    | 0    | resize the sector associated with a given wallet (erases current sector data)          |
| 0x45 | prop_upload   | 0    | propose an upload to the quorum (arguments are stored in one gob-encoded register)     |
| ---- | ----          | -    | convenience opcodes                                                                    |
| 0xE0 | switch        | 2    | if value and $1 are equal, branch to $2. The value is only consumed upon equality.     |
| 0xE1 | store_prefix  | 1    | same as data_copy, but using the first two bytes to determine the length               |
| 0xE2 | store_rest    | 1    | copy from data pointer to end of script into register $1 (dptr does not move)          |
| 0xE3 | push_prefix   | 0    | same as data_push, but using the first two bytes to determine the length               |
| 0xE4 | push_rest     | 0    | push from data pointer to end of script (dptr does not move)                           |
| 0xE5 | cond_reject   | 0    | if false, reject (otherwise no op)                                                     |
| 0xE6 | data_seek     | 1    | move data pointer past next occurence of $1 (ignoring this one)                        |
| ---- | ----          | -    | termination opcodes                                                                    |
| 0xFE | reject        | 0    | reject input, terminating execution                                                    |
| 0xFF | exit          | 0    | terminates execution                                                                   |
