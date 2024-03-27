## 🎉 sea - programming language 🚀

Sea, pronounced as `C` is a programming language that compiles to LLVM Backend. As a hobby project, it has served as a valuable teacher, providing significant insights into various aspects of computer science. It is early and unstable version and currently ongoing process.

Note that: Only tested on arm64_macos. In other architectures or OS may happens unexpected behaviours.

Here is example codes written in sea, until I write documentation the code will help, I hope.

### Hello world Application
```kotlin
   package main
   
   fun i32 main() {
        println("Hello, world!")
        return 0
   } 
```

### Simple Json Parser as an example
```kotlin
    package main
    
    extern fun string open_file_read(string path)
    extern fun i32 str_len(string str)
    extern fun void exit(i32 code)
    extern fun char* to_char_pointer(string s)
    
    const string ARRAY_KIND = "array"
    const string OBJECT_KIND = "object"
    const string KEY_VALUE_KIND = "key_value"
    const string STRING_KIND = "string"
    const string BOOL_KIND = "bool"
    const string NUMBER_KIND = "number"
    
    struct JsonNode {
        # node kind
        string kind
    
        # for number literals
        string numberValue
        # for boolean literals
        string boolValue
    
        # for string values
        string stringValue
    
        # for object key value pointer
        JsonNode*[] key_value_nodes
    
        # for object key
        string key
    
        # for object value
        JsonNode* value
    
        # for array elems
        JsonNode*[] elems
    
    }
    
    impl JsonNode {
        fun void print_node(i32 indent) {
             if this.kind == ARRAY_KIND {
               print_indent(indent)
               println("[")
               for var i64 i = 0; i < len(this.elems); ++i; {
                    this.elems[i].print_node(indent + 1)
               }
               print_indent(indent)
               println("]")
             } else if this.kind == OBJECT_KIND {
                print_indent(indent)
                println("{")
                for var i64 i = 0; i < len(this.key_value_nodes); i = i++ {
                    this.key_value_nodes[i].print_node(indent + 1)
                    if i != len(this.key_value_nodes) - 1 {
                        print(",")
                    }
                    println("")
                }
                print_indent(indent)
                println("}")
             } else if this.kind == KEY_VALUE_KIND {
               print_indent(indent)
               print("\"" + this.key + "\": ")
               this.value.print_node(indent + 1)
             } else if this.kind == STRING_KIND {
                print("\"" + this.stringValue + "\"")
             } else if this.kind == BOOL_KIND {
                print(this.boolValue)
             } else if this.kind == NUMBER_KIND {
                print(this.numberValue)
             }
        }
    }
    
    struct JsonParser {
        string file
        i64 current_pos
        i64 file_len
        string current_token
        bool is_initialized
    }
    
    
    impl JsonParser {
    
        fun void eat_whitespaces() {
            for this.file_len > this.current_pos {
                var bool is_whitespace = this.is_whitespace()
                if !is_whitespace {
                    break
                }
                this.advance_pos()
            }
            return
        }
    
        fun string parse_quote_literal() {
            var string value = ""
    
            for this.current_char() != '"' {
                value = value + this.current_char()
                this.advance_pos()
            }
    
            this.advance_pos()
            return value
        }
    
        fun string parse_num_literal() {
            var string value = ""
            for this.is_numeric() {
                value = value + this.current_char()
                this.advance_pos()
            }
    
            return value
        }
    
        fun string next_token() {
            this.is_initialized = true
            var string value = ""
            for this.file_len > this.current_pos {
                if this.is_whitespace() {
                    this.eat_whitespaces()
                    continue
                }
    
                if this.is_numeric() {
                    value = this.parse_num_literal()
                    this.current_token = value
                    return "<number>"
                }
    
    
                value = value + this.current_char()
                this.advance_pos()
    
    
                if value == "," {
                    this.current_token = value
                    return "<comma>"
                }
    
                if value == "[" {
                    this.current_token = value
                    return "<array_start>"
                }
    
                if value == "{" {
                    this.current_token = value
                    return "<object_start>"
                }
    
                if value == "\"" {
                    value = this.parse_quote_literal()
                    this.current_token = value
                    return "<double_quote>"
                }
    
                if value == "null" {
                    this.current_token = value
                    return "<null>"
                }
    
                if value == "true" {
                    this.current_token = value
                    return "<true>"
                }
    
                if value == "false" {
                    this.current_token = value
                    return "<false>"
                }
    
                if value == ":" {
                    this.current_token = value
                    return "<colon>"
                }
    
                if value == "}" {
                    this.current_token = value
                    return "<object_end>"
                }
    
                if value == "]" {
                    this.current_token = value
                    return "<array_end>"
                }
    
            }
    
            this.current_token = "<eof>"
            return this.current_token
        }
    
        fun string get_current_token() {
            if !this.is_initialized {
                this.next_token()
                this.is_initialized = true
            }
    
            return this.current_token
        }
    
        fun JsonNode* parse() {
           var string kind = this.next_token()
           if kind == "<array_start>" {
               return this.parse_array()
           } else if kind == "<object_start>" {
                return this.parse_object()
           } else if kind == "<number>" {
                var string numberValue = this.get_current_token()
                var JsonNode* node = &JsonNode{kind: NUMBER_KIND, numberValue: numberValue}
                this.next_token()
                return node
           } else if kind == "<true>" || kind == "<false>" {
                var string value = this.get_current_token()
                var JsonNode* node = &JsonNode{kind: BOOL_KIND, boolValue: value}
                this.next_token()
                return node
           } else if kind == "<double_quote>" {
                var string value = this.get_current_token()
                var JsonNode* node = &JsonNode{kind: STRING_KIND, stringValue: value}
                this.next_token()
                return node
           } else {
               printf_internal("invalid token: %s", kind)
               exit(1)
           }
    
           print("in here nil")
           return nil
        }
    
        fun JsonNode* parse_array() {
            var JsonNode* jsonNode = &JsonNode{kind: ARRAY_KIND, elems: []}
            for this.get_current_token() != "<array_end>" && this.get_current_token() != "<eof>"  {
                var JsonNode* child = this.parse()
                append(jsonNode.elems, child)
                this.next_token()
            }
            return jsonNode
        }
    
        fun JsonNode* parse_object() {
            var JsonNode* jsonNode = &JsonNode{kind: OBJECT_KIND, key_value_nodes: []}
            for true {
                var JsonNode* keyValueNode = &JsonNode{kind: KEY_VALUE_KIND}
                var string key = this.next_token()
                keyValueNode.key = this.get_current_token()
                this.next_token() # ignore that; colon
                var JsonNode* val = this.parse()
                keyValueNode.value = val
                append(jsonNode.key_value_nodes, keyValueNode)
    
                if this.get_current_token() == "}" {
                    this.next_token()
                    break
                }
            }
    
            return jsonNode
        }
    
        fun bool is_whitespace() {
            return this.current_char() == '\t' || this.current_char() == '\n' || this.current_char() == ' ' || this.current_char() == '\r'
        }
    
        fun bool is_numeric() {
            return this.current_char() >= '0' && this.current_char() <= '9'
        }
    
        fun char current_char() {
            return this.file[this.current_pos]
        }
    
        fun void advance_pos() {
            if to_i64(str_len(this.file)) <= this.current_pos {
                return
            }
            this.current_pos = this.current_pos + 1
            return
        }
    
    }
    
    fun string value_node_str(JsonNode* jsonNode) {
        if jsonNode == nil {
            return ""
        }
    
        if jsonNode.kind == STRING_KIND {
            return jsonNode.stringValue
        } else if jsonNode.kind == BOOL_KIND {
            return jsonNode.boolValue
        } else if jsonNode.kind == NUMBER_KIND {
            return jsonNode.numberValue
        }
    
        return "<T_ERR | unknown value kind>"
    }
    
    
    fun void print_indent(i32 indent) {
        for var i32 i = 0; i < indent; i++ {
            print(" ")
        }
        return
    }
    
    
    # main entry function
    fun i32 main(i32 argc, string[] args) {
          var string file = open_file_read("./example.json")
          var JsonParser* jsonParser = &JsonParser{
                file: file,
                current_pos: 0,
                is_initialized: false,
                file_len: to_i64(str_len(file)),
          }
    
       var JsonNode* json_node = jsonParser.parse()
       json_node.print_node(0)
       return 0
    }

```

#### Commands:

`sea
    sealang run <file_name>`

`sealang build <file_name> -o <output_path>`


### Known Bugs

1. [ ] Probably too many