List of bytecodes
-----------------

| Hex  |  Name  | Args |  Description |
|------|--------|------|--------------|
| 0x00 |  nop   |  0   |  do nothing |
| 0x01 |  push  |  1   |  push a value onto the stack |
| 0x02 |  pop   |  0   |  pop stack value, discarding it |
| 0x03 |  dup   |  0   |  duplicate stack value |
| 0x04 |  swap  |  0   |  swap two values on the stack |
| 0x05 |  add   |  0   |  add two stack values |
| 0x06 |  sub   |  0   |  subtract |
| 0x07 |  mul   |  0   |  multiply |
| 0x08 |  div   |  0   |  divide |
| 0x09 |  mod   |  0   |  modulus |
| 0x0A |  neg   |  0   |  negate |
| 0x0B |  eq    |  0   |  test for equality; pushes result onto stack |
| 0x0C |  ne    |  0   |  inequality |
| 0x0D |  lt    |  0   |  less than |
| 0x0E |  gt    |  0   |  greater than |
| 0x0F |  not   |  0   |  boolean negation |
| 0x10 |  or    |  0   |  boolean or |
| 0x11 |  and   |  0   |  boolean and |
| 0x12 |  if    |  1   |  if true, jump to instruction at arg |
| 0x13 |  goto  |  0   |  jump to instruction at arg |
| 0x14 |  store |  1   |  pop stack value into arg |
| 0x15 |  load  |  1   |  push arg value onto stack |
| 0x16 |  inc   |  1   |  increment variable |
| 0x17 |  dec   |  1   |  decrement variable |
| 0x18 |  asib  |  1   |  adds sibling defined in data block; pushes success value |
| 0xFF |  exit  |  0   |  terminates execution |
