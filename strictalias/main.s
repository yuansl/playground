	.file	"main.c"
	.text
	.p2align 4
	.globl	foo
	.type	foo, @function
foo:
.LFB23:
	.cfi_startproc
	endbr64
	movl	$1030792151, %eax
	ret
	.cfi_endproc
.LFE23:
	.size	foo, .-foo
	.p2align 4
	.globl	swap
	.type	swap, @function
swap:
.LFB24:
	.cfi_startproc
	endbr64
	movl	$3, %eax
	movw	%di, -2(%rsp)
	movw	%ax, -4(%rsp)
	movl	-4(%rsp), %eax
	ret
	.cfi_endproc
.LFE24:
	.size	swap, .-swap
	.p2align 4
	.globl	foo2
	.type	foo2, @function
foo2:
.LFB25:
	.cfi_startproc
	endbr64
	movzwl	(%rdi), %eax
	ret
	.cfi_endproc
.LFE25:
	.size	foo2, .-foo2
	.p2align 4
	.globl	accumulate
	.type	accumulate, @function
accumulate:
.LFB26:
	.cfi_startproc
	endbr64
	leal	(%rdi,%rdi,4), %edi
	leal	0(,%rdi,4), %eax
	ret
	.cfi_endproc
.LFE26:
	.size	accumulate, .-accumulate
	.p2align 4
	.globl	foo3
	.type	foo3, @function
foo3:
.LFB27:
	.cfi_startproc
	endbr64
	movq	.LC0(%rip), %rax
	movl	$12, a(%rip)
	movq	%rax, (%rdi)
	movl	a(%rip), %eax
	ret
	.cfi_endproc
.LFE27:
	.size	foo3, .-foo3
	.section	.rodata.str1.1,"aMS",@progbits,1
.LC1:
	.string	"x = %#x\n"
.LC2:
	.string	"x = %d\n"
.LC3:
	.string	"accumulate = %u\n"
.LC4:
	.string	"a = %d\n"
	.section	.text.startup,"ax",@progbits
	.p2align 4
	.globl	main
	.type	main, @function
main:
.LFB28:
	.cfi_startproc
	endbr64
	subq	$24, %rsp
	.cfi_def_cfa_offset 32
	movl	$50593795, %edx
	movl	$1, %edi
	movq	%fs:40, %rax
	movq	%rax, 8(%rsp)
	xorl	%eax, %eax
	leaq	.LC1(%rip), %rsi
	movl	$50593795, 4(%rsp)
	call	__printf_chk@PLT
	movl	$1030792151, %edx
	leaq	.LC2(%rip), %rsi
	xorl	%eax, %eax
	movl	$1, %edi
	call	__printf_chk@PLT
	movl	$15440, %edx
	leaq	.LC3(%rip), %rsi
	xorl	%eax, %eax
	movl	$1, %edi
	call	__printf_chk@PLT
	movl	$1958505087, %edx
	movl	$1, %edi
	movq	.LC0(%rip), %rax
	leaq	.LC4(%rip), %rsi
	movq	%rax, a(%rip)
	xorl	%eax, %eax
	call	__printf_chk@PLT
	movq	8(%rsp), %rax
	xorq	%fs:40, %rax
	jne	.L10
	xorl	%eax, %eax
	addq	$24, %rsp
	.cfi_remember_state
	.cfi_def_cfa_offset 8
	ret
.L10:
	.cfi_restore_state
	call	__stack_chk_fail@PLT
	.cfi_endproc
.LFE28:
	.size	main, .-main
	.globl	b
	.section	.data.rel.local,"aw"
	.align 8
	.type	b, @object
	.size	b, 8
b:
	.quad	a
	.globl	a
	.data
	.align 4
	.type	a, @object
	.size	a, 4
a:
	.long	5
	.section	.rodata.cst8,"aM",@progbits,8
	.align 8
.LC0:
	.long	1958505087
	.long	1070864531
	.ident	"GCC: (Ubuntu 9.3.0-17ubuntu1~20.04) 9.3.0"
	.section	.note.GNU-stack,"",@progbits
	.section	.note.gnu.property,"a"
	.align 8
	.long	 1f - 0f
	.long	 4f - 1f
	.long	 5
0:
	.string	 "GNU"
1:
	.align 8
	.long	 0xc0000002
	.long	 3f - 2f
2:
	.long	 0x3
3:
	.align 8
4:
