module.exports = {
  mode: "development",
  entry: "./src/js/main.tsx",
  output: {
    path: `${__dirname}/dist`,
    filename: "bundle.js",
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: "ts-loader",
      },
      {
        test: /\.js?$/,
        exclude: /node_modules/,
        use: "babel-loader",
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        use: "babel-loader",
      },
    ],
  },
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".jsx", ".json"],
  },
  devServer: {
    contentBase: "dist",
    proxy: {
      "/auth": "http://localhost:9080",
      "/api": "http://localhost:9080",
    },
  },
  target: ["web", "es5"],
};
