#include "go_asm.h"
#include "textflag.h"

TEXT ·getg(SB), NOSPLIT, $0-8
    MOV    g, A0
    MOV    A0, ret+0(FP)
    RET
