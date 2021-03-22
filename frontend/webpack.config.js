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
    ],
  },
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".json"],
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
