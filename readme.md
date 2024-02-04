## sea - a toy programming language frontend

Sea (pronounced as "C" :)) is a programming language that compiles to LLVM code. As a hobby project, it has served as a valuable teacher, providing significant insights into various aspects of computer science. It is currently ongoing process, so it has a lot of bugs and it is just hobby project.


### Example
```sea
    package main

    # use "another_package" is not implemented yet
  
    struct Type {
        i32 field_1 # 32-bit integer
        f64 field_2 # 64-bit flot
    }
    
    # define some methods to `Type*` type. These methods defined in  
    impl Type {
    
        fun i32 getField1() {
            return this.field_1
        }
        
        fun void setField1(i32 field_1) {
            this.field_1 = field_1
        }
        
    }
    
    # this is a single line comment
    fun i32 main() {
        /*
               This is multi-line comment
               You can write anything you want
        */
        
        var i32 f1 = 1
        var f64 f2 = 1.0
        
        var Type t = Type{
            field_1: f1,
            field_2: f2
        }
        
        # this is pointer type
        var Type* t2 = &t
        
        var Type* t3 = &Type{
            # casting operations is like function call with "cast_" prefix and primitive type postfix
            field_1: cast_i32(f2),
            field_2: cast_f64(f1)
        }
        
        return 0
    }

```


### Known Bugs

1. [x] Couldn't count :/