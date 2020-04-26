// An EventStream is an encapsulation around an EventSource that ensures
// that the client remains connected to the EventSource.  It watches for
// errors that may occur, and reconnects when one happens.
export class EventStream {
  constructor(url) {
    this.url = url;
    this.source = null;
    this.watchdog = null;
  }

  setHandler(handler) {
    this.handler = handler;
    this.stop();
    this.start();
  }

  start() {
    if (this.handler === null) {
      throw new Error("must set a handler before starting");
    }

    const start = this.start.bind(this);
    const stop = this.stop.bind(this);
    const source = new EventSource(this.url);
    source.onerror = () => {
      stop();
      setTimeout(start, 2000);
    };

    // Keep track of whether or not we've received data since the last watchdog
    // check.
    let received = false;
    source.onmessage = (message) => {
      received = true;
      this.handler(message);
    };
    this.source = source;

    // Wakeup every minute and if we haven't received a message then restart
    // the stream.
    this.watchdog = setInterval(() => {
      if (received) {
        received = false;
        return;
      }

      console.log("message hasn't been received in last 60 seconds, restarting event stream");
      stop();
      setTimeout(start, 100);
    }, 60000);
  }

  stop() {
    if (this.watchdog !== null) {
      clearInterval(this.watchdog);
      this.watchdog = null;
    }

    if (this.source !== null) {
      this.source.close();
      this.source = null;
    }
  }
}