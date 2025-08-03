#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

try {
    // Read and parse package.json
    const packagePath = path.join(process.cwd(), 'vscode-extension', 'package.json');
    const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    
    console.log('[SUCCESS] package.json is valid JSON');
    
    // Check required fields for VSCode extension
    const required = ['name', 'version', 'engines', 'contributes'];
    const missing = required.filter(field => !pkg[field]);
    
    if (missing.length > 0) {
        console.log('[ERROR] Missing required fields:', missing);
        process.exit(1);
    }
    
    console.log('[SUCCESS] Required extension fields present');
    console.log('📦 Extension:', pkg.name, 'v' + pkg.version);
    
} catch (error) {
    console.error('[ERROR] Failed to validate package.json:', error.message);
    process.exit(1);
}