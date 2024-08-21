use crate::{Constant, FunctionDefinition};

pub fn run(
    instructions: &[u8],
    constants: &[Constant],
    variable_names: &[String],
    functions: &[FunctionDefinition],
) {
    let mut variables: Vec<i64> = Vec::new();
    for _ in variable_names {
        // TODO: is there a better way to initialize variables with zeroes?
        variables.push(0);
    }
    run2(
        instructions,
        constants,
        variable_names,
        variables,
        functions,
    );
}

fn run2(
    instructions: &[u8],
    constants: &[Constant],
    variable_names: &[String],
    mut variables: Vec<i64>,
    functions: &[FunctionDefinition],
) -> i64 {
    println!("===== running vm =====");
    println!("instructions ({}): {:?}", instructions.len(), instructions);

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
            OP_JUMP_IF_FALSE => {
                let b1 = instructions[ip] as usize;
                let b2 = instructions[ip + 1] as usize;
                ip += 2;
                let jump_amount = b1 * 256 + b2;
                let value = stack.pop().unwrap();
                println!("jump if false; value: {value}; jump amount: {jump_amount}");
                if value == 0 {
                    ip += jump_amount
                }
            }
            OP_JUMP_FORWARD => {
                let b1 = instructions[ip] as usize;
                let b2 = instructions[ip + 1] as usize;
                ip += 2;
                let jump_amount = b1 * 256 + b2;
                println!("jump forward by {jump_amount}");
                ip += jump_amount
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
                    OP_BINARY_GREATER_THAN => stack.push(if left > right { 1 } else { 0 }),
                    OP_BINARY_LESS_THAN => stack.push(if left < right { 1 } else { 0 }),
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
            OP_READ_VARIABLE => {
                let id = instructions[ip] as usize;
                let value = variables[id];
                println!("read variable {} = {:?}", id, value);
                stack.push(value);
                ip += 1;
            }
            OP_SET_VARIABLE => {
                let id = instructions[ip] as usize;
                let value = stack.pop().unwrap();
                println!("set variable {} = {:?}", id, value);
                variables[id] = value;
                ip += 1;
            }
            OP_PRINTLN => {
                let arg_count = instructions[ip] as usize;
                println!("println with {} arguments", arg_count);
                let mut values: Vec<i64> = Vec::new();
                for _ in 0..arg_count {
                    values.push(stack.pop().unwrap());
                }
                values.reverse(); // TODO: reverse during iteration
                println!("println({:?})", values);
                stack.push(0); // TODO: the println return value is technically Go's "nil" (wich Toi doesn't support)
                ip += 1;
            }
            OP_CALL_FUNCTION => {
                let function_name_id = instructions[ip] as usize;
                ip += 1;
                let constant = &constants[function_name_id];
                let function_name = if let Constant::String(str) = constant {
                    str
                } else {
                    panic!("epxected string");
                };
                let mut ff: Option<&FunctionDefinition> = Option::None;
                for f in functions {
                    if f.name == *function_name {
                        ff = Option::Some(f);
                    }
                }
                let function = ff.unwrap();
                let mut function_variables: Vec<i64> = Vec::new();
                for _ in &function.variables {
                    // TODO: is there a better way to initialize variables with zeroes?
                    function_variables.push(0);
                }

                let mut arguments: Vec<i64> = Vec::new(); // TODO: start with right size
                for _ in 0..function.parameters.len() {
                    arguments.push(stack.pop().unwrap());
                }
                arguments.reverse(); // TODO: reverse while iterating
                for i in 0..arguments.len() {
                    // TODO: fill while iterating above
                    function_variables[i] = arguments[i]
                }

                println!(
                    "calling function {function_name} with arguments: {:?}; variables: {:?}",
                    arguments, function_variables
                );

                let out_var = run2(
                    &function.instructions,
                    constants,
                    &function.variables,
                    function_variables,
                    functions,
                );
                stack.push(out_var);
            }
            _ => panic!("unknown instruction {instruction} at index {}", ip - 1),
        }
    }

    println!("===== end vm ip {ip} =====");

    if !variables.is_empty() {
        return *variables.last().unwrap();
    } else {
        return 0;
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
const OP_READ_VARIABLE: u8 = 8;
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
