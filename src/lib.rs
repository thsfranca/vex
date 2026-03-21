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
use std::path::{Path, PathBuf};
use types::VexType;

pub struct VexrtFiles {
    pub option_go: String,
    pub result_go: String,
    pub collections_go: String,
}

pub struct GoPackage {
    pub name: String,
    pub source: String,
}

pub struct CompileResult {
    pub go_source: String,
    pub go_mod: String,
    pub vexrt: Option<VexrtFiles>,
    pub extra_packages: Vec<GoPackage>,
    pub diagnostics: Vec<Diagnostic>,
    pub source_map: SourceMap,
}

pub fn resolve_module_path(source_dir: &Path, module_name: &str) -> PathBuf {
    let parts: Vec<&str> = module_name.split('.').collect();
    let mut path = source_dir.to_path_buf();
    for part in &parts[..parts.len() - 1] {
        path.push(part);
    }
    path.push(format!("{}.vx", parts[parts.len() - 1]));
    path
}

pub fn collect_imports(program: &[ast::TopForm]) -> Vec<(String, Vec<String>)> {
    program
        .iter()
        .filter_map(|form| match form {
            ast::TopForm::Import {
                module_path,
                symbols,
                ..
            } => Some((module_path.clone(), symbols.clone())),
            _ => None,
        })
        .collect()
}

pub fn collect_exports(program: &[ast::TopForm]) -> Vec<String> {
    program
        .iter()
        .filter_map(|form| match form {
            ast::TopForm::Export { symbols, .. } => Some(symbols.clone()),
            _ => None,
        })
        .flatten()
        .collect()
}

pub fn extract_exported_types(
    hir_module: &hir::Module,
    exported_names: &[String],
) -> Vec<(String, VexType)> {
    let mut result = Vec::new();
    for form in &hir_module.top_forms {
        match form {
            hir::TopForm::Defn {
                name,
                params,
                return_type,
                ..
            } => {
                if exported_names.contains(name) {
                    let fn_type = VexType::Fn {
                        params: params.iter().map(|p| p.ty.clone()).collect(),
                        ret: Box::new(return_type.clone()),
                    };
                    result.push((name.clone(), fn_type));
                }
            }
            hir::TopForm::Def { name, ty, .. } => {
                if exported_names.contains(name) {
                    result.push((name.clone(), ty.clone()));
                }
            }
            _ => {}
        }
    }
    result
}

fn compile_single(
    source: &str,
    file_name: &str,
    source_map: &mut SourceMap,
    imported_symbols: &[(String, VexType)],
) -> Result<(Vec<ast::TopForm>, hir::Module), Vec<Diagnostic>> {
    let file_id = source_map.add_file(file_name.to_string(), source.to_string());
    let (tokens, lex_diags) = lexer::lex(source, file_id);
    if !lex_diags.is_empty() {
        return Err(lex_diags);
    }

    let (ast, parse_diags) = parser::parse(&tokens);
    if !parse_diags.is_empty() {
        return Err(parse_diags);
    }

    let (hir_module, check_diags) = typechecker::check_with_imports(&ast, imported_symbols);
    if !check_diags.is_empty() {
        return Err(check_diags);
    }

    Ok((ast, hir_module))
}

pub fn compile(source: &str, file_name: &str) -> CompileResult {
    let mut source_map = SourceMap::new();
    let mut all_diagnostics = Vec::new();
    let mut extra_packages = Vec::new();
    let mut imported_symbols: Vec<(String, VexType)> = Vec::new();

    let source_path = Path::new(file_name);
    let source_dir = source_path.parent().unwrap_or(Path::new("."));

    let file_id = source_map.add_file(file_name.to_string(), source.to_string());
    let (tokens, lex_diags) = lexer::lex(source, file_id);
    if !lex_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            extra_packages: Vec::new(),
            diagnostics: lex_diags,
            source_map,
        };
    }

    let (main_ast, parse_diags) = parser::parse(&tokens);
    if !parse_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            extra_packages: Vec::new(),
            diagnostics: parse_diags,
            source_map,
        };
    }

    let imports = collect_imports(&main_ast);
    for (module_path, symbols) in &imports {
        let module_file = resolve_module_path(source_dir, module_path);
        let module_source = match std::fs::read_to_string(&module_file) {
            Ok(s) => s,
            Err(_) => {
                all_diagnostics.push(Diagnostic::error(
                    format!(
                        "could not find module '{}' (looked for {})",
                        module_path,
                        module_file.display()
                    ),
                    source::Span::new(file_id, 0, 0),
                ));
                continue;
            }
        };

        let module_file_name = module_file.display().to_string();
        match compile_single(&module_source, &module_file_name, &mut source_map, &[]) {
            Ok((dep_ast, dep_hir)) => {
                let exported = collect_exports(&dep_ast);
                for sym in symbols {
                    if !exported.contains(sym) {
                        all_diagnostics.push(Diagnostic::error(
                            format!(
                                "symbol '{}' is not exported by module '{}'",
                                sym, module_path
                            ),
                            source::Span::new(file_id, 0, 0),
                        ));
                    }
                }

                let types = extract_exported_types(&dep_hir, symbols);
                imported_symbols.extend(types);

                let pkg_name = module_path.replace('.', "/");
                let go_source = codegen::generate_package(&dep_hir, &pkg_name);
                extra_packages.push(GoPackage {
                    name: pkg_name,
                    source: go_source,
                });
            }
            Err(diags) => {
                all_diagnostics.extend(diags);
            }
        }
    }

    if !all_diagnostics.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            extra_packages: Vec::new(),
            diagnostics: all_diagnostics,
            source_map,
        };
    }

    let import_map: std::collections::HashMap<String, String> = imports
        .iter()
        .flat_map(|(module_path, symbols)| {
            symbols
                .iter()
                .map(move |sym| (sym.clone(), module_path.replace('.', "/")))
        })
        .collect();

    let (hir_module, check_diags) = typechecker::check_with_imports(&main_ast, &imported_symbols);

    if !check_diags.is_empty() {
        return CompileResult {
            go_source: String::new(),
            go_mod: String::new(),
            vexrt: None,
            extra_packages: Vec::new(),
            diagnostics: check_diags,
            source_map,
        };
    }

    let go_source = codegen::generate_with_imports(&hir_module, &import_map);
    let go_mod = codegen::generate_go_mod();

    let needs_rt = codegen::needs_vexrt(&hir_module);

    let vexrt = if needs_rt {
        Some(VexrtFiles {
            option_go: codegen::generate_vexrt_option(),
            result_go: codegen::generate_vexrt_result(),
            collections_go: codegen::generate_vexrt_collections(),
        })
    } else {
        None
    };

    CompileResult {
        go_source,
        go_mod,
        vexrt,
        extra_packages,
        diagnostics: Vec::new(),
        source_map,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn resolve_simple_module() {
        let path = resolve_module_path(Path::new("/project/src"), "math");
        assert_eq!(path, PathBuf::from("/project/src/math.vx"));
    }

    #[test]
    fn resolve_qualified_module() {
        let path = resolve_module_path(Path::new("/project/src"), "vex.http.server");
        assert_eq!(path, PathBuf::from("/project/src/vex/http/server.vx"));
    }

    #[test]
    fn resolve_relative_dir() {
        let path = resolve_module_path(Path::new("."), "utils");
        assert_eq!(path, PathBuf::from("./utils.vx"));
    }

    #[test]
    fn collect_imports_from_ast() {
        let program = vec![
            ast::TopForm::Module {
                name: "myapp".into(),
                span: source::Span::new(source::FileId::new(0), 0, 14),
            },
            ast::TopForm::Import {
                module_path: "math".into(),
                symbols: vec!["add".into(), "sub".into()],
                span: source::Span::new(source::FileId::new(0), 15, 35),
            },
        ];
        let imports = collect_imports(&program);
        assert_eq!(imports.len(), 1);
        assert_eq!(imports[0].0, "math");
        assert_eq!(imports[0].1, vec!["add", "sub"]);
    }

    #[test]
    fn collect_imports_empty() {
        let program = vec![ast::TopForm::Expr(ast::Expr::Int(
            42,
            source::Span::new(source::FileId::new(0), 0, 2),
        ))];
        let imports = collect_imports(&program);
        assert!(imports.is_empty());
    }

    #[test]
    fn collect_exports_from_ast() {
        let program = vec![ast::TopForm::Export {
            symbols: vec!["foo".into(), "bar".into()],
            span: source::Span::new(source::FileId::new(0), 0, 18),
        }];
        let exports = collect_exports(&program);
        assert_eq!(exports, vec!["foo", "bar"]);
    }

    #[test]
    fn extract_defn_types() {
        let hir_module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "add".into(),
                params: vec![
                    hir::Param {
                        name: "a".into(),
                        ty: VexType::Int,
                        span: source::Span::new(source::FileId::new(0), 0, 1),
                    },
                    hir::Param {
                        name: "b".into(),
                        ty: VexType::Int,
                        span: source::Span::new(source::FileId::new(0), 0, 1),
                    },
                ],
                return_type: VexType::Int,
                body: vec![hir::Expr::Int(
                    0,
                    source::Span::new(source::FileId::new(0), 0, 1),
                )],
                span: source::Span::new(source::FileId::new(0), 0, 10),
            }],
        };
        let types = extract_exported_types(&hir_module, &["add".into()]);
        assert_eq!(types.len(), 1);
        assert_eq!(types[0].0, "add");
        assert_eq!(
            types[0].1,
            VexType::Fn {
                params: vec![VexType::Int, VexType::Int],
                ret: Box::new(VexType::Int),
            }
        );
    }

    #[test]
    fn extract_def_types() {
        let hir_module = hir::Module {
            top_forms: vec![hir::TopForm::Def {
                name: "pi".into(),
                ty: VexType::Float,
                value: hir::Expr::Float(3.14, source::Span::new(source::FileId::new(0), 0, 4)),
                span: source::Span::new(source::FileId::new(0), 0, 10),
            }],
        };
        let types = extract_exported_types(&hir_module, &["pi".into()]);
        assert_eq!(types.len(), 1);
        assert_eq!(types[0].0, "pi");
        assert_eq!(types[0].1, VexType::Float);
    }

    #[test]
    fn extract_skips_non_exported() {
        let hir_module = hir::Module {
            top_forms: vec![hir::TopForm::Defn {
                name: "private".into(),
                params: vec![],
                return_type: VexType::Unit,
                body: vec![hir::Expr::Nil(source::Span::new(
                    source::FileId::new(0),
                    0,
                    3,
                ))],
                span: source::Span::new(source::FileId::new(0), 0, 10),
            }],
        };
        let types = extract_exported_types(&hir_module, &["add".into()]);
        assert!(types.is_empty());
    }
}
