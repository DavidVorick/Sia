List of bytecodes
-----------------

| Hex  |  Name  | Args |  Description 
|------|--------|------|--------------
| 0x00 |  nop   |  0   |  do nothing |
| 0x01 |  pushb |  1   |  push a byte onto the stack |
| 0x02 |  pushs |  2   |  push a short onto the stack |
| 0x03 |  pop   |  0   |  pop a value, discarding it |
| 0x04 |  dup   |  0   |  duplicate a value |
| 0x05 |  swap  |  0   |  swap two values |
| 0x06 |  addi  |  0   |  integer addition |
| 0x07 |  subi  |  0   |  integer subtraction |
| 0x08 |  muli  |  0   |  integer multiplication |
| 0x09 |  divi  |  0   |  integer division |
| 0x0A |  modi  |  0   |  integer modulus |
| 0x0B |  negi  |  0   |  integer negatation |
| 0x0C |  bor   |  0   |  integer binary or |
| 0x0D |  band  |  0   |  integer binary and |
| 0x0E |  bxor  |  0   |  integer binary xor |
| 0x0F |  shln  |  1   |  shift integer left by $1 bits |
| 0x10 |  shrn  |  1   |  shift integer right by $1 bits |
| 0x11 |  eq    |  0   |  test values for equality; push 1 (true) or 0 (false) |
| 0x12 |  ne    |  0   |  inequality |
| 0x13 |  lti   |  0   |  integer less than |
| 0x14 |  gti   |  0   |  integer greater than |
| 0x15 |  lnot  |  0   |  logical negation |
| 0x16 |  lor   |  0   |  logical or |
| 0x17 |  land  |  0   |  logical and |
| 0x18 |  if    |  2   |  if true, jump to instruction at offset formed by ($1 << 8) + $2 |
| 0x19 |  goto  |  2   |  jump to instruction |
| 0x1A |  regs  |  1   |  store a value in register $1 |
| 0x1B |  regl  |  1   |  load a value from register $1 |
| 0x1C |  inci  |  1   |  increment integer in register $1 |
| 0x1D |  deci  |  1   |  decrement integer in register $1 |
| 0x1E |  blks  |  2   |  store a value in data block at address ($1 << 8) + $2 |
| 0x1F |  blkl  |  2   |  load a value from data block |
| 0x20 |  rej   |  0   |  reject input, terminating execution | 
| 0x21 |  asib  |  1   |  adds sibling defined in data block; pushes success value |
| 0xFF |  exit  |  0   |  terminates execution |
