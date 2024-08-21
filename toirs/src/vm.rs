use crate::{Constant, FunctionDefinition};

pub fn run(
    instructions: &[u8],
    constants: &[Constant],
    variable_names: &[String],
    functions: &[FunctionDefinition],
) {
    let variables: Vec<i64> = vec![0; variable_names.len()];
    let mut stack = Stack::new();
    run2(instructions, constants, variables, functions, &mut stack);
    println!("stack max: {}", stack.max);
}

struct Stack {
    stack: Vec<i64>,
    max: usize,
}

impl Stack {
    fn new() -> Stack {
        return Stack {
            stack: Vec::with_capacity(20),
            max: 0,
        };
    }

    fn push(&mut self, value: i64) {
        self.stack.push(value);
        self.max = std::cmp::max(self.max, self.stack.len());
    }

    fn pop(&mut self) -> i64 {
        let value = self.stack.pop().unwrap();
        return value;
    }
}

fn run2(
    instructions: &[u8],
    constants: &[Constant],
    mut variables: Vec<i64>,
    functions: &[FunctionDefinition],
    mut stack: &mut Stack,
) {
    let mut ip = 0;
    while ip < instructions.len() {
        let instruction = instructions[ip];
        ip += 1;
        match instruction {
            OP_POP => {
                stack.pop();
            }
            OP_JUMP_IF_FALSE => {
                let b1 = instructions[ip] as usize;
                let b2 = instructions[ip + 1] as usize;
                ip += 2;
                let jump_amount = b1 * 256 + b2;
                let value = stack.pop();
                if value == 0 {
                    ip += jump_amount
                }
            }
            OP_JUMP_FORWARD => {
                let b1 = instructions[ip] as usize;
                let b2 = instructions[ip + 1] as usize;
                ip += 2;
                let jump_amount = b1 * 256 + b2;
                ip += jump_amount
            }
            OP_BINARY => {
                let binop = instructions[ip];
                ip += 1;
                let right = stack.pop();
                let left = stack.pop();

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
                stack.push(value);
                ip += 1;
            }
            OP_LOAD_CONSTANT => {
                let id = instructions[ip] as usize;
                let constant = &constants[id];
                match constant {
                    Constant::Number(number) => {
                        stack.push(*number);
                    }
                    Constant::String(_) => {
                        todo!("cannot handle strings yet");
                    }
                }
                ip += 1;
            }
            OP_READ_VARIABLE => {
                let id = instructions[ip] as usize;
                let value = variables[id];
                stack.push(value);
                ip += 1;
            }
            OP_SET_VARIABLE => {
                let id = instructions[ip] as usize;
                let value = stack.pop();
                variables[id] = value;
                ip += 1;
            }
            OP_PRINTLN => {
                let arg_count = instructions[ip] as usize;
                let mut values: Vec<i64> = vec![0; arg_count];
                for i in 0..arg_count {
                    values[i] = stack.pop();
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
                let function = functions.iter().find(|f| f.name == *function_name).unwrap();
                let mut function_variables: Vec<i64> = vec![0; function.variables.len()];
                for i in (0..function.parameters.len()).rev() {
                    function_variables[i as usize] = stack.pop();
                }

                run2(
                    &function.instructions,
                    constants,
                    function_variables,
                    functions,
                    &mut stack,
                );
            }
            _ => panic!("unknown instruction {instruction} at index {}", ip - 1),
        }
    }

    // TODO: should check has_out_var
    if !variables.is_empty() {
        let out_value = *variables.last().unwrap();
        stack.push(out_value);
    } else {
        stack.push(0);
    }
}

const OP_POP: u8 = 0;
const OP_BINARY: u8 = 1;
// TODO const OP_NOT: u8 = 2;
const OP_JUMP_IF_FALSE: u8 = 3;
const OP_JUMP_FORWARD: u8 = 4;
// TODO const OP_JUMP_BACK: u8 = 5;
const OP_INLINE_NUMBER: u8 = 6;
const OP_LOAD_CONSTANT: u8 = 7;
const OP_READ_VARIABLE: u8 = 8;
const OP_SET_VARIABLE: u8 = 9;
// TODO const OP_CALL_BUILTIN: u8 = 10;
const OP_CALL_FUNCTION: u8 = 11;
const OP_PRINTLN: u8 = 12;
// TODO const OP_DUPLICATE: u8 = 13;
// TODO const OP_INVALID: u8 = 14;

// OP_BINARY arguments
const OP_BINARY_PLUS: u8 = 0;
const OP_BINARY_SUBTRACT: u8 = 1;
const OP_BINARY_MULTIPLY: u8 = 2;
const OP_BINARY_DIVIDE: u8 = 3;
// TODO const OP_BINARY_REMAINDER: u8 = 4;
// TODO const OP_BINARY_EQUAL: u8 = 5;
const OP_BINARY_GREATER_THAN: u8 = 6;
const OP_BINARY_LESS_THAN: u8 = 7;
// TODO const OP_BINARY_CONCAT: u8 = 8;
