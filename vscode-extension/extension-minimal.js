const vscode = require('vscode');

function activate(context) {
    console.log('Vex extension activated successfully!');
    
    // Simple test command
    let disposable = vscode.commands.registerCommand('vex.test', function () {
        vscode.window.showInformationMessage('Vex extension is working!');
    });

    context.subscriptions.push(disposable);
}

function deactivate() {
    console.log('Vex extension deactivated');
}

module.exports = {
    activate,
    deactivate
};