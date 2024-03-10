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

void compare_string(string a, string b) {
    puts_str(a);
    puts_str(b);
}


void handle_signal() {
    signal(SIGSEGV, h_s);
}
