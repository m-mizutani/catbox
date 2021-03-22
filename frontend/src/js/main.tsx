import * as React from "react";
import * as ReactDOM from "react-dom";
import { Button } from "@material-ui/core";

class App extends React.Component {
  render() {
    return (
      <div>
        <h1>Hello React!</h1>
        <Button color="primary">Hey Hey Hey</Button>
      </div>
    );
  }
}

ReactDOM.render(<App />, document.querySelector("#app"));
