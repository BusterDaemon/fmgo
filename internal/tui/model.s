#include "textflag.h"

DATA divid+0(SB)/4, $0x497a0000 // 1024000
DATA divid+4(SB)/4, $0x44800000 // 1024
DATA divid+8(SB)/4, $0x3f800000 // 1
GLOBL divid(SB), RODATA, $12

TEXT ·getMBSize(SB), NOSPLIT, $0
    MOVQ $0, AX
    MOVSS x+0(FP), X1
    MOVSS divid+0(SB), X0
    DIVSS X0, X1
    MOVSS divid+8(SB), X0
    UCOMISS X1, X0
    JHI getKBs
    JG home

getKBs:
    MOVSS divid+4(SB), X0
    MULSS X1, X0
    MOVSS X0, X1
    MOVQ $1, AX
    JMP home

home:
    MOVSS X1, ret+8(FP)
    MOVQ AX, ret+16(FP)
    RET
