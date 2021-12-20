import express from "express";

export default class Server {
  constructor() {
    this._app = express();
    this._app.use(express.json());
  }

  defineRoute(path, router) {
    this._app.use(path, router);
  }

  start(port) {
    this._server = this._app.listen(port, () => {
      console.log(`Server running on port ${port}`);
    });
  }

  stop() {
    if (this._server) {
      this._server.close();
      this._server = null;
    }
  }
}
