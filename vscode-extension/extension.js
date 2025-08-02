const vscode = require('vscode');
const { exec } = require('child_process');
const path = require('path');

function activate(context) {
    console.log('Fugo extension is now active!');

    let disposable = vscode.commands.registerCommand('fugo.transpile', function () {
        const editor = vscode.window.activeTextEditor;
        
        if (!editor) {
            vscode.window.showErrorMessage('No active editor found');
            return;
        }

        if (editor.document.languageId !== 'fugo') {
            vscode.window.showErrorMessage('Current file is not a Fugo file');
            return;
        }

        const fugoCode = editor.document.getText();
        
        // Find workspace root to locate the transpiler
        const workspaceFolder = vscode.workspace.getWorkspaceFolder(editor.document.uri);
        if (!workspaceFolder) {
            vscode.window.showErrorMessage('No workspace folder found');
            return;
        }

        const workspacePath = workspaceFolder.uri.fsPath;
        
        // Create a temporary transpiler command
        const tempFile = path.join(workspacePath, 'temp_transpile.go');
        const transpilerCode = `
package main

import (
    "fmt"
    "fugo/pkg/transpiler"
)

func main() {
    t := transpiler.New()
    result, err := t.TranspileToGo(\`${fugoCode.replace(/`/g, '\\`')}\`)
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        return
    }
    fmt.Print(result)
}
`;

        require('fs').writeFileSync(tempFile, transpilerCode);

        // Run the transpiler
        exec(`cd "${workspacePath}" && go run ${tempFile}`, (error, stdout, stderr) => {
            // Clean up temp file
            try {
                require('fs').unlinkSync(tempFile);
            } catch (e) {
                // Ignore cleanup errors
            }

            if (error) {
                vscode.window.showErrorMessage(\`Transpilation failed: \${error.message}\`);
                return;
            }

            if (stderr) {
                vscode.window.showErrorMessage(\`Transpiler stderr: \${stderr}\`);
                return;
            }

            // Show the transpiled Go code in a new document
            vscode.workspace.openTextDocument({
                content: stdout,
                language: 'go'
            }).then(doc => {
                vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
            });
        });
    });

    context.subscriptions.push(disposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
};