pub mod ast;
pub mod builtins;
pub mod codegen;
pub mod diagnostics;
pub mod hir;
pub mod lexer;
pub mod parser;
pub mod source;
pub mod typechecker;
pub mod types;

use diagnostics::Diagnostic;
use source::SourceMap;

pub struct VexrtFiles {
    pub option_go: String,
    pub result_go: String,
}

pub struct CompileResult {
    pub go_source: String,
    pub go_mod: String,
    pub vexrt: Option<VexrtFiles>,
    pub diagnostics: Vec<Diagnostic>,
    pub source_map: SourceMap,
}

pub fn compile(source: &str, file_name: &str) -> CompileResult {
    let mut source_map = SourceMap::new();
    let file_id = source_map.add_file(file_name.to_string(), source.to_string());
    let (tokens, lex_diags) = lexer::lex(source, file_id);

    if !lex_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            diagnostics: lex_diags,
            source_map,
        };
    }

    let (ast, parse_diags) = parser::parse(&tokens);

    if !parse_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            diagnostics: parse_diags,
            source_map,
        };
    }

    let (hir_module, check_diags) = typechecker::check(&ast);

    if !check_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            diagnostics: check_diags,
            source_map,
        };
    }

    let go_source = codegen::generate(&hir_module);
    let go_mod = codegen::generate_go_mod();

    let vexrt = if codegen::needs_vexrt(&hir_module) {
        Some(VexrtFiles {
            option_go: codegen::generate_vexrt_option(),
            result_go: codegen::generate_vexrt_result(),
        })
    } else {
        None
    };

    CompileResult {
        go_source,
        go_mod,
        vexrt,
        diagnostics: Vec::new(),
        source_map,
    }
}
