#include "textflag.h"

DATA divid+0(SB)/4, $0x497a0000
GLOBL divid(SB), RODATA, $4

TEXT ·getMBSize(SB), NOSPLIT, $0-16
MOVSS x+0(FP), X1
MOVSS divid+0(SB), X0
DIVSS X0, X1
MOVSS X1, ret+8(FP)
RET
