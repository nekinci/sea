#include<stdio.h>
#include<stdlib.h>
#include <stdarg.h>
#include<stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <signal.h>
#include <stdbool.h>
#include <errno.h>
#include <setjmp.h>
#include <stdint.h>

#ifndef EXCEPTION_TABLE_SIZE
#define EXCEPTION_TABLE_SIZE 100
#endif

#ifndef ENV_STACK_SIZE
#define ENV_STACK_SIZE 100
#endif

#ifndef HEAP_METADATA_CAP
#define HEAP_METADATA_CAP 640000
#endif

#pragma clang diagnostic ignored "-Wvarargs"

uintptr_t stack_start;

typedef struct Heap_Data Heap_Data;
typedef struct error error;

void gc_collect();
void gc_mark(void* end, void* start);
void start_the_world();
void stop_the_world();
void gc_sweep();
bool can_collectable();
bool can_sweepable();

error* EXCEPTION_TABLE[EXCEPTION_TABLE_SIZE];
jmp_buf* env_stack[ENV_STACK_SIZE];
int env_index = 0;
int exception_index = 0;
size_t heap_data_index = 0;

typedef struct {
    char* buffer;
    size_t size;
    size_t cap;
} string;

typedef struct {
   char** data_list;
   size_t size;
   size_t cap;
} slice;

struct error {
    string message;
    int error_code;
    // maybe stacktrace vs.
};

struct Heap_Data {
    size_t size;
    void* start;
    bool is_reachable;
} ;

Heap_Data* heap_meta_data[HEAP_METADATA_CAP];


void ____add__exception____(error* err) {
    EXCEPTION_TABLE[exception_index++] = err;
}

error* ____get__last__exception__instance____() {
    if (exception_index <= 0) {
        printf("invalid exception index: %d", exception_index);
        exit(14);
    }
    error* instance = EXCEPTION_TABLE[exception_index-1];
    exception_index -= 1;
    return instance;
}

void ____pop__exception__instance____() {
    if (exception_index > 0) {
        int idx = exception_index - 1;
        free(EXCEPTION_TABLE[idx]);
        exception_index--;
    }
}

jmp_buf* ____push_new_exception_env____() {
    jmp_buf* env = malloc(sizeof(jmp_buf));
    env_stack[env_index] = env;
    env_index+= 1;
    return env;
}

jmp_buf* ____get_last_exception_env____() {
    return env_stack[env_index - 1];
}


jmp_buf* ____pop_exception_env____() {
    if (env_index == 0) {
        return env_stack[env_index];
    }
    env_index -= 1;
    jmp_buf* popped = env_stack[env_index];
    env_stack[env_index] = NULL;
    return popped;
}

slice make_slice() {
    slice s;
    s.cap = 2;
    s.size = 0;
    s.data_list = malloc(sizeof(char**) * s.cap);
    return s;
}


void append_slice_data(slice* s, char* data) {
    if (s == NULL) {
        printf("Null reference access error: \n");
        exit(1);
    }
    if (s -> size >= s -> cap) {
        s -> cap = s -> cap * 2;
        s -> data_list = realloc(s -> data_list, sizeof(char**) * s -> cap);
    }

    *(s -> data_list + s -> size) = data;
    s -> size++;
}

void append_slice_datap(slice* s, char** data) {
    if (s == NULL) {
        printf("Null reference access error: \n");
        exit(1);
    }


    if (s -> size >= s -> cap) {
        s -> cap = s -> cap * 2;
        s -> data_list = realloc(s -> data_list, sizeof(char**) * s -> cap);
    }

    *(s -> data_list + s -> size) = *data;
    s -> size++;
}

long len_slice(slice s) {
    return s.size;
}

char* access_slice_data(slice s, int index) {
    if (index >= s.size) {
        printf("Index out of bound error occurred: %d, slice size is: %zu", index, s.size);
        exit(255);
    }
    char** data = (char**)(s.data_list + index);
    return *data;
}

char* access_slice_datap(slice s, int index) {
    if (index >= s.size) {
        printf("Index out of bound error occurred: %d, slice size is: %zu", index, s.size);
        exit(255);
    }

    char **data = (char**)(s.data_list + index);
    long* t = (long*)(*data);
    return *data;
}

void* memcpy_internal(void* dest, const void* src, size_t sizeinbytes) {
    return memcpy(dest, src, sizeinbytes);
}

string make_string(char* buffer) {
    size_t size = strlen(buffer);
    string result;
    result.buffer = buffer;
    result.size = size;
    result.cap = 51;
    return result;
}

int printf_internal(const char* str, ...) {
    va_list args;
    va_start(args, str);
    int r = vprintf(str, args);
    va_end(args);
    return r;
}

int scanf_internal(string s, ...) {
     va_list args;
     va_start(args, s.buffer);
     int r = vscanf(s.buffer, args);
     va_end(args);
     return r;
}

void push_heap_data(Heap_Data* heap_data) {
    heap_meta_data[heap_data_index++] = heap_data;
}

void stop_the_world() {
    // TODO implement that when the multi-threading implemented
}

void start_the_world() {
    // TODO implement that when the multi-threading implemented
}

void gc_collect_start() {
    if (can_collectable()) {
        stop_the_world();
            gc_collect();
        start_the_world();
    }
}

void reset_heap_meta_data() {
    for (int i = 0; i < heap_data_index; ++i) {
        heap_meta_data[i] -> is_reachable = false;
    }
}

void gc_mark(void* stack_end_arg, void *stack_start_arg) {

    void* stack_end = stack_end_arg;
    void* stack_start = stack_start_arg;
    while (stack_end < stack_start + 1) {
        for (size_t i = 0; i < heap_data_index; i++) {
            Heap_Data* heap_data = (Heap_Data*) heap_meta_data[i];
            void *start = heap_data -> start;
            if (heap_data -> is_reachable) continue;

            if (*(uintptr_t*)stack_end == (uintptr_t)start) {
                heap_data -> is_reachable = true;
                printf("is reachable: %p %p\n", heap_data -> start, *(uintptr_t*)stack_end);
                gc_mark((void*)start, (void*)start + heap_data -> size);
            }
        }
        stack_end++;
    }
}

void deallocate(int index) {
    free(heap_meta_data[index] -> start);
    heap_meta_data[index] = NULL;
    for (size_t i = index; i < heap_data_index; ++i) {
        heap_meta_data[i] = heap_meta_data[i + 1];
    }
    heap_data_index--;
}

void gc_sweep() {
    size_t to_be_removed = 0;
    for (size_t i = 0; i < heap_data_index; ++i) {
        Heap_Data* heap_data = heap_meta_data[i];
        if (!heap_data -> is_reachable) {
            printf("freeing %p \n", heap_data -> start);
            deallocate(i--);
        }
    }

}


bool can_collectable() {
    return true;
}

bool can_sweepable() {
    return true;
}

void gc_collect() {
    reset_heap_meta_data();
    void* stack_end = (void*) __builtin_frame_address(0);
    gc_mark(stack_end, (void*)stack_start);
    if (can_sweepable()) {
        gc_sweep();
    }
}

void* malloc_internal(size_t s) {
    size_t size_in_bytes = s + sizeof(Heap_Data);
    void* allocated = malloc(size_in_bytes);
    Heap_Data* data = (Heap_Data*) (allocated + s);
    data -> start = allocated;
    data -> size = s;
    heap_meta_data[heap_data_index++] = data;
    return allocated;
}

size_t strlen_internal(string str) {
    return str.size;
}

string cstr_append(string s, char c, int* cap) {
  if (*cap == 0) {
    *cap = 1;
    s.buffer = realloc(s.buffer, *cap);
  }

  if (*cap <= s.size + 1) {
    *cap *= 2;
    s.buffer = realloc(s.buffer, *cap);
  }


  s.buffer[s.size] = c;
  s.buffer[s.size+1] = '\n';
  s.size += 1;
  return s;
}

// TODO it is temporarily, change it.
string open_file_read(string path) {
   FILE *file = fopen(path.buffer, "r");
     string data = {.buffer = NULL, .size = 0};
     int cap = 0;
     while (1) {
       int c = fgetc(file);
       if (feof(file)) {
         break;
       }
       data = cstr_append(data, (char)c, &cap);
     }

     return data;
}

string add(int a) {
    if (a > 0) {
        return (string) {.buffer = "hello"};
    } else {
        return (string) {.buffer = "abc"};
    }
}

void puts_int(int a) {
    printf("%d", a);
}

int puts_str(string s) {
    return fputs(s.buffer, stdout);
}


char* to_char_pointer(string s) {
    return s.buffer;
}

int compare_string(string a, string b) {
    if (a.size != b.size) return 0;
    int res = strcmp(a.buffer, b.buffer);
    return res == 0;
}

string concat_strings(string a, string b) {

   // printf("a: %s %zu, b: %s %zu\n", a.buffer, a.size, b.buffer, b.size);
    size_t len = a.size + b.size;
    string res;
    res.size = len;
    char* newStr = malloc(sizeof(char) * len);
    memcpy(newStr, a.buffer, a.size);
    memcpy(newStr + a.size, b.buffer, b.size);
    res.buffer = newStr;

    res.cap = 100; // TODO
    return res;
}

string concat_char_and_string(char c, string second) {
    char* cbuff = malloc(sizeof(char));
    *cbuff = c;
    string first = make_string(cbuff);
    return concat_strings(first, second);
}

string concat_string_and_char(string first, char c) {
    char* cbuff = malloc(sizeof(char));
    *cbuff = c;
    string second = make_string(cbuff);
    return concat_strings(first, second);
}

string concat_char_and_char(char first, char second) {
    char* fbuff = malloc(sizeof(char));
    *fbuff = first;
    char* sbuff = malloc(sizeof(char));
    *sbuff = second;

    return concat_strings(make_string(fbuff), make_string(sbuff));
}

int str_len(string str) {
    return str.size;
}

void __print_str__(string s) {
    if (s.size > 0) {
        fputs(s.buffer, stdout);
    }
}

void __print_char__(char c) {
    fputc(c, stdout);
}

void __print_i8__(short s) {
    printf("%d", s);
}

void __print_i16__(int s) {
    printf("%d", s);
}

void __print_i32__(int s) {
    printf("%d", s);
}

void __print_i64__(long s) {
    printf("%ld", s);
}

void __print_f16__(float f) {
    printf("%f", f);
}

void __print_f32__(float f) {
    printf("%f", f);
}

void __print_f64__(double f) {
    printf("%f", f);
}

void __print_charp__(const char* buffer) {
    fputs(buffer, stdout);
}

void __print_ln__() {
    fputs("\n", stdout);
}

void __print__bool__(int b) {
    if (b == 0) {
        fputs("false", stdout);
    } else {
        fputs("true", stdout);
    }
}

string __float_to_string__(float f) {
    int len = snprintf(NULL, 0, "%f", f);
    char *result = malloc(len + 1);
    snprintf(result, len + 1, "%f", f);
    return make_string(result);
}

string __double_to_string__(double d) {
    int len = snprintf(NULL, 0, "%f", d);
    char *result = malloc(len + 1);
    snprintf(result, len + 1, "%f", d);
    return make_string(result);
}

string __i8_to_string__(short s) {
    int len = snprintf(NULL, 0, "%d", s);
    char *result = malloc(len + 1);
    snprintf(result, len + 1, "%d", s);
    return make_string(result);
}


string __i32_to_string__(int s) {
    int len = snprintf(NULL, 0, "%d", s);
    char *result = malloc(len + 1);
    snprintf(result, len + 1, "%d", s);
    return make_string(result);
}

string __i16_to_string__(int s) {
   return __i32_to_string__(s);
}


string __i64_to_string__(long l) {
    int len = snprintf(NULL, 0, "%ld", l);
    char *result = malloc(len + 1);
    snprintf(result, len + 1, "%ld", l);
    return make_string(result);
}
string __bool_to_string__(int i) {
    if (i == 0) {
        return make_string("false");
    } else {
        return make_string("true");
    }
}


slice __get_argv_slice__(int argc, char** argv) {
    slice* s = malloc(sizeof(slice));
    *s = make_slice();
    for (int i = 0; i < argc; i++) {
        string arg = make_string(argv[i]);
        append_slice_data(s, (char*)&arg);
    }
    return *s;
}

// __init__()
extern void ____INIT____();
void init() {
    ____INIT____();
}

void ____handle__exception____() {
    error* err = EXCEPTION_TABLE[exception_index - 1];
    if (err == NULL) {
        printf("Runtime exception");
        exit(255);
    } else {
        printf("::Error code: %d, Error message: %s\n", err -> error_code, to_char_pointer(err -> message));
        exit(err -> error_code);
    }
}

void handle_signal(int sig_code) {
    string error_message;
    switch (sig_code) {
       case SIGSEGV: {
            error_message = make_string("nil pointer dereference exception");
       }; break;
       case SIGFPE: {
            error_message = make_string("floating point exception");
       }; break;
       case SIGILL: {
            error_message = make_string("illegal instruction");
       }; break;
       case SIGBUS: {
            error_message = make_string("bus error");
       }; break;
       case SIGABRT: {
            error_message = make_string("abort program");
       }; break;
       case SIGTRAP: {
            error_message = make_string("bad instruction sigtrap");
       }; break;
       case SIGEMT: {
            error_message = make_string("sigempt received");
       }; break;
       case SIGSYS: {
            error_message = make_string("bad system call");
       }; break;
    }

    jmp_buf* env = ____pop_exception_env____();
    error* err_instance = malloc_internal(sizeof(error));
    *err_instance = (error) {.message = error_message, .error_code = sig_code};
    ____add__exception____(err_instance);
    ____handle__exception____();
}

void ____handle__runtime__signals____() {
    signal(SIGSEGV, handle_signal);
    signal(SIGFPE, handle_signal);
    signal(SIGILL, handle_signal);
    signal(SIGBUS, handle_signal);
    signal(SIGABRT, handle_signal);
    signal(SIGTRAP, handle_signal);
    signal(SIGEMT, handle_signal);
    signal(SIGSYS, handle_signal);
}


// __main()__
extern int __main__(int argc, slice argv);
int main(int argc, char** argv) {
int *p = NULL;
    stack_start =(uintptr_t) __builtin_frame_address(0);
    jmp_buf* env;
    int value;
    env = ____push_new_exception_env____();
    value = setjmp(*env);
    ____handle__runtime__signals____();
    if (exception_index != 0) {
      ____handle__exception____();
    }

    init();
    return __main__(argc, __get_argv_slice__(argc, argv));
}