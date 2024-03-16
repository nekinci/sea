#include<stdio.h>
#include<stdlib.h>
#include <stdarg.h>
#include<stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <signal.h>
#include <stdbool.h>
#include <errno.h>

#pragma clang diagnostic ignored "-Wvarargs"


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

slice make_slice() {
    slice s;
    s.cap = 2;
    s.size = 0;
    s.data_list = malloc(sizeof(char*) * s.cap);
    return s;
}

void append_slice_data(slice* s, char* data) {
    if (s -> size >= s -> cap) {
        s -> cap *= 2;
        s -> data_list = realloc(s -> data_list, sizeof(char*) * s -> cap);
    }

    *(s -> data_list + s -> size) = data;
    s -> size++;
}

char* access_slice_data(slice s, int index) {
    if (index >= s.size) {
        printf("Index out of bound error occurred: %d, slice size is: %zu", index, s.size);
        exit(255);
    }

    char** data = (char**)(s.data_list + index);
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


int r_runtime_scanf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vscanf(fmt, args);
    va_end(args);
    return r;
}


int r_runtime_printf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vprintf(fmt, args);
    va_end(args);
    return r;
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

int another_scanf(string* s) {
    s -> size = 5;
    return 0;
}

int yet_another_scanf(string s) {
s.size = 4;
    return 0;
}

string* abc(int x) {
    return NULL;
}

int cba(int* x) {
    return 0;
}


void* malloc_internal(size_t s) {
    return malloc(s);
}

size_t strlen_internal(string str) {
    return str.size;
}

int sum(int a, int b) {
    return a +b;
}

void r_runtime_exit(int status) {
    exit(status);
}

int open_internal(const char* path, int oflag) {
// TODO implement that function to handle file operations
    exit(38);
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
    return puts(s.buffer);
}


void doSomething(int* ptr) {
    ptr[1] = 11;
    printf("%d\n", ptr[0]);
}

int main2() {
    return 0;
}


void h_s(int signo) {
    printf("caught signal %d\n", signo);
    exit(255);
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

void handle_signal() {
    signal(SIGSEGV, h_s);
}
