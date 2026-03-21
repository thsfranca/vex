# Documentation Guidelines

## Core Principle

Write for humans first. Every sentence should be clear to someone reading the document for the first time.

## Voice and Tone

- Use active voice. Say "the parser produces an AST" instead of "an AST is produced by the parser."
- Address the reader directly with "you" when giving instructions.
- Keep sentences short. If a sentence has more than one idea, split it.

## Structure

- Prefer bullet points or numbered lists over dense paragraphs.
- Use numbered lists for sequential steps or ordered processes.
- Use bullet points for unordered collections of facts or rules.
- Start each section with a one-line summary before diving into details.
- Use headings (`##`, `###`) to break content into scannable sections.

## Formatting

- Use code blocks with language tags for code examples.
- Use backticks for inline references to files, functions, types, and commands.
- Use bold for key terms on first introduction.
- Avoid italics for emphasis — use bold or restructure the sentence instead.

## Content Rules

- Lead with the "what" and "why" before the "how."
- One idea per bullet point.
- Remove filler words: "basically," "simply," "just," "actually," "really."
- Don't explain what the reader can see from the code. Explain intent, trade-offs, and constraints.
- Keep examples concrete. Show real values, not abstract placeholders.

## What to Avoid

- Passive voice ("the file is read by the lexer" → "the lexer reads the file").
- Long paragraphs with multiple ideas packed together.
- Jargon without context — define terms the reader might not know.
- Redundant documentation that restates what the code already says.
