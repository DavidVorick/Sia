List of bytecodes
-----------------

| Hex  | Name          | Args | Description                                                                            |
|------|---------------|------|----------------------------------------------------------------------------------------|
| 0x00 | no_op         | 0    | do nothing                                                                             |
| 0x01 | push_byte     | 1    | push a byte ($1) onto the stack                                                        |
| 0x02 | push_short    | 2    | push a short ($1$2) onto the stack                                                     |
| 0x03 | pop           | 0    | pop a value, discarding it                                                             |
| 0x04 | dup           | 0    | duplicate a value                                                                      |
| 0x05 | swap          | 0    | swap two values                                                                        |
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
| 0x16 | equal         | 0    | test values for equality; push 1 (true) or 0 (false)                                   |
| 0x17 | not_equal     | 0    | inequality                                                                             |
| 0x18 | less_int      | 0    | integer less than                                                                      |
| 0x19 | less_float    | 0    | floating point less than                                                               |
| 0x1A | greater_int   | 0    | integer greater than                                                                   |
| 0x1B | greater_float | 0    | floating point greater than                                                            |
| 0x1C | logical_not   | 0    | logical negation                                                                       |
| 0x1D | logical_or    | 0    | logical or                                                                             |
| 0x1E | logical_and   | 0    | logical and                                                                            |
| 0x1F | if            | 2    | if non-zero, jump to instruction at offset formed by $1$2                              |
| 0x20 | goto          | 2    | unconditional jump                                                                     |
| 0x21 | reg_store     | 1    | store a value in register $1                                                           |
| 0x22 | reg_load      | 1    | load a value from register $1                                                          |
| 0x23 | reg_inc       | 1    | increment integer in register $1                                                       |
| 0x24 | reg_dec       | 1    | decrement integer in register $1                                                       |
| 0x25 | data_move     | 2    | move data pointer by offset $1$2                                                       |
| 0x26 | data_goto     | 2    | move data pointer to address $1$2                                                      |
| 0x27 | data_push     | 1    | push (and move dptr) $1 bytes (zero-padded) from data pointer onto stack               |
| 0x28 | data_reg      | 2    | store (and move dptr) $1 bytes (zero-padded) from data pointer in register $2          |
| 0x29 | replace_byte  | 0    | pop stack value into byte at data pointer                                              |
| 0x2A | replace_short | 0    | pop stack value into short at data pointer                                             |
| 0x2B | buf_copy      | 1    | copy (and move dptr) popped number of bytes (zero-padded) into buffer $1               |
| 0x2C | buf_paste     | 1    | paste popped number of bytes (zero-padded) from buffer $1, overwriting existing bytes  |
| 0x2D | buf_prefix    | 1    | same as buffer_copy, but using the first two bytes to determine the length             |
| 0x2E | buf_rest      | 1    | copy from data pointer to end of script into buffer $1                                 |
| 0x2F | transfer      | 0    | move instruction pointer to data pointer                                               |
| 0x30 | reject        | 0    | reject input, terminating execution                                                    |
| 0x31 | add_sibling   | 1    | adds sibling; pushes success value                                                     |
| 0x32 | make_wallet   | 1    | adds a wallet with an initial balance and script                                       |
| 0x33 | send          | 0    | sends siacoins from host wallet to recipient                                           |
| 0x34 | verify        | 2    | verifies that a signature is valid; pushes bool                                        |
| 0x35 | switch        | 2    | if value and $1 are equal, branch to $2. The value is only consumed upon equality.     |
| 0x36 | if_move       | 2    | same as if, but with a relative, rather than absolute, address                         |
| 0x37 | move          | 2    | move instruction pointer by offset $1$2                                                |
| 0xFF | exit          | 0    | terminates execution                                                                   |
