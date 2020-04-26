import React from "react";

export function Timer(props) {
  // Parse the duration string (1h10m3s) into the total number of seconds.
  const total_solve_duration = ((duration) => {
    let re = /(?:(?<h>[0-9]+)h)?(?:(?<m>[0-9]+)m)?(?:(?<s>[0-9.]+)s)?/;
    let match = re.exec(duration);

    return (parseInt(match.groups.h || 0, 10)) * 3600 +
      (parseInt(match.groups.m || 0, 10)) * 60 +
      Math.round(parseFloat(match.groups.s || 0));
  })(props.total_solve_duration);

  // Convert the last start time into a timestamp as a number of seconds since
  // the epoch.  If there isn't a last start time, then this will return NaN.
  const last_start_time = Date.parse(props.last_start_time) / 1000;

  // Given an amount of time that the solve has gone for in the past as well as
  // time time that the current segment was started at, compute the total
  // duration in seconds that the solve has been going for.
  const compute = (total, start) => {
    if (!isNaN(start)) {
      total += new Date().getTime() / 1000 - start;
    }
    return Math.round(total);
  };

  const [total, setTotal] = React.useState(
    compute(total_solve_duration, last_start_time)
  );

  React.useEffect(() => {
    const interval = setInterval(() => {
      const total = compute(total_solve_duration, last_start_time);
      setTotal(total);
    }, 500);
    return () => clearInterval(interval)
  }, [total_solve_duration, last_start_time, setTotal]);

  return (
    <div className="timer">
      <Duration total={total}/>
    </div>
  );
}

function Duration(props) {
  const pad = (n) => {
    return (n < 10) ? "0" + n : n;
  };

  const total = props.total;
  const hours = Math.floor(total / 3600);
  const minutes = pad(Math.floor(total % 3600 / 60));
  const seconds = pad(Math.floor(total % 60));

  return (
    <React.Fragment>
      {`${hours}h ${minutes}m ${seconds}s`}
    </React.Fragment>
  );
}