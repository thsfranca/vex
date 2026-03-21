use std::env;
use std::fs;
use std::path::Path;
use std::process::{self, Command};

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 3 {
        eprintln!("Usage: vex build <file.vx> [-o <output>] [--emit-go <dir>]");
        process::exit(3);
    }

    let subcommand = &args[1];
    if subcommand != "build" && subcommand != "run" {
        eprintln!("Unknown command: {}", subcommand);
        eprintln!("Usage: vex build <file.vx> [-o <output>] [--emit-go <dir>]");
        process::exit(3);
    }

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

fn tempdir() -> std::path::PathBuf {
    let mut path = env::temp_dir();
    path.push(format!("vex-build-{}", std::process::id()));
    fs::create_dir_all(&path).expect("failed to create temp directory");
    path
}
