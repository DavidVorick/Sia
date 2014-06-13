List of bytecodes
-----------------

| Hex  |  Name  | Args |  Description 
|------|--------|------|--------------
| 0x00 |  nop   |  0   |  do nothing |
| 0x01 |  pushb |  1   |  push a byte ($1) onto the stack |
| 0x02 |  pushs |  2   |  push a short ($1$2) onto the stack |
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
| 0x0F |  negi  |  0   |  integer negation |
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
| 0x1F |  if    |  2   |  if true, jump to instruction at offset formed by $1$2 |
| 0x20 |  goto  |  2   |  unconditional jump |
| 0x21 |  regs  |  1   |  store a value in register $1 |
| 0x22 |  regl  |  1   |  load a value from register $1 |
| 0x23 |  inci  |  1   |  increment integer in register $1 |
| 0x24 |  deci  |  1   |  decrement integer in register $1 |
| 0x25 |  dmov  |  2   |  move data pointer by offset $1$2 |
| 0x26 |  dgoto |  2   |  move data pointer to address $1$2 |
| 0x27 |  dpush |  1   |  push $1 bytes (zero-padded) from data pointer onto stack |
| 0x28 |  dregs |  2   |  store $1 bytes (zero-padded) from data pointer in register $2 |
| 0x29 |  repb  |  0   |  pop stack value into byte at data pointer |
| 0x2A |  reps  |  0   |  pop stack value into short at data pointer |
| 0x2B |  bufc  |  2   |  copy $1$2 bytes (zero-padded) from data pointer to buffer |
| 0x2C |  bufp  |  2   |  paste $1$2 bytes (zero-padded) from buffer to data pointer, overwriting existing bytes |
| 0x2D |  xfer  |  0   |  move instruction pointer to data pointer |
| 0x2E |  rej   |  0   |  reject input, terminating execution |
| 0x2F |  asib  |  0   |  adds sibling; pushes success value |
| 0x30 |  awall |  0   |  adds a wallet with an initial balance and script |
| 0x31 |  send  |  0   |  sends siacoins from host wallet to recipient |
| 0xFF |  exit  |  0   |  terminates execution |
