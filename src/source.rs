#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct FileId(u32);

impl FileId {
    pub fn new(id: u32) -> Self {
        Self(id)
    }

    pub fn index(self) -> usize {
        self.0 as usize
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Span {
    pub file: FileId,
    pub start: u32,
    pub end: u32,
}

impl Span {
    pub fn new(file: FileId, start: u32, end: u32) -> Self {
        Self { file, start, end }
    }

    pub fn len(self) -> u32 {
        self.end - self.start
    }

    pub fn is_empty(self) -> bool {
        self.start == self.end
    }
}

struct FileEntry {
    name: String,
    source: String,
    line_offsets: Vec<u32>,
}

impl FileEntry {
    fn new(name: String, source: String) -> Self {
        let line_offsets = std::iter::once(0)
            .chain(
                source
                    .bytes()
                    .enumerate()
                    .filter(|&(_, b)| b == b'\n')
                    .map(|(i, _)| (i + 1) as u32),
            )
            .collect();
        Self {
            name,
            source,
            line_offsets,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LineCol {
    pub line: u32,
    pub col: u32,
}

#[derive(Default)]
pub struct SourceMap {
    files: Vec<FileEntry>,
}

impl SourceMap {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_file(&mut self, name: String, source: String) -> FileId {
        let id = FileId::new(self.files.len() as u32);
        self.files.push(FileEntry::new(name, source));
        id
    }

    pub fn name(&self, file: FileId) -> &str {
        &self.files[file.index()].name
    }

    pub fn source(&self, file: FileId) -> &str {
        &self.files[file.index()].source
    }

    pub fn line_col(&self, file: FileId, offset: u32) -> LineCol {
        let entry = &self.files[file.index()];
        let line = entry
            .line_offsets
            .partition_point(|&o| o <= offset)
            .saturating_sub(1);
        let col = offset - entry.line_offsets[line];
        LineCol {
            line: line as u32 + 1,
            col: col + 1,
        }
    }

    pub fn line_source(&self, file: FileId, line: u32) -> &str {
        let entry = &self.files[file.index()];
        let idx = (line - 1) as usize;
        let start = entry.line_offsets[idx] as usize;
        let end = entry
            .line_offsets
            .get(idx + 1)
            .map(|&o| o as usize)
            .unwrap_or(entry.source.len());
        entry.source[start..end].trim_end_matches('\n')
    }

    pub fn span_text(&self, span: Span) -> &str {
        let source = self.source(span.file);
        &source[span.start as usize..span.end as usize]
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn file_id_roundtrip() {
        let id = FileId::new(3);
        assert_eq!(id.index(), 3);
    }

    #[test]
    fn span_len_and_empty() {
        let file = FileId::new(0);
        let span = Span::new(file, 5, 10);
        assert_eq!(span.len(), 5);
        assert!(!span.is_empty());

        let empty = Span::new(file, 5, 5);
        assert_eq!(empty.len(), 0);
        assert!(empty.is_empty());
    }

    #[test]
    fn source_map_add_and_lookup() {
        let mut map = SourceMap::new();
        let id = map.add_file("test.vx".into(), "(println \"hi\")".into());
        assert_eq!(map.name(id), "test.vx");
        assert_eq!(map.source(id), "(println \"hi\")");
    }

    #[test]
    fn line_col_single_line() {
        let mut map = SourceMap::new();
        let id = map.add_file("test.vx".into(), "(println \"hi\")".into());
        assert_eq!(map.line_col(id, 0), LineCol { line: 1, col: 1 });
        assert_eq!(map.line_col(id, 1), LineCol { line: 1, col: 2 });
    }

    #[test]
    fn line_col_multi_line() {
        let mut map = SourceMap::new();
        let src = "(defn main []\n  (println \"hi\"))";
        let id = map.add_file("test.vx".into(), src.into());

        assert_eq!(map.line_col(id, 0), LineCol { line: 1, col: 1 });
        assert_eq!(map.line_col(id, 14), LineCol { line: 2, col: 1 });
        assert_eq!(map.line_col(id, 16), LineCol { line: 2, col: 3 });
    }

    #[test]
    fn line_source_returns_correct_line() {
        let mut map = SourceMap::new();
        let src = "(defn main []\n  (println \"hi\"))";
        let id = map.add_file("test.vx".into(), src.into());

        assert_eq!(map.line_source(id, 1), "(defn main []");
        assert_eq!(map.line_source(id, 2), "  (println \"hi\"))");
    }

    #[test]
    fn span_text_extracts_slice() {
        let mut map = SourceMap::new();
        let id = map.add_file("test.vx".into(), "(println \"hi\")".into());
        let span = Span::new(id, 1, 8);
        assert_eq!(map.span_text(span), "println");
    }

    #[test]
    fn multiple_files() {
        let mut map = SourceMap::new();
        let a = map.add_file("a.vx".into(), "aaa".into());
        let b = map.add_file("b.vx".into(), "bbb".into());

        assert_eq!(map.name(a), "a.vx");
        assert_eq!(map.name(b), "b.vx");
        assert_eq!(map.source(a), "aaa");
        assert_eq!(map.source(b), "bbb");
    }
}
