@0 = unnamed_addr constant [39 x i8] c"===== Welcome to the calculator =====\0A\00", align 1
@1 = unnamed_addr constant [80 x i8] c"What do you want to do?\0APlease enter the your choice, please select a number: \0A\00", align 1
@2 = unnamed_addr constant [29 x i8] c"Press '1' for sum operation\0A\00", align 1
@3 = unnamed_addr constant [34 x i8] c"Press '2' for subtract operation\0A\00", align 1
@4 = unnamed_addr constant [34 x i8] c"Press '3' for multiply operation\0A\00", align 1
@5 = unnamed_addr constant [34 x i8] c"Press '4' for division operation\0A\00", align 1
@6 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@7 = unnamed_addr constant [17 x i8] c"Invalid choice.\0A\00", align 1
@8 = unnamed_addr constant [15 x i8] c"First number: \00", align 1
@9 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@10 = unnamed_addr constant [16 x i8] c"Second number: \00", align 1
@11 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@12 = unnamed_addr constant [15 x i8] c"Sum result: %d\00", align 1
@13 = unnamed_addr constant [15 x i8] c"Sub result: %d\00", align 1
@14 = unnamed_addr constant [20 x i8] c"Multiply result: %d\00", align 1
@15 = unnamed_addr constant [20 x i8] c"Division result: %d\00", align 1

declare external i32 @r_runtime_printf(i8* %0, ...)

declare external i32 @r_runtime_scanf(i8* %0, ...)

define i32 @main() {
entry:
	%0 = alloca i32
	store i32 1, i32* %0
	%1 = alloca i32
	store i32 2, i32* %1
	%2 = alloca i32
	%3 = mul i32 -1, 1
	store i32 %3, i32* %2
	%4 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([39 x i8], [39 x i8]* @0, i8 0, i8 0))
	%5 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([80 x i8], [80 x i8]* @1, i8 0, i8 0))
	%6 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([29 x i8], [29 x i8]* @2, i8 0, i8 0))
	%7 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([34 x i8], [34 x i8]* @3, i8 0, i8 0))
	%8 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([34 x i8], [34 x i8]* @4, i8 0, i8 0))
	%9 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([34 x i8], [34 x i8]* @5, i8 0, i8 0))
	%10 = load i32, i32* %2
	%11 = call i32 (i8*, ...) @r_runtime_scanf(i8* getelementptr ([3 x i8], [3 x i8]* @6, i8 0, i8 0), i32* %2)
	%12 = load i32, i32* %2
	%13 = icmp sle i32 %12, 0
	%14 = load i32, i32* %2
	%15 = icmp sgt i32 %14, 4
	%16 = and i1 %13, %15
	br i1 %16, label %17, label %20

17:
	%18 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([17 x i8], [17 x i8]* @7, i8 0, i8 0))
	br label %19

19:
	ret i32 0

20:
	%21 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([15 x i8], [15 x i8]* @8, i8 0, i8 0))
	%22 = load i32, i32* %0
	%23 = call i32 (i8*, ...) @r_runtime_scanf(i8* getelementptr ([3 x i8], [3 x i8]* @9, i8 0, i8 0), i32* %0)
	%24 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([16 x i8], [16 x i8]* @10, i8 0, i8 0))
	%25 = load i32, i32* %1
	%26 = call i32 (i8*, ...) @r_runtime_scanf(i8* getelementptr ([3 x i8], [3 x i8]* @11, i8 0, i8 0), i32* %1)
	%27 = load i32, i32* %2
	%28 = icmp eq i32 %27, 1
	br i1 %28, label %29, label %35

29:
	%30 = load i32, i32* %0
	%31 = load i32, i32* %1
	%32 = add i32 %30, %31
	%33 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([15 x i8], [15 x i8]* @12, i8 0, i8 0), i32 %32)
	br label %34

34:
	br label %19

35:
	%36 = load i32, i32* %2
	%37 = icmp eq i32 %36, 2
	br i1 %37, label %38, label %44

38:
	%39 = load i32, i32* %0
	%40 = load i32, i32* %1
	%41 = sub i32 %39, %40
	%42 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([15 x i8], [15 x i8]* @13, i8 0, i8 0), i32 %41)
	br label %43

43:
	br label %34

44:
	%45 = load i32, i32* %2
	%46 = icmp eq i32 %45, 3
	br i1 %46, label %47, label %53

47:
	%48 = load i32, i32* %0
	%49 = load i32, i32* %1
	%50 = mul i32 %48, %49
	%51 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([20 x i8], [20 x i8]* @14, i8 0, i8 0), i32 %50)
	br label %52

52:
	br label %43

53:
	%54 = load i32, i32* %2
	%55 = icmp eq i32 %54, 4
	br i1 %55, label %56, label %61

56:
	%57 = load i32, i32* %0
	%58 = load i32, i32* %1
	%59 = udiv i32 %57, %58
	%60 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([20 x i8], [20 x i8]* @15, i8 0, i8 0), i32 %59)
	br label %61

61:
	br label %52
}
