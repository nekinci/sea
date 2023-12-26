@0 = unnamed_addr constant [4 x i8] c"%d\0A\00", align 1

declare external i32 @r_runtime_printf(i8* %0, ...)

declare external i32 @r_runtime_scanf(i8* %0, ...)

define i32 @main() {
entry:
	%0 = alloca i32
	store i32 0, i32* %0
	br label %2

1:
	ret i32 0

2:
	br label %9

3:
	%4 = load i32, i32* %0
	%5 = add i32 %4, 1
	store i32 %5, i32* %0
	%6 = load i32, i32* %0
	%7 = load i32, i32* %0
	%8 = icmp eq i32 %7, 3
	br i1 %8, label %10, label %11

9:
	br i1 true, label %3, label %1

10:
	br label %9

11:
	%12 = load i32, i32* %0
	%13 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([4 x i8], [4 x i8]* @0, i8 0, i8 0), i32 %12)
	%14 = load i32, i32* %0
	%15 = icmp eq i32 %14, 10
	br i1 %15, label %16, label %17

16:
	br label %1

17:
	br label %9
}
