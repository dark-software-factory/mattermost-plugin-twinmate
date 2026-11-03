const fs = require('node:fs');
const path = require('node:path');

const stateFile = path.join(__dirname, '..', '.runtime', 'state.json');

function saveState(value) {
    fs.mkdirSync(path.dirname(stateFile), {recursive: true});
    fs.writeFileSync(stateFile, JSON.stringify(value), {mode: 0o600});
}
function loadState() {
    return JSON.parse(fs.readFileSync(stateFile, 'utf8'));
}
function clearState() {
    fs.rmSync(stateFile, {force: true});
}

module.exports = {saveState, loadState, clearState};
