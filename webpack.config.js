module.exports = {
    entry: './react/app.js',
    output: {
        path: __dirname + '/resources/js',
        filename: 'app.js',
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: {
                    loader: "babel-loader"
                }
            }
        ]
    }
};