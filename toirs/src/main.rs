use std::env;
use std::fs::File;
use std::io::{self, BufRead};

mod vm;

fn main() {
    let file_path = env::args().nth(1).expect("No file path provided");
    let lines = read_lines_from_file(&file_path);

    let rest: &[String] = &lines;

    let (constants, rest) = parse_constants(&rest);
    let (functions, rest) = parse_functions(&rest);
    let (variables, rest) = parse_variables(&rest);
    let (instructions, rest) = parse_instructions(&rest);

    if rest.len() != 0 {
        panic!("expected no more data, but got: {:?}", rest)
    }

    let start = std::time::Instant::now();
    vm::run(&instructions, &constants, &variables, &functions);
    println!("Run time: {:?}", start.elapsed());
}

fn assert(expected: &str, actual: &str) {
    if actual != expected {
        panic!("expected {} but got {}", expected, actual)
    }
}

fn read_lines_from_file(file_path: &str) -> Vec<String> {
    let file = File::open(file_path).unwrap();
    let reader = io::BufReader::new(file);

    return reader.lines().filter_map(Result::ok).collect();
}

fn parse_constants(lines: &[String]) -> (Vec<Constant>, &[String]) {
    assert("constants", &lines[0]);

    let count: usize = lines[1].parse().unwrap();
    let mut constants: Vec<Constant> = Vec::new();
    for i in 2..=count + 1 {
        let line = &lines[i];
        let parts: Vec<&str> = line.split(":").to_owned().collect();
        let constant = if parts[0] == "int" {
            Constant::Number(parts[1].parse().unwrap())
        } else if parts[0] == "string" {
            Constant::String(parts[1].into())
        } else {
            panic!("unsupported type {}", parts[0])
        };
        constants.push(constant);
    }
    return (constants, &lines[2 + count..]);
}

fn parse_functions(lines: &[String]) -> (Vec<FunctionDefinition>, &[String]) {
    assert("functions", &lines[0]);

    let count: usize = lines[1].parse().unwrap();
    let mut functions: Vec<FunctionDefinition> = Vec::new();
    let mut rest = &lines[2..];
    for _ in 2..=count + 1 {
        let (function, more) = parse_function(&rest);
        rest = more;
        functions.push(function);
    }

    return (functions, &rest);
}

fn parse_function(lines: &[String]) -> (FunctionDefinition, &[String]) {
    let name = &lines[0];
    let has_out_var: bool = lines[1].parse().unwrap();

    let rest: &[String] = &lines[2..];
    let (parameters, rest) = parse_strings("parameters", &rest);
    let (variables, rest) = parse_variables(&rest);
    let (instructions, rest) = parse_instructions(&rest);

    return (
        FunctionDefinition {
            name: name.to_owned(),
            has_out_var,
            parameters,
            variables,
            instructions,
        },
        &rest,
    );
}

fn parse_strings<'a>(header: &str, lines: &'a [String]) -> (Vec<String>, &'a [String]) {
    assert(header, &lines[0]);

    let count: usize = lines[1].parse().unwrap();
    let mut strings: Vec<String> = Vec::new();
    for i in 2..=count + 1 {
        let line = &lines[i];
        strings.push(line.to_owned());
    }
    return (strings, &lines[2 + count..]);
}

fn parse_variables(lines: &[String]) -> (Vec<String>, &[String]) {
    return parse_strings("variables", &lines);
}

fn parse_instructions(lines: &[String]) -> (Vec<u8>, &[String]) {
    assert("instructions", &lines[0]);

    let count: usize = lines[1].parse().unwrap();
    let mut instructions: Vec<u8> = Vec::new();
    for i in 2..=count + 1 {
        let line = &lines[i];
        instructions.push(line.parse().unwrap());
    }
    return (instructions, &lines[2 + count..]);
}

#[derive(Debug)]
enum Constant {
    Number(i64),
    String(String),
}

#[derive(Debug)]
struct FunctionDefinition {
    name: String,
    has_out_var: bool,
    parameters: Vec<String>,
    variables: Vec<String>,
    instructions: Vec<u8>,
}
