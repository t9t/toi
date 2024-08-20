use crate::Constant;

pub fn run(instructions: &[u8], constants: &[Constant]) {
    println!("===== running vm =====");
    println!("instructions: {:?}", instructions);

    let mut stack: Vec<i64> = Vec::new();

    let mut ip = 0;
    while ip < instructions.len() {
        let instruction = instructions[ip];
        println!("instruction {ip}: {}; stack: {:?}", instruction, stack);
        ip += 1;
        match instruction {
            OP_POP => {
                println!("pop: {:?}", stack.pop().unwrap());
            }
            OP_BINARY => {
                let binop = instructions[ip];
                println!("binary of {}", binop);
                ip += 1;
                let right = stack.pop().unwrap();
                let left = stack.pop().unwrap();

                match binop {
                    OP_BINARY_PLUS => stack.push(left + right),
                    OP_BINARY_SUBTRACT => stack.push(left - right),
                    OP_BINARY_MULTIPLY => stack.push(left * right),
                    OP_BINARY_DIVIDE => stack.push(left / right),
                    _ => panic!("unknown binary operation {binop} at index {}", ip - 1),
                }
            }
            OP_INLINE_NUMBER => {
                let value = instructions[ip] as i64;
                println!("inline number {}", value);
                stack.push(value);
                ip += 1;
            }
            OP_LOAD_CONSTANT => {
                let id = instructions[ip] as usize;
                let constant = &constants[id];
                println!("load constant {} = {:?}", id, constant);
                match constant {
                    Constant::Number(number) => {
                        println!("  number: {}", number);
                        stack.push(*number);
                    }
                    Constant::String(str) => {
                        println!("  string: {}", str);
                        todo!("cannot handle strings yet");
                    }
                }
                ip += 1;
            }
            OP_PRINTLN => {
                let arg_count = instructions[ip] as usize;
                println!("println with {} arguments", arg_count);
                let mut values: Vec<i64> = Vec::new();
                for _ in 0..arg_count {
                    values.push(stack.pop().unwrap());
                }
                println!("println({:?})", values);
                stack.push(0); // TODO: the println return value is technically Go's "nil" (wich Toi doesn't support)
                ip += 1;
            }
            _ => panic!("unknown instruction {instruction} at index {}", ip - 1),
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
