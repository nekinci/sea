#include<stdio.h>
#include<stdlib.h>
#include <stdarg.h>
#include<stdlib.h>
#include <string.h>
#include <fcntl.h>

#pragma clang diagnostic ignored "-Wvarargs"


typedef struct {
    char* buffer;
    size_t size;
    size_t cap;
} string;


void* memcpy_internal(void* dest, const void* src) {
    size_t s = sizeof(dest) * 2; // fixme
    return memcpy(dest, src, s);
}

string make_string(char* buffer) {
    size_t size = strlen(buffer);
    string result;
    result.buffer = buffer;
    result.size = size;
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

void pass_string(string s) {
    printf("zu: %l", s.size);
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


int main2() {
    return 0;
}
