#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

try {
    // Validate TextMate grammar
    const grammarPath = path.join(process.cwd(), 'vscode-extension', 'syntaxes');
    
    if (!fs.existsSync(grammarPath)) {
        console.log('[INFO] No syntaxes directory found - skipping grammar validation');
        process.exit(0);
    }
    
    const grammarFiles = fs.readdirSync(grammarPath).filter(f => f.endsWith('.tmLanguage.json'));
    
    if (grammarFiles.length === 0) {
        console.log('[INFO] No .tmLanguage.json files found - skipping grammar validation');
        process.exit(0);
    }
    
    grammarFiles.forEach(file => {
        const grammarFile = path.join(grammarPath, file);
        console.log(`[CHECK] Validating grammar: ${file}`);
        
        const grammar = JSON.parse(fs.readFileSync(grammarFile, 'utf8'));
        console.log('[SUCCESS] TextMate grammar is valid JSON');
        
        // Check required grammar fields
        const required = ['name', 'scopeName', 'patterns'];
        const missing = required.filter(field => !grammar[field]);
        
        if (missing.length > 0) {
            console.log('[ERROR] Missing required grammar fields:', missing);
            process.exit(1);
        }
        
        console.log('[SUCCESS] TextMate grammar structure valid');
        console.log('🔤 Grammar name:', grammar.name);
        console.log('🏷️  Scope name:', grammar.scopeName);
        
        if (grammar.patterns) {
            console.log('📝 Pattern rules found:', grammar.patterns.length);
        }
    });
    
} catch (error) {
    console.error('[ERROR] Failed to validate grammar:', error.message);
    process.exit(1);
}