use std::env;
use std::fs;
use std::io::{self, Write};
use std::path::Path;
use std::process::{self, Command};

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        print_usage();
        process::exit(3);
    }

    let subcommand = &args[1];
    match subcommand.as_str() {
        "build" | "run" => {
            if args.len() < 3 {
                print_usage();
                process::exit(3);
            }
            run_compile(&args);
        }
        "repl" => run_repl(),
        _ => {
            eprintln!("Unknown command: {}", subcommand);
            print_usage();
            process::exit(3);
        }
    }
}

fn print_usage() {
    eprintln!("Usage:");
    eprintln!("  vex build <file.vx> [-o <output>] [--emit-go <dir>]");
    eprintln!("  vex run <file.vx>");
    eprintln!("  vex repl");
}

fn run_compile(args: &[String]) {
    let subcommand = &args[1];
    let source_path = Path::new(&args[2]);
    if !source_path.exists() {
        eprintln!("File not found: {}", source_path.display());
        process::exit(3);
    }

    let mut output_name: Option<String> = None;
    let mut emit_go_dir: Option<String> = None;

    let mut i = 3;
    while i < args.len() {
        match args[i].as_str() {
            "-o" => {
                i += 1;
                if i >= args.len() {
                    eprintln!("-o requires an argument");
                    process::exit(3);
                }
                output_name = Some(args[i].clone());
            }
            "--emit-go" => {
                i += 1;
                if i >= args.len() {
                    eprintln!("--emit-go requires an argument");
                    process::exit(3);
                }
                emit_go_dir = Some(args[i].clone());
            }
            other => {
                eprintln!("Unknown option: {}", other);
                process::exit(3);
            }
        }
        i += 1;
    }

    let source = match fs::read_to_string(source_path) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Could not read {}: {}", source_path.display(), e);
            process::exit(3);
        }
    };

    let result = vex::compile(&source, &source_path.display().to_string());

    if !result.diagnostics.is_empty() {
        for diag in &result.diagnostics {
            eprintln!("{}", diag.render(&result.source_map));
        }
        process::exit(1);
    }

    let binary_name = output_name.unwrap_or_else(|| {
        source_path
            .file_stem()
            .unwrap()
            .to_string_lossy()
            .to_string()
    });

    let tmp_dir = tempdir();

    fs::write(tmp_dir.join("go.mod"), &result.go_mod).expect("failed to write go.mod");
    fs::write(tmp_dir.join("main.go"), &result.go_source).expect("failed to write main.go");

    if let Some(ref vexrt) = result.vexrt {
        let vexrt_dir = tmp_dir.join("vexrt");
        fs::create_dir_all(&vexrt_dir).expect("failed to create vexrt directory");
        fs::write(vexrt_dir.join("option.go"), &vexrt.option_go)
            .expect("failed to write option.go");
        fs::write(vexrt_dir.join("result.go"), &vexrt.result_go)
            .expect("failed to write result.go");
        fs::write(vexrt_dir.join("collections.go"), &vexrt.collections_go)
            .expect("failed to write collections.go");
    }

    for pkg in &result.extra_packages {
        let pkg_dir = tmp_dir.join(&pkg.name);
        fs::create_dir_all(&pkg_dir).expect("failed to create package directory");
        let file_name = pkg.name.rsplit('/').next().unwrap_or(&pkg.name);
        fs::write(pkg_dir.join(format!("{}.go", file_name)), &pkg.source)
            .expect("failed to write package file");
    }

    if let Some(ref dir) = emit_go_dir {
        let dest = Path::new(dir);
        fs::create_dir_all(dest).expect("failed to create emit-go directory");
        fs::write(dest.join("go.mod"), &result.go_mod).expect("failed to write go.mod");
        fs::write(dest.join("main.go"), &result.go_source).expect("failed to write main.go");
        if let Some(ref vexrt) = result.vexrt {
            let vexrt_dest = dest.join("vexrt");
            fs::create_dir_all(&vexrt_dest).expect("failed to create vexrt directory");
            fs::write(vexrt_dest.join("option.go"), &vexrt.option_go)
                .expect("failed to write option.go");
            fs::write(vexrt_dest.join("result.go"), &vexrt.result_go)
                .expect("failed to write result.go");
            fs::write(vexrt_dest.join("collections.go"), &vexrt.collections_go)
                .expect("failed to write collections.go");
        }
        for pkg in &result.extra_packages {
            let pkg_dest = dest.join(&pkg.name);
            fs::create_dir_all(&pkg_dest).expect("failed to create package directory");
            let file_name = pkg.name.rsplit('/').next().unwrap_or(&pkg.name);
            fs::write(pkg_dest.join(format!("{}.go", file_name)), &pkg.source)
                .expect("failed to write package file");
        }
    }

    let output_path = env::current_dir().unwrap().join(&binary_name);

    let go_build = Command::new("go")
        .arg("build")
        .arg("-o")
        .arg(&output_path)
        .current_dir(&tmp_dir)
        .output();

    let _ = fs::remove_dir_all(&tmp_dir);

    match go_build {
        Ok(output) => {
            if !output.status.success() {
                eprintln!(
                    "go build failed:\n{}",
                    String::from_utf8_lossy(&output.stderr)
                );
                process::exit(2);
            }
        }
        Err(e) => {
            eprintln!("Failed to run go build: {}", e);
            process::exit(2);
        }
    }

    if subcommand == "run" {
        let status = Command::new(&output_path).status();
        let _ = fs::remove_file(&output_path);
        match status {
            Ok(s) => process::exit(s.code().unwrap_or(1)),
            Err(e) => {
                eprintln!("Failed to run binary: {}", e);
                process::exit(1);
            }
        }
    }
}

fn run_repl() {
    println!("Vex REPL — type :quit to exit");

    let mut interpreter = vex::interpreter::Interpreter::new();
    let mut accumulated = String::new();
    let mut prev_count: usize = 0;

    while let Some(input) = read_input() {
        let trimmed = input.trim();
        if trimmed == ":quit" || trimmed == ":exit" || trimmed == ":q" {
            break;
        }
        if trimmed.is_empty() {
            continue;
        }

        let test_source = if accumulated.is_empty() {
            input.clone()
        } else {
            format!("{}\n{}", accumulated, input)
        };

        let mut source_map = vex::source::SourceMap::new();
        let file_id = source_map.add_file("repl".into(), test_source.clone());
        let (tokens, lex_diags) = vex::lexer::lex(&test_source, file_id);
        if !lex_diags.is_empty() {
            for d in &lex_diags {
                eprintln!("{}", d.render(&source_map));
            }
            continue;
        }

        let (ast, parse_diags) = vex::parser::parse(&tokens);
        if !parse_diags.is_empty() {
            for d in &parse_diags {
                eprintln!("{}", d.render(&source_map));
            }
            continue;
        }

        let (ast, expand_diags) = vex::macro_expand::expand(ast);
        if !expand_diags.is_empty() {
            for d in &expand_diags {
                eprintln!("{}", d.render(&source_map));
            }
            continue;
        }

        let (hir_module, check_diags) = vex::typechecker::check(&ast);
        if !check_diags.is_empty() {
            for d in &check_diags {
                eprintln!("{}", d.render(&source_map));
            }
            continue;
        }

        accumulated = test_source;

        for form in &hir_module.top_forms[prev_count..] {
            match interpreter.eval_top_form(form) {
                Ok(vex::interpreter::Value::Unit) => {}
                Ok(value) => println!("=> {}", value),
                Err(err) => eprintln!("{}", err),
            }
        }
        prev_count = hir_module.top_forms.len();
    }
}

fn read_input() -> Option<String> {
    let mut input = String::new();
    let mut depth: i32 = 0;

    loop {
        if input.is_empty() {
            print!("vex> ");
        } else {
            print!("...  ");
        }
        io::stdout().flush().ok();

        let mut line = String::new();
        match io::stdin().read_line(&mut line) {
            Ok(0) => return None,
            Err(_) => return None,
            _ => {}
        }

        for ch in line.chars() {
            match ch {
                '(' | '[' | '{' => depth += 1,
                ')' | ']' | '}' => depth -= 1,
                _ => {}
            }
        }

        input.push_str(&line);

        if depth <= 0 {
            return Some(input);
        }
    }
}

fn tempdir() -> std::path::PathBuf {
    let mut path = env::temp_dir();
    path.push(format!("vex-build-{}", std::process::id()));
    fs::create_dir_all(&path).expect("failed to create temp directory");
    path
}
