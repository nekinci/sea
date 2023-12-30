%string = type { i8*, i64 }

@0 = unnamed_addr constant [13 x i8] c"strlen = %d\0A\00", align 1
@1 = unnamed_addr constant [39 x i8] c"===== Welcome to the calculator =====\0A\00", align 1
@2 = unnamed_addr constant [56 x i8] c"What do you want to do?\0APlease enter the your choice: \0A\00", align 1
@3 = unnamed_addr constant [29 x i8] c"Press '1' for sum operation\0A\00", align 1
@4 = unnamed_addr constant [34 x i8] c"Press '2' for subtract operation\0A\00", align 1
@5 = unnamed_addr constant [34 x i8] c"Press '3' for multiply operation\0A\00", align 1
@6 = unnamed_addr constant [34 x i8] c"Press '4' for division operation\0A\00", align 1
@7 = unnamed_addr constant [29 x i8] c"Press -1 for exit operation\0A\00", align 1
@8 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@9 = unnamed_addr constant [11 x i8] c"Se\C3\A7im %d\0A\00", align 1
@10 = unnamed_addr constant [11 x i8] c"Good bye!\0A\00", align 1
@11 = unnamed_addr constant [17 x i8] c"Invalid choice.\0A\00", align 1
@12 = unnamed_addr constant [15 x i8] c"First number: \00", align 1
@13 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@14 = unnamed_addr constant [16 x i8] c"Second number: \00", align 1
@15 = unnamed_addr constant [3 x i8] c"%d\00", align 1
@16 = unnamed_addr constant [18 x i8] c"================\0A\00", align 1
@17 = unnamed_addr constant [16 x i8] c"Sum result: %d\0A\00", align 1
@18 = unnamed_addr constant [16 x i8] c"Sub result: %d\0A\00", align 1
@19 = unnamed_addr constant [21 x i8] c"Multiply result: %d\0A\00", align 1
@20 = unnamed_addr constant [21 x i8] c"Division result: %d\0A\00", align 1
@21 = unnamed_addr constant [18 x i8] c"================\0A\00", align 1

declare external i32 @r_runtime_printf(i8* %0, ...)

declare external i32 @r_runtime_scanf(i8* %0, ...)

declare external void @r_runtime_exit(i32 %0)

declare external i32 @printf_internal(%string %0, ...)

declare external %string @make_string(i8* %0)

declare external i32 @scanf_internal(%string %0, ...)

declare external i64 @strlen_internal(%string %0)

declare external i32 @sum(i32 %a, i32 %b)

define i32 @main() {
entry:
	%0 = alloca i1
	store i1 true, i1* %0
	%1 = alloca %string
	%2 = call %string @make_string(i8* getelementptr ([13 x i8], [13 x i8]* @0, i8 0, i8 0))
	store %string %2, %string* %1
	%3 = load %string, %string* %1
	%4 = load %string, %string* %1
	%5 = call i64 @strlen_internal(%string %4)
	%6 = call i32 (%string, ...) @printf_internal(%string %3, i64 %5)
	%7 = call %string @make_string(i8* getelementptr ([39 x i8], [39 x i8]* @1, i8 0, i8 0))
	%8 = call i32 (%string, ...) @printf_internal(%string %7)
	%9 = call %string @make_string(i8* getelementptr ([56 x i8], [56 x i8]* @2, i8 0, i8 0))
	%10 = call i32 (%string, ...) @printf_internal(%string %9)
	br label %initBlock

funcBlock:
	ret i32 0

initBlock:
	br label %condBlock

forBlock:
	%11 = alloca i32
	store i32 1, i32* %11
	%12 = alloca i32
	store i32 2, i32* %12
	%13 = alloca i32
	%14 = mul i32 -1, 1
	store i32 %14, i32* %13
	%15 = call %string @make_string(i8* getelementptr ([29 x i8], [29 x i8]* @3, i8 0, i8 0))
	%16 = call i32 (%string, ...) @printf_internal(%string %15)
	%17 = call %string @make_string(i8* getelementptr ([34 x i8], [34 x i8]* @4, i8 0, i8 0))
	%18 = call i32 (%string, ...) @printf_internal(%string %17)
	%19 = call %string @make_string(i8* getelementptr ([34 x i8], [34 x i8]* @5, i8 0, i8 0))
	%20 = call i32 (%string, ...) @printf_internal(%string %19)
	%21 = call %string @make_string(i8* getelementptr ([34 x i8], [34 x i8]* @6, i8 0, i8 0))
	%22 = call i32 (%string, ...) @printf_internal(%string %21)
	%23 = call %string @make_string(i8* getelementptr ([29 x i8], [29 x i8]* @7, i8 0, i8 0))
	%24 = call i32 (%string, ...) @printf_internal(%string %23)
	%25 = call %string @make_string(i8* getelementptr ([3 x i8], [3 x i8]* @8, i8 0, i8 0))
	%26 = load i32, i32* %13
	%27 = load i32, i32* %13
	%28 = call i32 (%string, ...) @scanf_internal(%string %25, i32* %13)
	%29 = call %string @make_string(i8* getelementptr ([11 x i8], [11 x i8]* @9, i8 0, i8 0))
	%30 = load i32, i32* %13
	%31 = call i32 (%string, ...) @printf_internal(%string %29, i32 %30)
	%32 = load i32, i32* %13
	%33 = mul i32 -1, 1
	%34 = icmp eq i32 %32, %33
	br i1 %34, label %37, label %40

condBlock:
	%35 = load i1, i1* %0
	%36 = and i1 %35, false
	br i1 %36, label %forBlock, label %funcBlock

StepBlock:
	br label %condBlock

37:
	%38 = call %string @make_string(i8* getelementptr ([11 x i8], [11 x i8]* @10, i8 0, i8 0))
	%39 = call i32 (%string, ...) @printf_internal(%string %38)
	call void @r_runtime_exit(i32 3)
	br label %40

40:
	%41 = load i32, i32* %13
	%42 = icmp sle i32 %41, 0
	%43 = load i32, i32* %13
	%44 = icmp sgt i32 %43, 4
	%45 = or i1 %42, %44
	br i1 %45, label %46, label %50

46:
	%47 = call %string @make_string(i8* getelementptr ([17 x i8], [17 x i8]* @11, i8 0, i8 0))
	%48 = call i32 (%string, ...) @printf_internal(%string %47)
	br label %49

49:
	br label %condBlock

50:
	%51 = call %string @make_string(i8* getelementptr ([15 x i8], [15 x i8]* @12, i8 0, i8 0))
	%52 = call i32 (%string, ...) @printf_internal(%string %51)
	%53 = call %string @make_string(i8* getelementptr ([3 x i8], [3 x i8]* @13, i8 0, i8 0))
	%54 = load i32, i32* %11
	%55 = load i32, i32* %11
	%56 = call i32 (%string, ...) @scanf_internal(%string %53, i32* %11)
	%57 = call %string @make_string(i8* getelementptr ([16 x i8], [16 x i8]* @14, i8 0, i8 0))
	%58 = call i32 (%string, ...) @printf_internal(%string %57)
	%59 = call %string @make_string(i8* getelementptr ([3 x i8], [3 x i8]* @15, i8 0, i8 0))
	%60 = load i32, i32* %12
	%61 = load i32, i32* %12
	%62 = call i32 (%string, ...) @scanf_internal(%string %59, i32* %12)
	%63 = call %string @make_string(i8* getelementptr ([18 x i8], [18 x i8]* @16, i8 0, i8 0))
	%64 = call i32 (%string, ...) @printf_internal(%string %63)
	%65 = load i32, i32* %13
	%66 = icmp eq i32 %65, 1
	br i1 %66, label %67, label %76

67:
	%68 = call %string @make_string(i8* getelementptr ([16 x i8], [16 x i8]* @17, i8 0, i8 0))
	%69 = load i32, i32* %11
	%70 = load i32, i32* %12
	%71 = call i32 @sum(i32 %69, i32 %70)
	%72 = call i32 (%string, ...) @printf_internal(%string %68, i32 %71)
	br label %73

73:
	%74 = call %string @make_string(i8* getelementptr ([18 x i8], [18 x i8]* @21, i8 0, i8 0))
	%75 = call i32 (%string, ...) @printf_internal(%string %74)
	br label %49

76:
	%77 = load i32, i32* %13
	%78 = icmp eq i32 %77, 2
	br i1 %78, label %79, label %86

79:
	%80 = call %string @make_string(i8* getelementptr ([16 x i8], [16 x i8]* @18, i8 0, i8 0))
	%81 = load i32, i32* %11
	%82 = load i32, i32* %12
	%83 = sub i32 %81, %82
	%84 = call i32 (%string, ...) @printf_internal(%string %80, i32 %83)
	br label %85

85:
	br label %73

86:
	%87 = load i32, i32* %13
	%88 = icmp eq i32 %87, 3
	br i1 %88, label %89, label %96

89:
	%90 = call %string @make_string(i8* getelementptr ([21 x i8], [21 x i8]* @19, i8 0, i8 0))
	%91 = load i32, i32* %11
	%92 = load i32, i32* %12
	%93 = mul i32 %91, %92
	%94 = call i32 (%string, ...) @printf_internal(%string %90, i32 %93)
	br label %95

95:
	br label %85

96:
	%97 = load i32, i32* %13
	%98 = icmp eq i32 %97, 4
	br i1 %98, label %99, label %105

99:
	%100 = call %string @make_string(i8* getelementptr ([21 x i8], [21 x i8]* @20, i8 0, i8 0))
	%101 = load i32, i32* %11
	%102 = load i32, i32* %12
	%103 = udiv i32 %101, %102
	%104 = call i32 (%string, ...) @printf_internal(%string %100, i32 %103)
	br label %105

105:
	br label %95
}
