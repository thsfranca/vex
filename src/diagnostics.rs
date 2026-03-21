use std::fmt;

use crate::source::{SourceMap, Span};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Severity {
    Error,
    Warning,
}

impl fmt::Display for Severity {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Severity::Error => write!(f, "error"),
            Severity::Warning => write!(f, "warning"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Label {
    pub span: Span,
    pub message: String,
}

impl Label {
    pub fn new(span: Span, message: impl Into<String>) -> Self {
        Self {
            span,
            message: message.into(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Diagnostic {
    pub severity: Severity,
    pub message: String,
    pub span: Span,
    pub labels: Vec<Label>,
}

impl Diagnostic {
    pub fn error(message: impl Into<String>, span: Span) -> Self {
        Self {
            severity: Severity::Error,
            message: message.into(),
            span,
            labels: Vec::new(),
        }
    }

    pub fn warning(message: impl Into<String>, span: Span) -> Self {
        Self {
            severity: Severity::Warning,
            message: message.into(),
            span,
            labels: Vec::new(),
        }
    }

    pub fn with_label(mut self, label: Label) -> Self {
        self.labels.push(label);
        self
    }

    pub fn render(&self, source_map: &SourceMap) -> String {
        let loc = source_map.line_col(self.span.file, self.span.start);
        let file_name = source_map.name(self.span.file);
        let line_src = source_map.line_source(self.span.file, loc.line);
        let line_num = loc.line.to_string();
        let padding = " ".repeat(line_num.len());

        let underline_offset = " ".repeat((loc.col - 1) as usize);
        let underline_len = (self.span.len() as usize).max(1);
        let underline_char = match self.severity {
            Severity::Error => '^',
            Severity::Warning => '~',
        };
        let underline = underline_char.to_string().repeat(underline_len);

        let mut out = String::new();
        out.push_str(&format!("{}: {}\n", self.severity, self.message));
        out.push_str(&format!(
            "{padding} --> {file_name}:{line}:{col}\n",
            line = loc.line,
            col = loc.col
        ));
        out.push_str(&format!("{padding} |\n"));
        out.push_str(&format!("{line_num} | {line_src}\n"));
        out.push_str(&format!("{padding} | {underline_offset}{underline}"));

        for label in &self.labels {
            let label_loc = source_map.line_col(label.span.file, label.span.start);
            let label_file = source_map.name(label.span.file);
            let label_line_src = source_map.line_source(label.span.file, label_loc.line);
            let label_line_num = label_loc.line.to_string();
            let label_padding = " ".repeat(label_line_num.len());
            let label_underline_offset = " ".repeat((label_loc.col - 1) as usize);
            let label_underline_len = (label.span.len() as usize).max(1);
            let label_underline = "~".repeat(label_underline_len);

            out.push('\n');
            out.push_str(&format!(
                "{label_padding} --> {label_file}:{line}:{col}\n",
                line = label_loc.line,
                col = label_loc.col
            ));
            out.push_str(&format!("{label_padding} |\n"));
            out.push_str(&format!("{label_line_num} | {label_line_src}\n"));
            out.push_str(&format!(
                "{label_padding} | {label_underline_offset}{label_underline} {}",
                label.message
            ));
        }

        out
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::source::{FileId, SourceMap, Span};

    fn setup() -> (SourceMap, FileId) {
        let mut map = SourceMap::new();
        let id = map.add_file(
            "test.vx".into(),
            "(defn main []\n  (println \"hi\"))".into(),
        );
        (map, id)
    }

    #[test]
    fn error_render_single_line() {
        let (map, id) = setup();
        let span = Span::new(id, 1, 5);
        let diag = Diagnostic::error("unexpected token", span);
        let rendered = diag.render(&map);

        assert!(rendered.contains("error: unexpected token"));
        assert!(rendered.contains("test.vx:1:2"));
        assert!(rendered.contains("(defn main []"));
        assert!(rendered.contains("^^^^"));
    }

    #[test]
    fn warning_render() {
        let (map, id) = setup();
        let span = Span::new(id, 1, 5);
        let diag = Diagnostic::warning("unused variable", span);
        let rendered = diag.render(&map);

        assert!(rendered.contains("warning: unused variable"));
        assert!(rendered.contains("~~~~"));
    }

    #[test]
    fn render_with_label() {
        let (map, id) = setup();
        let span = Span::new(id, 1, 5);
        let label = Label::new(Span::new(id, 16, 23), "first defined here");
        let diag = Diagnostic::error("duplicate definition", span).with_label(label);
        let rendered = diag.render(&map);

        assert!(rendered.contains("error: duplicate definition"));
        assert!(rendered.contains("first defined here"));
        assert!(rendered.contains("~~~~~~~"));
    }

    #[test]
    fn render_second_line() {
        let (map, id) = setup();
        let span = Span::new(id, 16, 23);
        let diag = Diagnostic::error("unknown function", span);
        let rendered = diag.render(&map);

        assert!(rendered.contains("test.vx:2:3"));
        assert!(rendered.contains("println"));
        assert!(rendered.contains("^^^^^^^"));
    }

    #[test]
    fn severity_display() {
        assert_eq!(format!("{}", Severity::Error), "error");
        assert_eq!(format!("{}", Severity::Warning), "warning");
    }

    #[test]
    fn diagnostic_constructors() {
        let span = Span::new(FileId::new(0), 0, 1);
        let err = Diagnostic::error("bad", span);
        assert_eq!(err.severity, Severity::Error);
        assert_eq!(err.message, "bad");
        assert!(err.labels.is_empty());

        let warn = Diagnostic::warning("meh", span);
        assert_eq!(warn.severity, Severity::Warning);
    }
}
