; ModuleID = './runtime/runtime.c'
source_filename = "./runtime/runtime.c"
target datalayout = "e-m:o-i64:64-i128:128-n32:64-S128"
target triple = "arm64-apple-macosx14.0.0"

%struct.error = type { %struct.string, i32 }
%struct.string = type { i8*, i64, i64 }
%struct.Heap_Data = type { i64, i8*, i8 }
%struct.__sFILE = type { i8*, i32, i32, i16, i16, %struct.__sbuf, i32, i8*, i32 (i8*)*, i32 (i8*, i8*, i32)*, i64 (i8*, i64, i32)*, i32 (i8*, i8*, i32)*, %struct.__sbuf, %struct.__sFILEX*, i32, [3 x i8], [1 x i8], %struct.__sbuf, i32, i64 }
%struct.__sFILEX = type opaque
%struct.__sbuf = type { i8*, i32 }
%struct.slice = type { i8**, i64, i64 }

@env_index = global i32 0, align 4
@exception_index = global i32 0, align 4
@heap_data_index = global i64 0, align 8
@EXCEPTION_TABLE = global [100 x %struct.error*] zeroinitializer, align 8
@.str = private unnamed_addr constant [28 x i8] c"invalid exception index: %d\00", align 1
@env_stack = global [100 x [48 x i32]*] zeroinitializer, align 8
@.str.1 = private unnamed_addr constant [31 x i8] c"Null reference access error: \0A\00", align 1
@.str.2 = private unnamed_addr constant [58 x i8] c"Index out of bound error occurred: %d, slice size is: %zu\00", align 1
@heap_meta_data = global [640000 x %struct.Heap_Data*] zeroinitializer, align 8
@.str.3 = private unnamed_addr constant [21 x i8] c"is reachable: %p %p\0A\00", align 1
@.str.4 = private unnamed_addr constant [13 x i8] c"freeing %p \0A\00", align 1
@stack_start = global i64 0, align 8
@.str.5 = private unnamed_addr constant [2 x i8] c"r\00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"hello\00", align 1
@.str.7 = private unnamed_addr constant [4 x i8] c"abc\00", align 1
@.str.8 = private unnamed_addr constant [3 x i8] c"%d\00", align 1
@__stdoutp = external global %struct.__sFILE*, align 8
@.str.9 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.10 = private unnamed_addr constant [3 x i8] c"%f\00", align 1
@.str.11 = private unnamed_addr constant [2 x i8] c"\0A\00", align 1
@.str.12 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.13 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.14 = private unnamed_addr constant [18 x i8] c"Runtime exception\00", align 1
@.str.15 = private unnamed_addr constant [37 x i8] c"::Error code: %d, Error message: %s\0A\00", align 1
@.str.16 = private unnamed_addr constant [34 x i8] c"nil pointer dereference exception\00", align 1
@.str.17 = private unnamed_addr constant [25 x i8] c"floating point exception\00", align 1
@.str.18 = private unnamed_addr constant [20 x i8] c"illegal instruction\00", align 1
@.str.19 = private unnamed_addr constant [10 x i8] c"bus error\00", align 1
@.str.20 = private unnamed_addr constant [14 x i8] c"abort program\00", align 1
@.str.21 = private unnamed_addr constant [24 x i8] c"bad instruction sigtrap\00", align 1
@.str.22 = private unnamed_addr constant [17 x i8] c"sigempt received\00", align 1
@.str.23 = private unnamed_addr constant [16 x i8] c"bad system call\00", align 1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @____add__exception____(%struct.error* noundef %0) #0 {
  %2 = alloca %struct.error*, align 8
  store %struct.error* %0, %struct.error** %2, align 8
  %3 = load %struct.error*, %struct.error** %2, align 8
  %4 = load i32, i32* @exception_index, align 4
  %5 = add nsw i32 %4, 1
  store i32 %5, i32* @exception_index, align 4
  %6 = sext i32 %4 to i64
  %7 = getelementptr inbounds [100 x %struct.error*], [100 x %struct.error*]* @EXCEPTION_TABLE, i64 0, i64 %6
  store %struct.error* %3, %struct.error** %7, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define %struct.error* @____get__last__exception__instance____() #0 {
  %1 = alloca %struct.error*, align 8
  %2 = load i32, i32* @exception_index, align 4
  %3 = icmp sle i32 %2, 0
  br i1 %3, label %4, label %7

4:                                                ; preds = %0
  %5 = load i32, i32* @exception_index, align 4
  %6 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([28 x i8], [28 x i8]* @.str, i64 0, i64 0), i32 noundef %5)
  call void @exit(i32 noundef 14) #12
  unreachable

7:                                                ; preds = %0
  %8 = load i32, i32* @exception_index, align 4
  %9 = sub nsw i32 %8, 1
  %10 = sext i32 %9 to i64
  %11 = getelementptr inbounds [100 x %struct.error*], [100 x %struct.error*]* @EXCEPTION_TABLE, i64 0, i64 %10
  %12 = load %struct.error*, %struct.error** %11, align 8
  store %struct.error* %12, %struct.error** %1, align 8
  %13 = load i32, i32* @exception_index, align 4
  %14 = sub nsw i32 %13, 1
  store i32 %14, i32* @exception_index, align 4
  %15 = load %struct.error*, %struct.error** %1, align 8
  ret %struct.error* %15
}

declare i32 @printf(i8* noundef, ...) #1

; Function Attrs: noreturn
declare void @exit(i32 noundef) #2

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @____pop__exception__instance____() #0 {
  %1 = alloca i32, align 4
  %2 = load i32, i32* @exception_index, align 4
  %3 = icmp sgt i32 %2, 0
  br i1 %3, label %4, label %14

4:                                                ; preds = %0
  %5 = load i32, i32* @exception_index, align 4
  %6 = sub nsw i32 %5, 1
  store i32 %6, i32* %1, align 4
  %7 = load i32, i32* %1, align 4
  %8 = sext i32 %7 to i64
  %9 = getelementptr inbounds [100 x %struct.error*], [100 x %struct.error*]* @EXCEPTION_TABLE, i64 0, i64 %8
  %10 = load %struct.error*, %struct.error** %9, align 8
  %11 = bitcast %struct.error* %10 to i8*
  call void @free(i8* noundef %11)
  %12 = load i32, i32* @exception_index, align 4
  %13 = add nsw i32 %12, -1
  store i32 %13, i32* @exception_index, align 4
  br label %14

14:                                               ; preds = %4, %0
  ret void
}

declare void @free(i8* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define [48 x i32]* @____push_new_exception_env____() #0 {
  %1 = alloca [48 x i32]*, align 8
  %2 = call i8* @malloc(i64 noundef 192) #13
  %3 = bitcast i8* %2 to [48 x i32]*
  store [48 x i32]* %3, [48 x i32]** %1, align 8
  %4 = load [48 x i32]*, [48 x i32]** %1, align 8
  %5 = load i32, i32* @env_index, align 4
  %6 = sext i32 %5 to i64
  %7 = getelementptr inbounds [100 x [48 x i32]*], [100 x [48 x i32]*]* @env_stack, i64 0, i64 %6
  store [48 x i32]* %4, [48 x i32]** %7, align 8
  %8 = load i32, i32* @env_index, align 4
  %9 = add nsw i32 %8, 1
  store i32 %9, i32* @env_index, align 4
  %10 = load [48 x i32]*, [48 x i32]** %1, align 8
  ret [48 x i32]* %10
}

; Function Attrs: allocsize(0)
declare i8* @malloc(i64 noundef) #3

; Function Attrs: noinline nounwind optnone ssp uwtable
define [48 x i32]* @____get_last_exception_env____() #0 {
  %1 = load i32, i32* @env_index, align 4
  %2 = sub nsw i32 %1, 1
  %3 = sext i32 %2 to i64
  %4 = getelementptr inbounds [100 x [48 x i32]*], [100 x [48 x i32]*]* @env_stack, i64 0, i64 %3
  %5 = load [48 x i32]*, [48 x i32]** %4, align 8
  ret [48 x i32]* %5
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define [48 x i32]* @____pop_exception_env____() #0 {
  %1 = alloca [48 x i32]*, align 8
  %2 = alloca [48 x i32]*, align 8
  %3 = load i32, i32* @env_index, align 4
  %4 = icmp eq i32 %3, 0
  br i1 %4, label %5, label %10

5:                                                ; preds = %0
  %6 = load i32, i32* @env_index, align 4
  %7 = sext i32 %6 to i64
  %8 = getelementptr inbounds [100 x [48 x i32]*], [100 x [48 x i32]*]* @env_stack, i64 0, i64 %7
  %9 = load [48 x i32]*, [48 x i32]** %8, align 8
  store [48 x i32]* %9, [48 x i32]** %1, align 8
  br label %21

10:                                               ; preds = %0
  %11 = load i32, i32* @env_index, align 4
  %12 = sub nsw i32 %11, 1
  store i32 %12, i32* @env_index, align 4
  %13 = load i32, i32* @env_index, align 4
  %14 = sext i32 %13 to i64
  %15 = getelementptr inbounds [100 x [48 x i32]*], [100 x [48 x i32]*]* @env_stack, i64 0, i64 %14
  %16 = load [48 x i32]*, [48 x i32]** %15, align 8
  store [48 x i32]* %16, [48 x i32]** %2, align 8
  %17 = load i32, i32* @env_index, align 4
  %18 = sext i32 %17 to i64
  %19 = getelementptr inbounds [100 x [48 x i32]*], [100 x [48 x i32]*]* @env_stack, i64 0, i64 %18
  store [48 x i32]* null, [48 x i32]** %19, align 8
  %20 = load [48 x i32]*, [48 x i32]** %2, align 8
  store [48 x i32]* %20, [48 x i32]** %1, align 8
  br label %21

21:                                               ; preds = %10, %5
  %22 = load [48 x i32]*, [48 x i32]** %1, align 8
  ret [48 x i32]* %22
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @make_slice(%struct.slice* noalias sret(%struct.slice) align 8 %0) #0 {
  %2 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 2
  store i64 2, i64* %2, align 8
  %3 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  store i64 0, i64* %3, align 8
  %4 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 2
  %5 = load i64, i64* %4, align 8
  %6 = mul i64 8, %5
  %7 = call i8* @malloc(i64 noundef %6) #13
  %8 = bitcast i8* %7 to i8**
  %9 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 0
  store i8** %8, i8*** %9, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @append_slice_data(%struct.slice* noundef %0, i8* noundef %1) #0 {
  %3 = alloca %struct.slice*, align 8
  %4 = alloca i8*, align 8
  store %struct.slice* %0, %struct.slice** %3, align 8
  store i8* %1, i8** %4, align 8
  %5 = load %struct.slice*, %struct.slice** %3, align 8
  %6 = icmp eq %struct.slice* %5, null
  br i1 %6, label %7, label %9

7:                                                ; preds = %2
  %8 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([31 x i8], [31 x i8]* @.str.1, i64 0, i64 0))
  call void @exit(i32 noundef 1) #12
  unreachable

9:                                                ; preds = %2
  %10 = load %struct.slice*, %struct.slice** %3, align 8
  %11 = getelementptr inbounds %struct.slice, %struct.slice* %10, i32 0, i32 1
  %12 = load i64, i64* %11, align 8
  %13 = load %struct.slice*, %struct.slice** %3, align 8
  %14 = getelementptr inbounds %struct.slice, %struct.slice* %13, i32 0, i32 2
  %15 = load i64, i64* %14, align 8
  %16 = icmp uge i64 %12, %15
  br i1 %16, label %17, label %36

17:                                               ; preds = %9
  %18 = load %struct.slice*, %struct.slice** %3, align 8
  %19 = getelementptr inbounds %struct.slice, %struct.slice* %18, i32 0, i32 2
  %20 = load i64, i64* %19, align 8
  %21 = mul i64 %20, 2
  %22 = load %struct.slice*, %struct.slice** %3, align 8
  %23 = getelementptr inbounds %struct.slice, %struct.slice* %22, i32 0, i32 2
  store i64 %21, i64* %23, align 8
  %24 = load %struct.slice*, %struct.slice** %3, align 8
  %25 = getelementptr inbounds %struct.slice, %struct.slice* %24, i32 0, i32 0
  %26 = load i8**, i8*** %25, align 8
  %27 = bitcast i8** %26 to i8*
  %28 = load %struct.slice*, %struct.slice** %3, align 8
  %29 = getelementptr inbounds %struct.slice, %struct.slice* %28, i32 0, i32 2
  %30 = load i64, i64* %29, align 8
  %31 = mul i64 8, %30
  %32 = call i8* @realloc(i8* noundef %27, i64 noundef %31) #14
  %33 = bitcast i8* %32 to i8**
  %34 = load %struct.slice*, %struct.slice** %3, align 8
  %35 = getelementptr inbounds %struct.slice, %struct.slice* %34, i32 0, i32 0
  store i8** %33, i8*** %35, align 8
  br label %36

36:                                               ; preds = %17, %9
  %37 = load i8*, i8** %4, align 8
  %38 = load %struct.slice*, %struct.slice** %3, align 8
  %39 = getelementptr inbounds %struct.slice, %struct.slice* %38, i32 0, i32 0
  %40 = load i8**, i8*** %39, align 8
  %41 = load %struct.slice*, %struct.slice** %3, align 8
  %42 = getelementptr inbounds %struct.slice, %struct.slice* %41, i32 0, i32 1
  %43 = load i64, i64* %42, align 8
  %44 = getelementptr inbounds i8*, i8** %40, i64 %43
  store i8* %37, i8** %44, align 8
  %45 = load %struct.slice*, %struct.slice** %3, align 8
  %46 = getelementptr inbounds %struct.slice, %struct.slice* %45, i32 0, i32 1
  %47 = load i64, i64* %46, align 8
  %48 = add i64 %47, 1
  store i64 %48, i64* %46, align 8
  ret void
}

; Function Attrs: allocsize(1)
declare i8* @realloc(i8* noundef, i64 noundef) #4

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @append_slice_datap(%struct.slice* noundef %0, i8** noundef %1) #0 {
  %3 = alloca %struct.slice*, align 8
  %4 = alloca i8**, align 8
  store %struct.slice* %0, %struct.slice** %3, align 8
  store i8** %1, i8*** %4, align 8
  %5 = load %struct.slice*, %struct.slice** %3, align 8
  %6 = icmp eq %struct.slice* %5, null
  br i1 %6, label %7, label %9

7:                                                ; preds = %2
  %8 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([31 x i8], [31 x i8]* @.str.1, i64 0, i64 0))
  call void @exit(i32 noundef 1) #12
  unreachable

9:                                                ; preds = %2
  %10 = load %struct.slice*, %struct.slice** %3, align 8
  %11 = getelementptr inbounds %struct.slice, %struct.slice* %10, i32 0, i32 1
  %12 = load i64, i64* %11, align 8
  %13 = load %struct.slice*, %struct.slice** %3, align 8
  %14 = getelementptr inbounds %struct.slice, %struct.slice* %13, i32 0, i32 2
  %15 = load i64, i64* %14, align 8
  %16 = icmp uge i64 %12, %15
  br i1 %16, label %17, label %36

17:                                               ; preds = %9
  %18 = load %struct.slice*, %struct.slice** %3, align 8
  %19 = getelementptr inbounds %struct.slice, %struct.slice* %18, i32 0, i32 2
  %20 = load i64, i64* %19, align 8
  %21 = mul i64 %20, 2
  %22 = load %struct.slice*, %struct.slice** %3, align 8
  %23 = getelementptr inbounds %struct.slice, %struct.slice* %22, i32 0, i32 2
  store i64 %21, i64* %23, align 8
  %24 = load %struct.slice*, %struct.slice** %3, align 8
  %25 = getelementptr inbounds %struct.slice, %struct.slice* %24, i32 0, i32 0
  %26 = load i8**, i8*** %25, align 8
  %27 = bitcast i8** %26 to i8*
  %28 = load %struct.slice*, %struct.slice** %3, align 8
  %29 = getelementptr inbounds %struct.slice, %struct.slice* %28, i32 0, i32 2
  %30 = load i64, i64* %29, align 8
  %31 = mul i64 8, %30
  %32 = call i8* @realloc(i8* noundef %27, i64 noundef %31) #14
  %33 = bitcast i8* %32 to i8**
  %34 = load %struct.slice*, %struct.slice** %3, align 8
  %35 = getelementptr inbounds %struct.slice, %struct.slice* %34, i32 0, i32 0
  store i8** %33, i8*** %35, align 8
  br label %36

36:                                               ; preds = %17, %9
  %37 = load i8**, i8*** %4, align 8
  %38 = load i8*, i8** %37, align 8
  %39 = load %struct.slice*, %struct.slice** %3, align 8
  %40 = getelementptr inbounds %struct.slice, %struct.slice* %39, i32 0, i32 0
  %41 = load i8**, i8*** %40, align 8
  %42 = load %struct.slice*, %struct.slice** %3, align 8
  %43 = getelementptr inbounds %struct.slice, %struct.slice* %42, i32 0, i32 1
  %44 = load i64, i64* %43, align 8
  %45 = getelementptr inbounds i8*, i8** %41, i64 %44
  store i8* %38, i8** %45, align 8
  %46 = load %struct.slice*, %struct.slice** %3, align 8
  %47 = getelementptr inbounds %struct.slice, %struct.slice* %46, i32 0, i32 1
  %48 = load i64, i64* %47, align 8
  %49 = add i64 %48, 1
  store i64 %49, i64* %47, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i64 @len_slice(%struct.slice* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  %3 = load i64, i64* %2, align 8
  ret i64 %3
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i8* @access_slice_data(%struct.slice* noundef %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i8**, align 8
  store i32 %1, i32* %3, align 4
  %5 = load i32, i32* %3, align 4
  %6 = sext i32 %5 to i64
  %7 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  %8 = load i64, i64* %7, align 8
  %9 = icmp uge i64 %6, %8
  br i1 %9, label %10, label %15

10:                                               ; preds = %2
  %11 = load i32, i32* %3, align 4
  %12 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  %13 = load i64, i64* %12, align 8
  %14 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([58 x i8], [58 x i8]* @.str.2, i64 0, i64 0), i32 noundef %11, i64 noundef %13)
  call void @exit(i32 noundef 255) #12
  unreachable

15:                                               ; preds = %2
  %16 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 0
  %17 = load i8**, i8*** %16, align 8
  %18 = load i32, i32* %3, align 4
  %19 = sext i32 %18 to i64
  %20 = getelementptr inbounds i8*, i8** %17, i64 %19
  store i8** %20, i8*** %4, align 8
  %21 = load i8**, i8*** %4, align 8
  %22 = load i8*, i8** %21, align 8
  ret i8* %22
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i8* @access_slice_datap(%struct.slice* noundef %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i8**, align 8
  %5 = alloca i64*, align 8
  store i32 %1, i32* %3, align 4
  %6 = load i32, i32* %3, align 4
  %7 = sext i32 %6 to i64
  %8 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  %9 = load i64, i64* %8, align 8
  %10 = icmp uge i64 %7, %9
  br i1 %10, label %11, label %16

11:                                               ; preds = %2
  %12 = load i32, i32* %3, align 4
  %13 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 1
  %14 = load i64, i64* %13, align 8
  %15 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([58 x i8], [58 x i8]* @.str.2, i64 0, i64 0), i32 noundef %12, i64 noundef %14)
  call void @exit(i32 noundef 255) #12
  unreachable

16:                                               ; preds = %2
  %17 = getelementptr inbounds %struct.slice, %struct.slice* %0, i32 0, i32 0
  %18 = load i8**, i8*** %17, align 8
  %19 = load i32, i32* %3, align 4
  %20 = sext i32 %19 to i64
  %21 = getelementptr inbounds i8*, i8** %18, i64 %20
  store i8** %21, i8*** %4, align 8
  %22 = load i8**, i8*** %4, align 8
  %23 = load i8*, i8** %22, align 8
  %24 = bitcast i8* %23 to i64*
  store i64* %24, i64** %5, align 8
  %25 = load i8**, i8*** %4, align 8
  %26 = load i8*, i8** %25, align 8
  ret i8* %26
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i8* @memcpy_internal(i8* noundef %0, i8* noundef %1, i64 noundef %2) #0 {
  %4 = alloca i8*, align 8
  %5 = alloca i8*, align 8
  %6 = alloca i64, align 8
  store i8* %0, i8** %4, align 8
  store i8* %1, i8** %5, align 8
  store i64 %2, i64* %6, align 8
  %7 = load i8*, i8** %4, align 8
  %8 = load i8*, i8** %5, align 8
  %9 = load i64, i64* %6, align 8
  %10 = load i8*, i8** %4, align 8
  %11 = call i64 @llvm.objectsize.i64.p0i8(i8* %10, i1 false, i1 true, i1 false)
  %12 = call i8* @__memcpy_chk(i8* noundef %7, i8* noundef %8, i64 noundef %9, i64 noundef %11) #15
  ret i8* %12
}

; Function Attrs: nounwind
declare i8* @__memcpy_chk(i8* noundef, i8* noundef, i64 noundef, i64 noundef) #5

; Function Attrs: nofree nosync nounwind readnone speculatable willreturn
declare i64 @llvm.objectsize.i64.p0i8(i8*, i1 immarg, i1 immarg, i1 immarg) #6

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @make_string(%struct.string* noalias sret(%struct.string) align 8 %0, i8* noundef %1) #0 {
  %3 = alloca i8*, align 8
  %4 = alloca i64, align 8
  store i8* %1, i8** %3, align 8
  %5 = load i8*, i8** %3, align 8
  %6 = call i64 @strlen(i8* noundef %5)
  store i64 %6, i64* %4, align 8
  %7 = load i8*, i8** %3, align 8
  %8 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  store i8* %7, i8** %8, align 8
  %9 = load i64, i64* %4, align 8
  %10 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  store i64 %9, i64* %10, align 8
  %11 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 2
  store i64 51, i64* %11, align 8
  ret void
}

declare i64 @strlen(i8* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @printf_internal(i8* noundef %0, ...) #0 {
  %2 = alloca i8*, align 8
  %3 = alloca i8*, align 8
  %4 = alloca i32, align 4
  store i8* %0, i8** %2, align 8
  %5 = bitcast i8** %3 to i8*
  call void @llvm.va_start(i8* %5)
  %6 = load i8*, i8** %2, align 8
  %7 = load i8*, i8** %3, align 8
  %8 = call i32 @vprintf(i8* noundef %6, i8* noundef %7)
  store i32 %8, i32* %4, align 4
  %9 = bitcast i8** %3 to i8*
  call void @llvm.va_end(i8* %9)
  %10 = load i32, i32* %4, align 4
  ret i32 %10
}

; Function Attrs: nofree nosync nounwind willreturn
declare void @llvm.va_start(i8*) #7

declare i32 @vprintf(i8* noundef, i8* noundef) #1

; Function Attrs: nofree nosync nounwind willreturn
declare void @llvm.va_end(i8*) #7

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @scanf_internal(%struct.string* noundef %0, ...) #0 {
  %2 = alloca i8*, align 8
  %3 = alloca i32, align 4
  %4 = bitcast i8** %2 to i8*
  call void @llvm.va_start(i8* %4)
  %5 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  %6 = load i8*, i8** %5, align 8
  %7 = load i8*, i8** %2, align 8
  %8 = call i32 @vscanf(i8* noundef %6, i8* noundef %7)
  store i32 %8, i32* %3, align 4
  %9 = bitcast i8** %2 to i8*
  call void @llvm.va_end(i8* %9)
  %10 = load i32, i32* %3, align 4
  ret i32 %10
}

declare i32 @vscanf(i8* noundef, i8* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @push_heap_data(%struct.Heap_Data* noundef %0) #0 {
  %2 = alloca %struct.Heap_Data*, align 8
  store %struct.Heap_Data* %0, %struct.Heap_Data** %2, align 8
  %3 = load %struct.Heap_Data*, %struct.Heap_Data** %2, align 8
  %4 = load i64, i64* @heap_data_index, align 8
  %5 = add i64 %4, 1
  store i64 %5, i64* @heap_data_index, align 8
  %6 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %4
  store %struct.Heap_Data* %3, %struct.Heap_Data** %6, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @stop_the_world() #0 {
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @start_the_world() #0 {
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @gc_collect_start() #0 {
  %1 = call zeroext i1 @can_collectable()
  br i1 %1, label %2, label %3

2:                                                ; preds = %0
  call void @stop_the_world()
  call void @gc_collect()
  call void @start_the_world()
  br label %3

3:                                                ; preds = %2, %0
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @reset_heap_meta_data() #0 {
  %1 = alloca i32, align 4
  store i32 0, i32* %1, align 4
  br label %2

2:                                                ; preds = %13, %0
  %3 = load i32, i32* %1, align 4
  %4 = sext i32 %3 to i64
  %5 = load i64, i64* @heap_data_index, align 8
  %6 = icmp ult i64 %4, %5
  br i1 %6, label %7, label %16

7:                                                ; preds = %2
  %8 = load i32, i32* %1, align 4
  %9 = sext i32 %8 to i64
  %10 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %9
  %11 = load %struct.Heap_Data*, %struct.Heap_Data** %10, align 8
  %12 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %11, i32 0, i32 2
  store i8 0, i8* %12, align 8
  br label %13

13:                                               ; preds = %7
  %14 = load i32, i32* %1, align 4
  %15 = add nsw i32 %14, 1
  store i32 %15, i32* %1, align 4
  br label %2, !llvm.loop !9

16:                                               ; preds = %2
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @gc_mark(i8* noundef %0, i8* noundef %1) #0 {
  %3 = alloca i8*, align 8
  %4 = alloca i8*, align 8
  %5 = alloca i8*, align 8
  %6 = alloca i8*, align 8
  %7 = alloca i64, align 8
  %8 = alloca %struct.Heap_Data*, align 8
  %9 = alloca i8*, align 8
  store i8* %0, i8** %3, align 8
  store i8* %1, i8** %4, align 8
  %10 = load i8*, i8** %3, align 8
  store i8* %10, i8** %5, align 8
  %11 = load i8*, i8** %4, align 8
  store i8* %11, i8** %6, align 8
  br label %12

12:                                               ; preds = %61, %2
  %13 = load i8*, i8** %5, align 8
  %14 = load i8*, i8** %6, align 8
  %15 = getelementptr i8, i8* %14, i64 1
  %16 = icmp ult i8* %13, %15
  br i1 %16, label %17, label %64

17:                                               ; preds = %12
  store i64 0, i64* %7, align 8
  br label %18

18:                                               ; preds = %58, %17
  %19 = load i64, i64* %7, align 8
  %20 = load i64, i64* @heap_data_index, align 8
  %21 = icmp ult i64 %19, %20
  br i1 %21, label %22, label %61

22:                                               ; preds = %18
  %23 = load i64, i64* %7, align 8
  %24 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %23
  %25 = load %struct.Heap_Data*, %struct.Heap_Data** %24, align 8
  store %struct.Heap_Data* %25, %struct.Heap_Data** %8, align 8
  %26 = load %struct.Heap_Data*, %struct.Heap_Data** %8, align 8
  %27 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %26, i32 0, i32 1
  %28 = load i8*, i8** %27, align 8
  store i8* %28, i8** %9, align 8
  %29 = load %struct.Heap_Data*, %struct.Heap_Data** %8, align 8
  %30 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %29, i32 0, i32 2
  %31 = load i8, i8* %30, align 8
  %32 = trunc i8 %31 to i1
  br i1 %32, label %33, label %34

33:                                               ; preds = %22
  br label %58

34:                                               ; preds = %22
  %35 = load i8*, i8** %5, align 8
  %36 = bitcast i8* %35 to i64*
  %37 = load i64, i64* %36, align 8
  %38 = load i8*, i8** %9, align 8
  %39 = ptrtoint i8* %38 to i64
  %40 = icmp eq i64 %37, %39
  br i1 %40, label %41, label %57

41:                                               ; preds = %34
  %42 = load %struct.Heap_Data*, %struct.Heap_Data** %8, align 8
  %43 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %42, i32 0, i32 2
  store i8 1, i8* %43, align 8
  %44 = load %struct.Heap_Data*, %struct.Heap_Data** %8, align 8
  %45 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %44, i32 0, i32 1
  %46 = load i8*, i8** %45, align 8
  %47 = load i8*, i8** %5, align 8
  %48 = bitcast i8* %47 to i64*
  %49 = load i64, i64* %48, align 8
  %50 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([21 x i8], [21 x i8]* @.str.3, i64 0, i64 0), i8* noundef %46, i64 noundef %49)
  %51 = load i8*, i8** %9, align 8
  %52 = load i8*, i8** %9, align 8
  %53 = load %struct.Heap_Data*, %struct.Heap_Data** %8, align 8
  %54 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %53, i32 0, i32 0
  %55 = load i64, i64* %54, align 8
  %56 = getelementptr i8, i8* %52, i64 %55
  call void @gc_mark(i8* noundef %51, i8* noundef %56)
  br label %57

57:                                               ; preds = %41, %34
  br label %58

58:                                               ; preds = %57, %33
  %59 = load i64, i64* %7, align 8
  %60 = add i64 %59, 1
  store i64 %60, i64* %7, align 8
  br label %18, !llvm.loop !11

61:                                               ; preds = %18
  %62 = load i8*, i8** %5, align 8
  %63 = getelementptr inbounds i8, i8* %62, i32 1
  store i8* %63, i8** %5, align 8
  br label %12, !llvm.loop !12

64:                                               ; preds = %12
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @deallocate(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  %3 = alloca i64, align 8
  store i32 %0, i32* %2, align 4
  %4 = load i32, i32* %2, align 4
  %5 = sext i32 %4 to i64
  %6 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %5
  %7 = load %struct.Heap_Data*, %struct.Heap_Data** %6, align 8
  %8 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %7, i32 0, i32 1
  %9 = load i8*, i8** %8, align 8
  call void @free(i8* noundef %9)
  %10 = load i32, i32* %2, align 4
  %11 = sext i32 %10 to i64
  %12 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %11
  store %struct.Heap_Data* null, %struct.Heap_Data** %12, align 8
  %13 = load i32, i32* %2, align 4
  %14 = sext i32 %13 to i64
  store i64 %14, i64* %3, align 8
  br label %15

15:                                               ; preds = %26, %1
  %16 = load i64, i64* %3, align 8
  %17 = load i64, i64* @heap_data_index, align 8
  %18 = icmp ult i64 %16, %17
  br i1 %18, label %19, label %29

19:                                               ; preds = %15
  %20 = load i64, i64* %3, align 8
  %21 = add i64 %20, 1
  %22 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %21
  %23 = load %struct.Heap_Data*, %struct.Heap_Data** %22, align 8
  %24 = load i64, i64* %3, align 8
  %25 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %24
  store %struct.Heap_Data* %23, %struct.Heap_Data** %25, align 8
  br label %26

26:                                               ; preds = %19
  %27 = load i64, i64* %3, align 8
  %28 = add i64 %27, 1
  store i64 %28, i64* %3, align 8
  br label %15, !llvm.loop !13

29:                                               ; preds = %15
  %30 = load i64, i64* @heap_data_index, align 8
  %31 = add i64 %30, -1
  store i64 %31, i64* @heap_data_index, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @gc_sweep() #0 {
  %1 = alloca i64, align 8
  %2 = alloca i64, align 8
  %3 = alloca %struct.Heap_Data*, align 8
  store i64 0, i64* %1, align 8
  store i64 0, i64* %2, align 8
  br label %4

4:                                                ; preds = %25, %0
  %5 = load i64, i64* %2, align 8
  %6 = load i64, i64* @heap_data_index, align 8
  %7 = icmp ult i64 %5, %6
  br i1 %7, label %8, label %28

8:                                                ; preds = %4
  %9 = load i64, i64* %2, align 8
  %10 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %9
  %11 = load %struct.Heap_Data*, %struct.Heap_Data** %10, align 8
  store %struct.Heap_Data* %11, %struct.Heap_Data** %3, align 8
  %12 = load %struct.Heap_Data*, %struct.Heap_Data** %3, align 8
  %13 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %12, i32 0, i32 2
  %14 = load i8, i8* %13, align 8
  %15 = trunc i8 %14 to i1
  br i1 %15, label %24, label %16

16:                                               ; preds = %8
  %17 = load %struct.Heap_Data*, %struct.Heap_Data** %3, align 8
  %18 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %17, i32 0, i32 1
  %19 = load i8*, i8** %18, align 8
  %20 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([13 x i8], [13 x i8]* @.str.4, i64 0, i64 0), i8* noundef %19)
  %21 = load i64, i64* %2, align 8
  %22 = add i64 %21, -1
  store i64 %22, i64* %2, align 8
  %23 = trunc i64 %21 to i32
  call void @deallocate(i32 noundef %23)
  br label %24

24:                                               ; preds = %16, %8
  br label %25

25:                                               ; preds = %24
  %26 = load i64, i64* %2, align 8
  %27 = add i64 %26, 1
  store i64 %27, i64* %2, align 8
  br label %4, !llvm.loop !14

28:                                               ; preds = %4
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define zeroext i1 @can_collectable() #0 {
  ret i1 true
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define zeroext i1 @can_sweepable() #0 {
  ret i1 true
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @gc_collect() #0 {
  %1 = alloca i8*, align 8
  %2 = call i8* @llvm.frameaddress.p0i8(i32 0)
  store i8* %2, i8** %1, align 8
  %3 = load i8*, i8** %1, align 8
  %4 = load i64, i64* @stack_start, align 8
  %5 = inttoptr i64 %4 to i8*
  call void @gc_mark(i8* noundef %3, i8* noundef %5)
  %6 = call zeroext i1 @can_sweepable()
  br i1 %6, label %7, label %8

7:                                                ; preds = %0
  call void @gc_sweep()
  br label %8

8:                                                ; preds = %7, %0
  ret void
}

; Function Attrs: nofree nosync nounwind readnone willreturn
declare i8* @llvm.frameaddress.p0i8(i32 immarg) #8

; Function Attrs: noinline nounwind optnone ssp uwtable
define i8* @malloc_internal(i64 noundef %0) #0 {
  %2 = alloca i64, align 8
  %3 = alloca i64, align 8
  %4 = alloca i8*, align 8
  %5 = alloca %struct.Heap_Data*, align 8
  store i64 %0, i64* %2, align 8
  %6 = load i64, i64* %2, align 8
  %7 = add i64 %6, 24
  store i64 %7, i64* %3, align 8
  %8 = load i64, i64* %3, align 8
  %9 = call i8* @malloc(i64 noundef %8) #13
  store i8* %9, i8** %4, align 8
  %10 = load i8*, i8** %4, align 8
  %11 = load i64, i64* %2, align 8
  %12 = getelementptr i8, i8* %10, i64 %11
  %13 = bitcast i8* %12 to %struct.Heap_Data*
  store %struct.Heap_Data* %13, %struct.Heap_Data** %5, align 8
  %14 = load i8*, i8** %4, align 8
  %15 = load %struct.Heap_Data*, %struct.Heap_Data** %5, align 8
  %16 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %15, i32 0, i32 1
  store i8* %14, i8** %16, align 8
  %17 = load i64, i64* %2, align 8
  %18 = load %struct.Heap_Data*, %struct.Heap_Data** %5, align 8
  %19 = getelementptr inbounds %struct.Heap_Data, %struct.Heap_Data* %18, i32 0, i32 0
  store i64 %17, i64* %19, align 8
  %20 = load %struct.Heap_Data*, %struct.Heap_Data** %5, align 8
  %21 = load i64, i64* @heap_data_index, align 8
  %22 = add i64 %21, 1
  store i64 %22, i64* @heap_data_index, align 8
  %23 = getelementptr inbounds [640000 x %struct.Heap_Data*], [640000 x %struct.Heap_Data*]* @heap_meta_data, i64 0, i64 %21
  store %struct.Heap_Data* %20, %struct.Heap_Data** %23, align 8
  %24 = load i8*, i8** %4, align 8
  ret i8* %24
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i64 @strlen_internal(%struct.string* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  %3 = load i64, i64* %2, align 8
  ret i64 %3
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @cstr_append(%struct.string* noalias sret(%struct.string) align 8 %0, %struct.string* noundef %1, i8 noundef signext %2, i32* noundef %3) #0 {
  %5 = alloca i8, align 1
  %6 = alloca i32*, align 8
  store i8 %2, i8* %5, align 1
  store i32* %3, i32** %6, align 8
  %7 = load i32*, i32** %6, align 8
  %8 = load i32, i32* %7, align 4
  %9 = icmp eq i32 %8, 0
  br i1 %9, label %10, label %19

10:                                               ; preds = %4
  %11 = load i32*, i32** %6, align 8
  store i32 1, i32* %11, align 4
  %12 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %13 = load i8*, i8** %12, align 8
  %14 = load i32*, i32** %6, align 8
  %15 = load i32, i32* %14, align 4
  %16 = sext i32 %15 to i64
  %17 = call i8* @realloc(i8* noundef %13, i64 noundef %16) #14
  %18 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  store i8* %17, i8** %18, align 8
  br label %19

19:                                               ; preds = %10, %4
  %20 = load i32*, i32** %6, align 8
  %21 = load i32, i32* %20, align 4
  %22 = sext i32 %21 to i64
  %23 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %24 = load i64, i64* %23, align 8
  %25 = add i64 %24, 1
  %26 = icmp ule i64 %22, %25
  br i1 %26, label %27, label %38

27:                                               ; preds = %19
  %28 = load i32*, i32** %6, align 8
  %29 = load i32, i32* %28, align 4
  %30 = mul nsw i32 %29, 2
  store i32 %30, i32* %28, align 4
  %31 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %32 = load i8*, i8** %31, align 8
  %33 = load i32*, i32** %6, align 8
  %34 = load i32, i32* %33, align 4
  %35 = sext i32 %34 to i64
  %36 = call i8* @realloc(i8* noundef %32, i64 noundef %35) #14
  %37 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  store i8* %36, i8** %37, align 8
  br label %38

38:                                               ; preds = %27, %19
  %39 = load i8, i8* %5, align 1
  %40 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %41 = load i8*, i8** %40, align 8
  %42 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %43 = load i64, i64* %42, align 8
  %44 = getelementptr inbounds i8, i8* %41, i64 %43
  store i8 %39, i8* %44, align 1
  %45 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %46 = load i8*, i8** %45, align 8
  %47 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %48 = load i64, i64* %47, align 8
  %49 = add i64 %48, 1
  %50 = getelementptr inbounds i8, i8* %46, i64 %49
  store i8 10, i8* %50, align 1
  %51 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %52 = load i64, i64* %51, align 8
  %53 = add i64 %52, 1
  store i64 %53, i64* %51, align 8
  %54 = bitcast %struct.string* %0 to i8*
  %55 = bitcast %struct.string* %1 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %54, i8* align 8 %55, i64 24, i1 false)
  ret void
}

; Function Attrs: argmemonly nofree nounwind willreturn
declare void @llvm.memcpy.p0i8.p0i8.i64(i8* noalias nocapture writeonly, i8* noalias nocapture readonly, i64, i1 immarg) #9

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @open_file_read(%struct.string* noalias sret(%struct.string) align 8 %0, %struct.string* noundef %1) #0 {
  %3 = alloca %struct.__sFILE*, align 8
  %4 = alloca i32, align 4
  %5 = alloca i32, align 4
  %6 = alloca %struct.string, align 8
  %7 = alloca %struct.string, align 8
  %8 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %9 = load i8*, i8** %8, align 8
  %10 = call %struct.__sFILE* @"\01_fopen"(i8* noundef %9, i8* noundef getelementptr inbounds ([2 x i8], [2 x i8]* @.str.5, i64 0, i64 0))
  store %struct.__sFILE* %10, %struct.__sFILE** %3, align 8
  %11 = bitcast %struct.string* %0 to i8*
  call void @llvm.memset.p0i8.i64(i8* align 8 %11, i8 0, i64 24, i1 false)
  store i32 0, i32* %4, align 4
  br label %12

12:                                               ; preds = %2, %19
  %13 = load %struct.__sFILE*, %struct.__sFILE** %3, align 8
  %14 = call i32 @fgetc(%struct.__sFILE* noundef %13)
  store i32 %14, i32* %5, align 4
  %15 = load %struct.__sFILE*, %struct.__sFILE** %3, align 8
  %16 = call i32 @feof(%struct.__sFILE* noundef %15)
  %17 = icmp ne i32 %16, 0
  br i1 %17, label %18, label %19

18:                                               ; preds = %12
  br label %26

19:                                               ; preds = %12
  %20 = load i32, i32* %5, align 4
  %21 = trunc i32 %20 to i8
  %22 = bitcast %struct.string* %7 to i8*
  %23 = bitcast %struct.string* %0 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %22, i8* align 8 %23, i64 24, i1 false)
  call void @cstr_append(%struct.string* sret(%struct.string) align 8 %6, %struct.string* noundef %7, i8 noundef signext %21, i32* noundef %4)
  %24 = bitcast %struct.string* %0 to i8*
  %25 = bitcast %struct.string* %6 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %24, i8* align 8 %25, i64 24, i1 false)
  br label %12

26:                                               ; preds = %18
  ret void
}

declare %struct.__sFILE* @"\01_fopen"(i8* noundef, i8* noundef) #1

; Function Attrs: argmemonly nofree nounwind willreturn writeonly
declare void @llvm.memset.p0i8.i64(i8* nocapture writeonly, i8, i64, i1 immarg) #10

declare i32 @fgetc(%struct.__sFILE* noundef) #1

declare i32 @feof(%struct.__sFILE* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @add(%struct.string* noalias sret(%struct.string) align 8 %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  store i32 %1, i32* %3, align 4
  %4 = load i32, i32* %3, align 4
  %5 = icmp sgt i32 %4, 0
  br i1 %5, label %6, label %10

6:                                                ; preds = %2
  %7 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  store i8* getelementptr inbounds ([6 x i8], [6 x i8]* @.str.6, i64 0, i64 0), i8** %7, align 8
  %8 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  store i64 0, i64* %8, align 8
  %9 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 2
  store i64 0, i64* %9, align 8
  br label %14

10:                                               ; preds = %2
  %11 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  store i8* getelementptr inbounds ([4 x i8], [4 x i8]* @.str.7, i64 0, i64 0), i8** %11, align 8
  %12 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  store i64 0, i64* %12, align 8
  %13 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 2
  store i64 0, i64* %13, align 8
  br label %14

14:                                               ; preds = %10, %6
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @puts_int(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  store i32 %0, i32* %2, align 4
  %3 = load i32, i32* %2, align 4
  %4 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %3)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @puts_str(%struct.string* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  %3 = load i8*, i8** %2, align 8
  %4 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %5 = call i32 @"\01_fputs"(i8* noundef %3, %struct.__sFILE* noundef %4)
  ret i32 %5
}

declare i32 @"\01_fputs"(i8* noundef, %struct.__sFILE* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define i8* @to_char_pointer(%struct.string* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  %3 = load i8*, i8** %2, align 8
  ret i8* %3
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @compare_string(%struct.string* noundef %0, %struct.string* noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i32, align 4
  %5 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  %6 = load i64, i64* %5, align 8
  %7 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %8 = load i64, i64* %7, align 8
  %9 = icmp ne i64 %6, %8
  br i1 %9, label %10, label %11

10:                                               ; preds = %2
  store i32 0, i32* %3, align 4
  br label %20

11:                                               ; preds = %2
  %12 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  %13 = load i8*, i8** %12, align 8
  %14 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %15 = load i8*, i8** %14, align 8
  %16 = call i32 @strcmp(i8* noundef %13, i8* noundef %15)
  store i32 %16, i32* %4, align 4
  %17 = load i32, i32* %4, align 4
  %18 = icmp eq i32 %17, 0
  %19 = zext i1 %18 to i32
  store i32 %19, i32* %3, align 4
  br label %20

20:                                               ; preds = %11, %10
  %21 = load i32, i32* %3, align 4
  ret i32 %21
}

declare i32 @strcmp(i8* noundef, i8* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @concat_strings(%struct.string* noalias sret(%struct.string) align 8 %0, %struct.string* noundef %1, %struct.string* noundef %2) #0 {
  %4 = alloca i64, align 8
  %5 = alloca i8*, align 8
  %6 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %7 = load i64, i64* %6, align 8
  %8 = getelementptr inbounds %struct.string, %struct.string* %2, i32 0, i32 1
  %9 = load i64, i64* %8, align 8
  %10 = add i64 %7, %9
  store i64 %10, i64* %4, align 8
  %11 = load i64, i64* %4, align 8
  %12 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  store i64 %11, i64* %12, align 8
  %13 = load i64, i64* %4, align 8
  %14 = mul i64 1, %13
  %15 = call i8* @malloc(i64 noundef %14) #13
  store i8* %15, i8** %5, align 8
  %16 = load i8*, i8** %5, align 8
  %17 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 0
  %18 = load i8*, i8** %17, align 8
  %19 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %20 = load i64, i64* %19, align 8
  %21 = load i8*, i8** %5, align 8
  %22 = call i64 @llvm.objectsize.i64.p0i8(i8* %21, i1 false, i1 true, i1 false)
  %23 = call i8* @__memcpy_chk(i8* noundef %16, i8* noundef %18, i64 noundef %20, i64 noundef %22) #15
  %24 = load i8*, i8** %5, align 8
  %25 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %26 = load i64, i64* %25, align 8
  %27 = getelementptr inbounds i8, i8* %24, i64 %26
  %28 = getelementptr inbounds %struct.string, %struct.string* %2, i32 0, i32 0
  %29 = load i8*, i8** %28, align 8
  %30 = getelementptr inbounds %struct.string, %struct.string* %2, i32 0, i32 1
  %31 = load i64, i64* %30, align 8
  %32 = load i8*, i8** %5, align 8
  %33 = getelementptr inbounds %struct.string, %struct.string* %1, i32 0, i32 1
  %34 = load i64, i64* %33, align 8
  %35 = getelementptr inbounds i8, i8* %32, i64 %34
  %36 = call i64 @llvm.objectsize.i64.p0i8(i8* %35, i1 false, i1 true, i1 false)
  %37 = call i8* @__memcpy_chk(i8* noundef %27, i8* noundef %29, i64 noundef %31, i64 noundef %36) #15
  %38 = load i8*, i8** %5, align 8
  %39 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  store i8* %38, i8** %39, align 8
  %40 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 2
  store i64 100, i64* %40, align 8
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @concat_char_and_string(%struct.string* noalias sret(%struct.string) align 8 %0, i8 noundef signext %1, %struct.string* noundef %2) #0 {
  %4 = alloca i8, align 1
  %5 = alloca i8*, align 8
  %6 = alloca %struct.string, align 8
  %7 = alloca %struct.string, align 8
  %8 = alloca %struct.string, align 8
  store i8 %1, i8* %4, align 1
  %9 = call i8* @malloc(i64 noundef 1) #13
  store i8* %9, i8** %5, align 8
  %10 = load i8, i8* %4, align 1
  %11 = load i8*, i8** %5, align 8
  store i8 %10, i8* %11, align 1
  %12 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %6, i8* noundef %12)
  %13 = bitcast %struct.string* %7 to i8*
  %14 = bitcast %struct.string* %6 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %13, i8* align 8 %14, i64 24, i1 false)
  %15 = bitcast %struct.string* %8 to i8*
  %16 = bitcast %struct.string* %2 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %15, i8* align 8 %16, i64 24, i1 false)
  call void @concat_strings(%struct.string* sret(%struct.string) align 8 %0, %struct.string* noundef %7, %struct.string* noundef %8)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @concat_string_and_char(%struct.string* noalias sret(%struct.string) align 8 %0, %struct.string* noundef %1, i8 noundef signext %2) #0 {
  %4 = alloca i8, align 1
  %5 = alloca i8*, align 8
  %6 = alloca %struct.string, align 8
  %7 = alloca %struct.string, align 8
  %8 = alloca %struct.string, align 8
  store i8 %2, i8* %4, align 1
  %9 = call i8* @malloc(i64 noundef 1) #13
  store i8* %9, i8** %5, align 8
  %10 = load i8, i8* %4, align 1
  %11 = load i8*, i8** %5, align 8
  store i8 %10, i8* %11, align 1
  %12 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %6, i8* noundef %12)
  %13 = bitcast %struct.string* %7 to i8*
  %14 = bitcast %struct.string* %1 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %13, i8* align 8 %14, i64 24, i1 false)
  %15 = bitcast %struct.string* %8 to i8*
  %16 = bitcast %struct.string* %6 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %15, i8* align 8 %16, i64 24, i1 false)
  call void @concat_strings(%struct.string* sret(%struct.string) align 8 %0, %struct.string* noundef %7, %struct.string* noundef %8)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @concat_char_and_char(%struct.string* noalias sret(%struct.string) align 8 %0, i8 noundef signext %1, i8 noundef signext %2) #0 {
  %4 = alloca i8, align 1
  %5 = alloca i8, align 1
  %6 = alloca i8*, align 8
  %7 = alloca i8*, align 8
  %8 = alloca %struct.string, align 8
  %9 = alloca %struct.string, align 8
  store i8 %1, i8* %4, align 1
  store i8 %2, i8* %5, align 1
  %10 = call i8* @malloc(i64 noundef 1) #13
  store i8* %10, i8** %6, align 8
  %11 = load i8, i8* %4, align 1
  %12 = load i8*, i8** %6, align 8
  store i8 %11, i8* %12, align 1
  %13 = call i8* @malloc(i64 noundef 1) #13
  store i8* %13, i8** %7, align 8
  %14 = load i8, i8* %5, align 1
  %15 = load i8*, i8** %7, align 8
  store i8 %14, i8* %15, align 1
  %16 = load i8*, i8** %6, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %8, i8* noundef %16)
  %17 = load i8*, i8** %7, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %9, i8* noundef %17)
  call void @concat_strings(%struct.string* sret(%struct.string) align 8 %0, %struct.string* noundef %8, %struct.string* noundef %9)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @str_len(%struct.string* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  %3 = load i64, i64* %2, align 8
  %4 = trunc i64 %3 to i32
  ret i32 %4
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_str__(%struct.string* noundef %0) #0 {
  %2 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 1
  %3 = load i64, i64* %2, align 8
  %4 = icmp ugt i64 %3, 0
  br i1 %4, label %5, label %10

5:                                                ; preds = %1
  %6 = getelementptr inbounds %struct.string, %struct.string* %0, i32 0, i32 0
  %7 = load i8*, i8** %6, align 8
  %8 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %9 = call i32 @"\01_fputs"(i8* noundef %7, %struct.__sFILE* noundef %8)
  br label %10

10:                                               ; preds = %5, %1
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_char__(i8 noundef signext %0) #0 {
  %2 = alloca i8, align 1
  store i8 %0, i8* %2, align 1
  %3 = load i8, i8* %2, align 1
  %4 = sext i8 %3 to i32
  %5 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %6 = call i32 @fputc(i32 noundef %4, %struct.__sFILE* noundef %5)
  ret void
}

declare i32 @fputc(i32 noundef, %struct.__sFILE* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_i8__(i16 noundef signext %0) #0 {
  %2 = alloca i16, align 2
  store i16 %0, i16* %2, align 2
  %3 = load i16, i16* %2, align 2
  %4 = sext i16 %3 to i32
  %5 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %4)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_i16__(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  store i32 %0, i32* %2, align 4
  %3 = load i32, i32* %2, align 4
  %4 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %3)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_i32__(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  store i32 %0, i32* %2, align 4
  %3 = load i32, i32* %2, align 4
  %4 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %3)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_i64__(i64 noundef %0) #0 {
  %2 = alloca i64, align 8
  store i64 %0, i64* %2, align 8
  %3 = load i64, i64* %2, align 8
  %4 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([4 x i8], [4 x i8]* @.str.9, i64 0, i64 0), i64 noundef %3)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_f16__(float noundef %0) #0 {
  %2 = alloca float, align 4
  store float %0, float* %2, align 4
  %3 = load float, float* %2, align 4
  %4 = fpext float %3 to double
  %5 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %4)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_f32__(float noundef %0) #0 {
  %2 = alloca float, align 4
  store float %0, float* %2, align 4
  %3 = load float, float* %2, align 4
  %4 = fpext float %3 to double
  %5 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %4)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_f64__(double noundef %0) #0 {
  %2 = alloca double, align 8
  store double %0, double* %2, align 8
  %3 = load double, double* %2, align 8
  %4 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %3)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_charp__(i8* noundef %0) #0 {
  %2 = alloca i8*, align 8
  store i8* %0, i8** %2, align 8
  %3 = load i8*, i8** %2, align 8
  %4 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %5 = call i32 @"\01_fputs"(i8* noundef %3, %struct.__sFILE* noundef %4)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print_ln__() #0 {
  %1 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %2 = call i32 @"\01_fputs"(i8* noundef getelementptr inbounds ([2 x i8], [2 x i8]* @.str.11, i64 0, i64 0), %struct.__sFILE* noundef %1)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__print__bool__(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  store i32 %0, i32* %2, align 4
  %3 = load i32, i32* %2, align 4
  %4 = icmp eq i32 %3, 0
  br i1 %4, label %5, label %8

5:                                                ; preds = %1
  %6 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %7 = call i32 @"\01_fputs"(i8* noundef getelementptr inbounds ([6 x i8], [6 x i8]* @.str.12, i64 0, i64 0), %struct.__sFILE* noundef %6)
  br label %11

8:                                                ; preds = %1
  %9 = load %struct.__sFILE*, %struct.__sFILE** @__stdoutp, align 8
  %10 = call i32 @"\01_fputs"(i8* noundef getelementptr inbounds ([5 x i8], [5 x i8]* @.str.13, i64 0, i64 0), %struct.__sFILE* noundef %9)
  br label %11

11:                                               ; preds = %8, %5
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__float_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, float noundef %1) #0 {
  %3 = alloca float, align 4
  %4 = alloca i32, align 4
  %5 = alloca i8*, align 8
  store float %1, float* %3, align 4
  %6 = call i64 @llvm.objectsize.i64.p0i8(i8* null, i1 false, i1 true, i1 false)
  %7 = load float, float* %3, align 4
  %8 = fpext float %7 to double
  %9 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef null, i64 noundef 0, i32 noundef 0, i64 noundef %6, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %8)
  store i32 %9, i32* %4, align 4
  %10 = load i32, i32* %4, align 4
  %11 = add nsw i32 %10, 1
  %12 = sext i32 %11 to i64
  %13 = call i8* @malloc(i64 noundef %12) #13
  store i8* %13, i8** %5, align 8
  %14 = load i8*, i8** %5, align 8
  %15 = load i32, i32* %4, align 4
  %16 = add nsw i32 %15, 1
  %17 = sext i32 %16 to i64
  %18 = load i8*, i8** %5, align 8
  %19 = call i64 @llvm.objectsize.i64.p0i8(i8* %18, i1 false, i1 true, i1 false)
  %20 = load float, float* %3, align 4
  %21 = fpext float %20 to double
  %22 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef %14, i64 noundef %17, i32 noundef 0, i64 noundef %19, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %21)
  %23 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef %23)
  ret void
}

declare i32 @__snprintf_chk(i8* noundef, i64 noundef, i32 noundef, i64 noundef, i8* noundef, ...) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__double_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, double noundef %1) #0 {
  %3 = alloca double, align 8
  %4 = alloca i32, align 4
  %5 = alloca i8*, align 8
  store double %1, double* %3, align 8
  %6 = call i64 @llvm.objectsize.i64.p0i8(i8* null, i1 false, i1 true, i1 false)
  %7 = load double, double* %3, align 8
  %8 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef null, i64 noundef 0, i32 noundef 0, i64 noundef %6, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %7)
  store i32 %8, i32* %4, align 4
  %9 = load i32, i32* %4, align 4
  %10 = add nsw i32 %9, 1
  %11 = sext i32 %10 to i64
  %12 = call i8* @malloc(i64 noundef %11) #13
  store i8* %12, i8** %5, align 8
  %13 = load i8*, i8** %5, align 8
  %14 = load i32, i32* %4, align 4
  %15 = add nsw i32 %14, 1
  %16 = sext i32 %15 to i64
  %17 = load i8*, i8** %5, align 8
  %18 = call i64 @llvm.objectsize.i64.p0i8(i8* %17, i1 false, i1 true, i1 false)
  %19 = load double, double* %3, align 8
  %20 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef %13, i64 noundef %16, i32 noundef 0, i64 noundef %18, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.10, i64 0, i64 0), double noundef %19)
  %21 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef %21)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__i8_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, i16 noundef signext %1) #0 {
  %3 = alloca i16, align 2
  %4 = alloca i32, align 4
  %5 = alloca i8*, align 8
  store i16 %1, i16* %3, align 2
  %6 = call i64 @llvm.objectsize.i64.p0i8(i8* null, i1 false, i1 true, i1 false)
  %7 = load i16, i16* %3, align 2
  %8 = sext i16 %7 to i32
  %9 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef null, i64 noundef 0, i32 noundef 0, i64 noundef %6, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %8)
  store i32 %9, i32* %4, align 4
  %10 = load i32, i32* %4, align 4
  %11 = add nsw i32 %10, 1
  %12 = sext i32 %11 to i64
  %13 = call i8* @malloc(i64 noundef %12) #13
  store i8* %13, i8** %5, align 8
  %14 = load i8*, i8** %5, align 8
  %15 = load i32, i32* %4, align 4
  %16 = add nsw i32 %15, 1
  %17 = sext i32 %16 to i64
  %18 = load i8*, i8** %5, align 8
  %19 = call i64 @llvm.objectsize.i64.p0i8(i8* %18, i1 false, i1 true, i1 false)
  %20 = load i16, i16* %3, align 2
  %21 = sext i16 %20 to i32
  %22 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef %14, i64 noundef %17, i32 noundef 0, i64 noundef %19, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %21)
  %23 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef %23)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__i32_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i32, align 4
  %5 = alloca i8*, align 8
  store i32 %1, i32* %3, align 4
  %6 = call i64 @llvm.objectsize.i64.p0i8(i8* null, i1 false, i1 true, i1 false)
  %7 = load i32, i32* %3, align 4
  %8 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef null, i64 noundef 0, i32 noundef 0, i64 noundef %6, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %7)
  store i32 %8, i32* %4, align 4
  %9 = load i32, i32* %4, align 4
  %10 = add nsw i32 %9, 1
  %11 = sext i32 %10 to i64
  %12 = call i8* @malloc(i64 noundef %11) #13
  store i8* %12, i8** %5, align 8
  %13 = load i8*, i8** %5, align 8
  %14 = load i32, i32* %4, align 4
  %15 = add nsw i32 %14, 1
  %16 = sext i32 %15 to i64
  %17 = load i8*, i8** %5, align 8
  %18 = call i64 @llvm.objectsize.i64.p0i8(i8* %17, i1 false, i1 true, i1 false)
  %19 = load i32, i32* %3, align 4
  %20 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef %13, i64 noundef %16, i32 noundef 0, i64 noundef %18, i8* noundef getelementptr inbounds ([3 x i8], [3 x i8]* @.str.8, i64 0, i64 0), i32 noundef %19)
  %21 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef %21)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__i16_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  store i32 %1, i32* %3, align 4
  %4 = load i32, i32* %3, align 4
  call void @__i32_to_string__(%struct.string* sret(%struct.string) align 8 %0, i32 noundef %4)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__i64_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, i64 noundef %1) #0 {
  %3 = alloca i64, align 8
  %4 = alloca i32, align 4
  %5 = alloca i8*, align 8
  store i64 %1, i64* %3, align 8
  %6 = call i64 @llvm.objectsize.i64.p0i8(i8* null, i1 false, i1 true, i1 false)
  %7 = load i64, i64* %3, align 8
  %8 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef null, i64 noundef 0, i32 noundef 0, i64 noundef %6, i8* noundef getelementptr inbounds ([4 x i8], [4 x i8]* @.str.9, i64 0, i64 0), i64 noundef %7)
  store i32 %8, i32* %4, align 4
  %9 = load i32, i32* %4, align 4
  %10 = add nsw i32 %9, 1
  %11 = sext i32 %10 to i64
  %12 = call i8* @malloc(i64 noundef %11) #13
  store i8* %12, i8** %5, align 8
  %13 = load i8*, i8** %5, align 8
  %14 = load i32, i32* %4, align 4
  %15 = add nsw i32 %14, 1
  %16 = sext i32 %15 to i64
  %17 = load i8*, i8** %5, align 8
  %18 = call i64 @llvm.objectsize.i64.p0i8(i8* %17, i1 false, i1 true, i1 false)
  %19 = load i64, i64* %3, align 8
  %20 = call i32 (i8*, i64, i32, i64, i8*, ...) @__snprintf_chk(i8* noundef %13, i64 noundef %16, i32 noundef 0, i64 noundef %18, i8* noundef getelementptr inbounds ([4 x i8], [4 x i8]* @.str.9, i64 0, i64 0), i64 noundef %19)
  %21 = load i8*, i8** %5, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef %21)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__bool_to_string__(%struct.string* noalias sret(%struct.string) align 8 %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  store i32 %1, i32* %3, align 4
  %4 = load i32, i32* %3, align 4
  %5 = icmp eq i32 %4, 0
  br i1 %5, label %6, label %7

6:                                                ; preds = %2
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef getelementptr inbounds ([6 x i8], [6 x i8]* @.str.12, i64 0, i64 0))
  br label %8

7:                                                ; preds = %2
  call void @make_string(%struct.string* sret(%struct.string) align 8 %0, i8* noundef getelementptr inbounds ([5 x i8], [5 x i8]* @.str.13, i64 0, i64 0))
  br label %8

8:                                                ; preds = %7, %6
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @__get_argv_slice__(%struct.slice* noalias sret(%struct.slice) align 8 %0, i32 noundef %1, i8** noundef %2) #0 {
  %4 = alloca i32, align 4
  %5 = alloca i8**, align 8
  %6 = alloca %struct.slice*, align 8
  %7 = alloca %struct.slice, align 8
  %8 = alloca i32, align 4
  %9 = alloca %struct.string, align 8
  store i32 %1, i32* %4, align 4
  store i8** %2, i8*** %5, align 8
  %10 = call i8* @malloc(i64 noundef 24) #13
  %11 = bitcast i8* %10 to %struct.slice*
  store %struct.slice* %11, %struct.slice** %6, align 8
  %12 = load %struct.slice*, %struct.slice** %6, align 8
  call void @make_slice(%struct.slice* sret(%struct.slice) align 8 %7)
  %13 = bitcast %struct.slice* %12 to i8*
  %14 = bitcast %struct.slice* %7 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %13, i8* align 8 %14, i64 24, i1 false)
  store i32 0, i32* %8, align 4
  br label %15

15:                                               ; preds = %27, %3
  %16 = load i32, i32* %8, align 4
  %17 = load i32, i32* %4, align 4
  %18 = icmp slt i32 %16, %17
  br i1 %18, label %19, label %30

19:                                               ; preds = %15
  %20 = load i8**, i8*** %5, align 8
  %21 = load i32, i32* %8, align 4
  %22 = sext i32 %21 to i64
  %23 = getelementptr inbounds i8*, i8** %20, i64 %22
  %24 = load i8*, i8** %23, align 8
  call void @make_string(%struct.string* sret(%struct.string) align 8 %9, i8* noundef %24)
  %25 = load %struct.slice*, %struct.slice** %6, align 8
  %26 = bitcast %struct.string* %9 to i8*
  call void @append_slice_data(%struct.slice* noundef %25, i8* noundef %26)
  br label %27

27:                                               ; preds = %19
  %28 = load i32, i32* %8, align 4
  %29 = add nsw i32 %28, 1
  store i32 %29, i32* %8, align 4
  br label %15, !llvm.loop !15

30:                                               ; preds = %15
  %31 = load %struct.slice*, %struct.slice** %6, align 8
  %32 = bitcast %struct.slice* %0 to i8*
  %33 = bitcast %struct.slice* %31 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %32, i8* align 8 %33, i64 24, i1 false)
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @init() #0 {
  call void bitcast (void (...)* @____INIT____ to void ()*)()
  ret void
}

declare void @____INIT____(...) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @____handle__exception____() #0 {
  %1 = alloca %struct.error*, align 8
  %2 = alloca %struct.string, align 8
  %3 = load i32, i32* @exception_index, align 4
  %4 = sub nsw i32 %3, 1
  %5 = sext i32 %4 to i64
  %6 = getelementptr inbounds [100 x %struct.error*], [100 x %struct.error*]* @EXCEPTION_TABLE, i64 0, i64 %5
  %7 = load %struct.error*, %struct.error** %6, align 8
  store %struct.error* %7, %struct.error** %1, align 8
  %8 = load %struct.error*, %struct.error** %1, align 8
  %9 = icmp eq %struct.error* %8, null
  br i1 %9, label %10, label %12

10:                                               ; preds = %0
  %11 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([18 x i8], [18 x i8]* @.str.14, i64 0, i64 0))
  call void @exit(i32 noundef 255) #12
  unreachable

12:                                               ; preds = %0
  %13 = load %struct.error*, %struct.error** %1, align 8
  %14 = getelementptr inbounds %struct.error, %struct.error* %13, i32 0, i32 1
  %15 = load i32, i32* %14, align 8
  %16 = load %struct.error*, %struct.error** %1, align 8
  %17 = getelementptr inbounds %struct.error, %struct.error* %16, i32 0, i32 0
  %18 = bitcast %struct.string* %2 to i8*
  %19 = bitcast %struct.string* %17 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %18, i8* align 8 %19, i64 24, i1 false)
  %20 = call i8* @to_char_pointer(%struct.string* noundef %2)
  %21 = call i32 (i8*, ...) @printf(i8* noundef getelementptr inbounds ([37 x i8], [37 x i8]* @.str.15, i64 0, i64 0), i32 noundef %15, i8* noundef %20)
  %22 = load %struct.error*, %struct.error** %1, align 8
  %23 = getelementptr inbounds %struct.error, %struct.error* %22, i32 0, i32 1
  %24 = load i32, i32* %23, align 8
  call void @exit(i32 noundef %24) #12
  unreachable
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @handle_signal(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  %3 = alloca %struct.string, align 8
  %4 = alloca %struct.string, align 8
  %5 = alloca %struct.string, align 8
  %6 = alloca %struct.string, align 8
  %7 = alloca %struct.string, align 8
  %8 = alloca %struct.string, align 8
  %9 = alloca %struct.string, align 8
  %10 = alloca %struct.string, align 8
  %11 = alloca %struct.string, align 8
  %12 = alloca [48 x i32]*, align 8
  %13 = alloca %struct.error*, align 8
  %14 = alloca %struct.error, align 8
  store i32 %0, i32* %2, align 4
  %15 = load i32, i32* %2, align 4
  switch i32 %15, label %40 [
    i32 11, label %16
    i32 8, label %19
    i32 4, label %22
    i32 10, label %25
    i32 6, label %28
    i32 5, label %31
    i32 7, label %34
    i32 12, label %37
  ]

16:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %4, i8* noundef getelementptr inbounds ([34 x i8], [34 x i8]* @.str.16, i64 0, i64 0))
  %17 = bitcast %struct.string* %3 to i8*
  %18 = bitcast %struct.string* %4 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %17, i8* align 8 %18, i64 24, i1 false)
  br label %40

19:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %5, i8* noundef getelementptr inbounds ([25 x i8], [25 x i8]* @.str.17, i64 0, i64 0))
  %20 = bitcast %struct.string* %3 to i8*
  %21 = bitcast %struct.string* %5 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %20, i8* align 8 %21, i64 24, i1 false)
  br label %40

22:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %6, i8* noundef getelementptr inbounds ([20 x i8], [20 x i8]* @.str.18, i64 0, i64 0))
  %23 = bitcast %struct.string* %3 to i8*
  %24 = bitcast %struct.string* %6 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %23, i8* align 8 %24, i64 24, i1 false)
  br label %40

25:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %7, i8* noundef getelementptr inbounds ([10 x i8], [10 x i8]* @.str.19, i64 0, i64 0))
  %26 = bitcast %struct.string* %3 to i8*
  %27 = bitcast %struct.string* %7 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %26, i8* align 8 %27, i64 24, i1 false)
  br label %40

28:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %8, i8* noundef getelementptr inbounds ([14 x i8], [14 x i8]* @.str.20, i64 0, i64 0))
  %29 = bitcast %struct.string* %3 to i8*
  %30 = bitcast %struct.string* %8 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %29, i8* align 8 %30, i64 24, i1 false)
  br label %40

31:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %9, i8* noundef getelementptr inbounds ([24 x i8], [24 x i8]* @.str.21, i64 0, i64 0))
  %32 = bitcast %struct.string* %3 to i8*
  %33 = bitcast %struct.string* %9 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %32, i8* align 8 %33, i64 24, i1 false)
  br label %40

34:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %10, i8* noundef getelementptr inbounds ([17 x i8], [17 x i8]* @.str.22, i64 0, i64 0))
  %35 = bitcast %struct.string* %3 to i8*
  %36 = bitcast %struct.string* %10 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %35, i8* align 8 %36, i64 24, i1 false)
  br label %40

37:                                               ; preds = %1
  call void @make_string(%struct.string* sret(%struct.string) align 8 %11, i8* noundef getelementptr inbounds ([16 x i8], [16 x i8]* @.str.23, i64 0, i64 0))
  %38 = bitcast %struct.string* %3 to i8*
  %39 = bitcast %struct.string* %11 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %38, i8* align 8 %39, i64 24, i1 false)
  br label %40

40:                                               ; preds = %1, %37, %34, %31, %28, %25, %22, %19, %16
  %41 = call [48 x i32]* @____pop_exception_env____()
  store [48 x i32]* %41, [48 x i32]** %12, align 8
  %42 = call i8* @malloc_internal(i64 noundef 32)
  %43 = bitcast i8* %42 to %struct.error*
  store %struct.error* %43, %struct.error** %13, align 8
  %44 = load %struct.error*, %struct.error** %13, align 8
  %45 = getelementptr inbounds %struct.error, %struct.error* %14, i32 0, i32 0
  %46 = bitcast %struct.string* %45 to i8*
  %47 = bitcast %struct.string* %3 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %46, i8* align 8 %47, i64 24, i1 false)
  %48 = getelementptr inbounds %struct.error, %struct.error* %14, i32 0, i32 1
  %49 = load i32, i32* %2, align 4
  store i32 %49, i32* %48, align 8
  %50 = bitcast %struct.error* %44 to i8*
  %51 = bitcast %struct.error* %14 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* align 8 %50, i8* align 8 %51, i64 32, i1 false)
  %52 = load %struct.error*, %struct.error** %13, align 8
  call void @____add__exception____(%struct.error* noundef %52)
  call void @____handle__exception____()
  ret void
}

; Function Attrs: noinline nounwind optnone ssp uwtable
define void @____handle__runtime__signals____() #0 {
  %1 = call void (i32)* @signal(i32 noundef 11, void (i32)* noundef @handle_signal)
  %2 = call void (i32)* @signal(i32 noundef 8, void (i32)* noundef @handle_signal)
  %3 = call void (i32)* @signal(i32 noundef 4, void (i32)* noundef @handle_signal)
  %4 = call void (i32)* @signal(i32 noundef 10, void (i32)* noundef @handle_signal)
  %5 = call void (i32)* @signal(i32 noundef 6, void (i32)* noundef @handle_signal)
  %6 = call void (i32)* @signal(i32 noundef 5, void (i32)* noundef @handle_signal)
  %7 = call void (i32)* @signal(i32 noundef 7, void (i32)* noundef @handle_signal)
  %8 = call void (i32)* @signal(i32 noundef 12, void (i32)* noundef @handle_signal)
  ret void
}

declare void (i32)* @signal(i32 noundef, void (i32)* noundef) #1

; Function Attrs: noinline nounwind optnone ssp uwtable
define i32 @main(i32 noundef %0, i8** noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i32, align 4
  %5 = alloca i8**, align 8
  %6 = alloca i32*, align 8
  %7 = alloca [48 x i32]*, align 8
  %8 = alloca i32, align 4
  %9 = alloca %struct.slice, align 8
  store i32 0, i32* %3, align 4
  store i32 %0, i32* %4, align 4
  store i8** %1, i8*** %5, align 8
  store i32* null, i32** %6, align 8
  %10 = call i8* @llvm.frameaddress.p0i8(i32 0)
  %11 = ptrtoint i8* %10 to i64
  store i64 %11, i64* @stack_start, align 8
  %12 = call [48 x i32]* @____push_new_exception_env____()
  store [48 x i32]* %12, [48 x i32]** %7, align 8
  %13 = load [48 x i32]*, [48 x i32]** %7, align 8
  %14 = getelementptr inbounds [48 x i32], [48 x i32]* %13, i64 0, i64 0
  %15 = call i32 @setjmp(i32* noundef %14) #16
  store i32 %15, i32* %8, align 4
  call void @____handle__runtime__signals____()
  %16 = load i32, i32* @exception_index, align 4
  %17 = icmp ne i32 %16, 0
  br i1 %17, label %18, label %19

18:                                               ; preds = %2
  call void @____handle__exception____()
  br label %19

19:                                               ; preds = %18, %2
  call void @init()
  %20 = load i32, i32* %4, align 4
  %21 = load i32, i32* %4, align 4
  %22 = load i8**, i8*** %5, align 8
  call void @__get_argv_slice__(%struct.slice* sret(%struct.slice) align 8 %9, i32 noundef %21, i8** noundef %22)
  %23 = call i32 @__main__(i32 noundef %20, %struct.slice* noundef %9)
  ret i32 %23
}

; Function Attrs: returns_twice
declare i32 @setjmp(i32* noundef) #11

declare i32 @__main__(i32 noundef, %struct.slice* noundef) #1

attributes #0 = { noinline nounwind optnone ssp uwtable "frame-pointer"="non-leaf" "min-legal-vector-width"="0" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #1 = { "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #2 = { noreturn "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #3 = { allocsize(0) "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #4 = { allocsize(1) "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #5 = { nounwind "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #6 = { nofree nosync nounwind readnone speculatable willreturn }
attributes #7 = { nofree nosync nounwind willreturn }
attributes #8 = { nofree nosync nounwind readnone willreturn }
attributes #9 = { argmemonly nofree nounwind willreturn }
attributes #10 = { argmemonly nofree nounwind willreturn writeonly }
attributes #11 = { returns_twice "frame-pointer"="non-leaf" "no-trapping-math"="true" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+v8.5a,+zcm,+zcz" }
attributes #12 = { noreturn }
attributes #13 = { allocsize(0) }
attributes #14 = { allocsize(1) }
attributes #15 = { nounwind }
attributes #16 = { returns_twice }

!llvm.module.flags = !{!0, !1, !2, !3, !4, !5, !6, !7}
!llvm.ident = !{!8}

!0 = !{i32 1, !"wchar_size", i32 4}
!1 = !{i32 1, !"branch-target-enforcement", i32 0}
!2 = !{i32 1, !"sign-return-address", i32 0}
!3 = !{i32 1, !"sign-return-address-all", i32 0}
!4 = !{i32 1, !"sign-return-address-with-bkey", i32 0}
!5 = !{i32 7, !"PIC Level", i32 2}
!6 = !{i32 7, !"uwtable", i32 1}
!7 = !{i32 7, !"frame-pointer", i32 1}
!8 = !{!"Homebrew clang version 14.0.6"}
!9 = distinct !{!9, !10}
!10 = !{!"llvm.loop.mustprogress"}
!11 = distinct !{!11, !10}
!12 = distinct !{!12, !10}
!13 = distinct !{!13, !10}
!14 = distinct !{!14, !10}
!15 = distinct !{!15, !10}
