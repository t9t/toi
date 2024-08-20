use std::env;
use std::fs::File;
use std::io::{self, BufRead};

mod vm;

fn main() {
    let file_path = env::args().nth(1).expect("No file path provided");

    let lines = read_lines_from_file(&file_path);
    // Print out the lines.

    println!("lines: {}", lines.len());

    let rest: &[String] = &lines;

    let (constants, rest) = parse_constants(&rest);
    println!("Constants: {:?}", constants);
    println!("Rest: {:?}", rest);

    let (functions, rest) = parse_functions(&rest);
    println!("Functions: {:?}", functions);
    println!("Rest: {:?}", rest);

    let (variables, rest) = parse_variables(&rest);
    println!("Variables: {:?}", variables);
    println!("Rest: {:?}", rest);

    let (instructions, rest) = parse_instructions(&rest);
    println!("Instructions: {:?}", instructions);
    println!("Rest: {:?}", rest);

    if rest.len() != 0 {
        panic!("expected no more data, but got: {:?}", rest)
    }

    println!("done reading");
    vm::run(&instructions);
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
        println!("Reading line {line}");
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

fn parse_functions(lines: &[String]) -> (Vec<Constant>, &[String]) {
    assert("functions", &lines[0]);
    assert("0", &lines[1]);

    // TODO: implement
    return (vec![], &lines[2..]);
}

fn parse_variables(lines: &[String]) -> (Vec<Constant>, &[String]) {
    assert("variables", &lines[0]);
    assert("0", &lines[1]);

    // TODO: implement
    return (vec![], &lines[2..]);
}

fn parse_instructions(lines: &[String]) -> (Vec<u8>, &[String]) {
    assert("instructions", &lines[0]);

    let count: usize = lines[1].parse().unwrap();
    let mut instructions: Vec<u8> = Vec::new();
    for i in 2..=count + 1 {
        let line = &lines[i];
        println!("Reading line {line}");
        instructions.push(line.parse().unwrap());
    }
    return (instructions, &lines[2 + count..]);
}

#[derive(Debug)]
enum Constant {
    Number(i64),
    String(String),
}
