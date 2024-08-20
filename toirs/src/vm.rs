pub fn run(instructions: &[u8]) {
    println!("instructions: {:?}", instructions);

    let mut stack: Vec<i64> = Vec::new();

    let mut i = 0;
    while i < instructions.len() {
        let instruction = instructions[i];
        println!("instruction {i}: {}", instruction);
        i += 1;
        match instruction {
            OP_POP => println!("pop"),
            OP_BINARY => {
                println!("binary of {}", instructions[i]);
                i += 1;
            }
            OP_INLINE_NUMBER => {
                println!("inline number {}", instructions[i]);
                i += 1;
            }
            OP_LOAD_CONSTANT => {
                println!("load constant {}", instructions[i]);
                i += 1;
            }
            OP_PRINTLN => {
                let arg_count = instructions[i] as usize;
                println!("println with {} arguments", arg_count);
                i += 1;
            }
            _ => panic!("unknown instruction {instruction} at index {i}"),
        }
    }
}

const OP_POP: u8 = 0;
const OP_BINARY: u8 = 1;
const OP_NOT: u8 = 2;
const OP_JUMP_IF_FALSE: u8 = 3;
const OP_JUMP_FORWARD: u8 = 4;
const OP_JUMP_BACK: u8 = 5;
const OP_INLINE_NUMBER: u8 = 6;
const OP_LOAD_CONSTANT: u8 = 7;
const OP_READ_VARIABLE: u8 = 9;
const OP_SET_VARIABLE: u8 = 9;
const OP_CALL_BUILTIN: u8 = 10;
const OP_CALL_FUNCTION: u8 = 11;
const OP_PRINTLN: u8 = 12;
const OP_DUPLICATE: u8 = 13;
const OP_INVALID: u8 = 14;

// OP_BINARY arguments
const OP_BINARY_PLUS: u8 = 0;
const OP_BINARY_SUBTRACT: u8 = 1;
const OP_BINARY_MULTIPLY: u8 = 2;
const OP_BINARY_DIVIDE: u8 = 3;
const OP_BINARY_REMAINDER: u8 = 4;
const OP_BINARY_EQUAL: u8 = 5;
const OP_BINARY_GREATER_THAN: u8 = 6;
const OP_BINARY_LESS_THAN: u8 = 7;
const OP_BINARY_CONCAT: u8 = 8;
