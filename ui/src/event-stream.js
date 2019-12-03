// An EventStream is an encapsulation around an EventSource that ensures
// that the client remains connected to the EventSource.  It watches for
// errors that may occur, and reconnects when one happens.
export class EventStream {
  constructor(url) {
    this.url = url;
    this.source = null;
  }

  setHandler(handler) {
    this.handler = handler;
    this.start();
  }

  start() {
    if (this.handler === null) {
      throw new Error("must set a handler before starting");
    }

    if (this.source) {
      this.source.close();
    }

    const start = this.start.bind(this);
    const source = new EventSource(this.url);
    source.onerror = () => {
      this.source.close();
      this.source = null;
      setTimeout(start, 2000);
    };
    source.onmessage = this.handler;
    this.source = source;
  }
}