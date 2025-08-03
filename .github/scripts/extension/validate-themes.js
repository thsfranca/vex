#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

try {
    // Validate color themes
    const themesPath = path.join(process.cwd(), 'vscode-extension', 'themes');
    
    if (!fs.existsSync(themesPath)) {
        console.log('[INFO] No themes directory found - skipping theme validation');
        process.exit(0);
    }
    
    const themeFiles = fs.readdirSync(themesPath).filter(f => f.endsWith('.json'));
    
    if (themeFiles.length === 0) {
        console.log('[INFO] No theme JSON files found - skipping theme validation');
        process.exit(0);
    }
    
    themeFiles.forEach(file => {
        const themePath = path.join(themesPath, file);
        console.log(`[CHECK] Validating theme: ${file}`);
        
        const theme = JSON.parse(fs.readFileSync(themePath, 'utf8'));
        console.log('[SUCCESS] Theme JSON valid:', file);
        
        if (theme.colors) {
            console.log('🎨 Color definitions found:', Object.keys(theme.colors).length);
        }
        if (theme.tokenColors) {
            console.log('🔤 Token color rules found:', theme.tokenColors.length);
        }
    });
    
} catch (error) {
    console.error('[ERROR] Failed to validate themes:', error.message);
    process.exit(1);
}