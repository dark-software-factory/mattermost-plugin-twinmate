const path = require('node:path');

module.exports = {
    entry: './src/index.js',
    output: {path: path.resolve(__dirname, 'dist'), filename: 'main.js'},
    resolve: {extensions: ['.js']},
    externals: {react: 'React'},
    target: 'web',
};
