Toi is a toy programming language which I'm using to learn about language
design and implementation.


# Disclaimer
The language and its implementation should not be taken seriously. I'm using it
as a playground to test strange and onorthodox ideas and designs. While I strive
for some measure of code quality, the implementation is mostly thrown together
ad-hoc to make quick progress.


# Features and examples
See the [toi/] directory to see test "programs" and their (expected) output.
See the [aoc/] directory for an implementation of Advent of Code 2020.


## Variables, types, and assignments
Toi is a dynamic language. It supports integers, strings, arrays, and maps. It
only has global variables. Variables can be re-assigned to any new value of any
type, but types are strict (so you cannot add string `"3"` and integer `5` to
get the number `8` - nor the string `"35"` - for example).

```
i = 15
i = i + 27
println(i) // prints 42
i = "Hello"
println(i, "World") // prints "Hello, World"
```


## Statements
Each line is a statement terminated by a newline. Statements can be either
assignment (which is not an expression), a `while` loop, a loop exit, an `if`
statement, or an expression.

```
i = 15 // Assignment is a statement
println(i) // println(i) is an expression, written as a statement
i + 27 // This is an expression without side-effects, so not very useful
```

To continue a statement or expression on the next line, you can comment out the
newline using `//` (note that nothing but the newline can follow the `//`, or it
will not be commented out).

```
i = 1 + //
  3 * //
  5
// The above sets i to 16
```


## If statement and logical operators
If statements don't have parentheses, but the curly braces are mandatory. To
branch to statements when the if condition doesn't match, the `otherwise`
keyword is used.
All the usual logical operators are supported, like `<`, `>`, `>=`, `<=`, `==`,
and `<>` for "not equals".
Logical expressions can be composed by using `and` for logical AND and `or` for
logical OR.

Booleans do not exist (yet). The number zero (`0`) is false, all other numbers
are true (and other types cannot act as boolean).

```
if i == 42 { // equivalent to: "if i" because 42 is not false (0)
    println("i is 42")
} otherwise {
    println("is not 42")
}

if i < 42 and greeting == "world" or j <> 13 {
    // code
}
```


## Loops
Toi currently only supports while and for loops.

While loops run when their expression evaluates to true (not zero (`0`)) and
stops running when the expression evaluates to false (zero (`0`)).

For loops iterate over each element of an array or map.

A loop can be exited prematurely by using `exit loop`. You can commence to the
next iteration using `next iteration`.

```
i = 0
while 1 {
    i = i + 1
    if i == 30 {
        next iteration // skips printing 30
    }
    println("i is:", i)
    if i == 42 {
        exit loop
    }
}
```

```
array = array(42, 1337, 5521)

for value = [array]index {
    println("index", index, "value", value)
}
```


## Arrays and maps
Toi supports arrays and maps as container types. They are created using the
`array()` and `map()` built-in functions respectively. A wide range of built-in
functions exist to deal with them. Array and map access can be written using
square brackets, e.g. to get the 3rd element of an array: `[array]4`.

Arrays use integer indices, and maps use string keys.

```
items = map()
i = 0
while i < len(lines) {
    line = [lines]i // accesses element i from the lines array
    i = i + 1
    keyAndValue = split(line, "|") // splits a string into substrings separated by |
    key = [keyAndValue]0
    [items]key = [keyAndValue]1 // assigns to the items map using the key in the key variable
}

values = array()
i = 0
keys = keys(items) // the built-in function keys() returns the keys of the map
while i < len(keys) {
    key = [keys]i
    i = i + 1
    push(values, [items]key) // pushes a new value into the array
}

[values](len(values)-1) = "last item" // re-assigns the last item of the array
pop(values) // removes the last value from the array
```

Array and map can be created with values:
```
values = array(1, 1, 2, 3, 5, 8, 13, 21)
// Alternating keys and values:
items = map("a", 1, "b", 2, "c", 3)
```


## Strings
Toi has UTF-8 strings. Toi has no characters (yet?). A string literal is written
as any text in double quotes (`"`). Several built-in utility functions are
provided to work with strings.

```
s = "31 String Literal"
words = split(s, " ")
i = int([words]0) + 11 // int() converts a string to an int
s = string(i) _ " " _ [words]1 _ " " _ [words]2 // string() converts an int into a string
// _ concatenates strings
```

Double quotes inside strings can be escaped using `${"}` inside the string
literal, like so:

```
println("Toi is ${"}stable${"} and looks ${"}nice${"}")
```

## Functions
A simple function that does not take any arguments can be written like this:

```
printHelloWorld|| {
    println("Hello, world!")
}

printHelloWorld()
```

Parameter names are written between the opening `|` and the closing `|`:

```
printGreeting|greeting what| {
    println(greeting _ ", " _ what _ "!")
}

printGreeting("Hello", "world")
```

To get a value out of a function, an out-variable is used, which is written
after the closing `|`:

```
getGreeting|greeting what| fullGreeting {
    fullGreeting = greeting _ ", " _ what _ "!"
}

println(getGreeting("Hello", "world"))
```

A function can be exited early by using `exit function`:

```
printNumbers|maximum| {
    i = 0
    while 1 {
        println(i)
        if i == maximum {
            exit function
        }
        i = i + 1
    }
}
```


## Other built-in functions
`inputLines()` returns the standard input as lines
`chars(s)` returns an array with the characters in a string (each element is a string of length 1)
`isSet(map, key)` returns 1 if the key is set in the map, and 0 if it's not
`unset(map, key)` removes the key from the map
`sort(array)` sorts an array by lexicographically order; custom types are sorted by the order of their fields


## Custom types
Custom types can be defined to group data:

```
Item{id data}

item = Item(42, "Hello, world")
println(item.id) // Prints: 42
item.data = "Hello, custom type"
println(item) // Prints: Item{id=42,data=Hello, custom type}
```


# Implementation
* `tokenizer.go` lexes to tokens
* `parser.go` parses into an AST
* `interpreter.go` interprets directly from the AST
* `compiler.go` compiles the AST into a custom bytecode
* `vm.go` interprets the bytecode output by the compiler

When running a script, Toi runs it both using the tree interpreter and the VM
interpreter and tests that the outputs are the same (potential side effects that
do not print to standard output are not validated).
