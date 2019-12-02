import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "./nav.css";

export function Nav() {
  return (
    <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
      <div className="navbar-brand">Twitch Plays Crosswords</div>
    </nav>
  );
}
