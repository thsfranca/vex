const vscode = require('vscode');

function activate(context) {
    console.log('Fugo extension activated successfully!');
    
    // Simple test command
    let disposable = vscode.commands.registerCommand('fugo.test', function () {
        vscode.window.showInformationMessage('Fugo extension is working!');
    });

    context.subscriptions.push(disposable);
}

function deactivate() {
    console.log('Fugo extension deactivated');
}

module.exports = {
    activate,
    deactivate
};