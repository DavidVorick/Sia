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
| 0x07 |  addf  |  0   |  floating point addition |
| 0x08 |  subi  |  0   |  integer subtraction |
| 0x09 |  subf  |  0   |  floating point subtraction |
| 0x0A |  muli  |  0   |  integer multiplication |
| 0x0B |  mulf  |  0   |  floating point multiplication |
| 0x0C |  divi  |  0   |  integer division |
| 0x0D |  divf  |  0   |  floating point division |
| 0x0E |  modi  |  0   |  integer modulus |
| 0x0F |  negi  |  0   |  integer negatation |
| 0x10 |  negf  |  0   |  floating point negatation |
| 0x11 |  bor   |  0   |  integer binary or |
| 0x12 |  band  |  0   |  integer binary and |
| 0x13 |  bxor  |  0   |  integer binary xor |
| 0x14 |  shln  |  1   |  shift integer left by $1 bits |
| 0x15 |  shrn  |  1   |  shift integer right by $1 bits |
| 0x16 |  eq    |  0   |  test values for equality; push 1 (true) or 0 (false) |
| 0x17 |  ne    |  0   |  inequality |
| 0x18 |  lti   |  0   |  integer less than |
| 0x19 |  ltf   |  0   |  floating point less than |
| 0x1A |  gti   |  0   |  integer greater than |
| 0x1B |  gtf   |  0   |  floating point greater than |
| 0x1C |  lnot  |  0   |  logical negation |
| 0x1D |  lor   |  0   |  logical or |
| 0x1E |  land  |  0   |  logical and |
| 0x1F |  if    |  2   |  if true, jump to instruction at offset formed by ($1 << 8) + $2 |
| 0x20 |  goto  |  2   |  jump to instruction |
| 0x21 |  regs  |  1   |  store a value in register $1 |
| 0x22 |  regl  |  1   |  load a value from register $1 |
| 0x23 |  inci  |  1   |  increment integer in register $1 |
| 0x24 |  deci  |  1   |  decrement integer in register $1 |
| 0x25 |  blks  |  2   |  store a value in data block at address ($1 << 8) + $2 |
| 0x26 |  blkl  |  2   |  load a value from data block |
| 0x27 |  rej   |  0   |  reject input, terminating execution | 
| 0x28 |  xfer  |  0   |  transfer control to input | 
| 0x29 |  asib  |  2   |  adds sibling defined in data block; pushes success value |
| 0xFF |  exit  |  0   |  terminates execution |
