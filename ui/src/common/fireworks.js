import React from "react";
import "./fireworks.css";

// This implementation is based heavily on https://jsfiddle.net/XWMpq/.

export class Fireworks extends React.Component {
  static maxNumRockets = 15;
  static screen = {
    w: window.innerWidth,
    h: window.innerHeight,
  };
  static fps = 60;

  constructor(props) {
    super(props);
    this.canvasRef = React.createRef();
  }

  componentDidMount() {
    this.launchInterval = setInterval(() => this.launch(), 1000);
    this.updateInterval = setInterval(() => this.update(), 1000/Fireworks.fps);

    this.rockets = [];
    this.particles = [];

    const canvas = this.canvasRef.current;
    canvas.width = Fireworks.screen.w;
    canvas.height = Fireworks.screen.h;
  }

  componentWillUnmount() {
    clearInterval(this.updateInterval);
    clearInterval(this.launchInterval);
  }

  launch() {
    if (this.rockets.length < Fireworks.maxNumRockets) {
      // Determine how many rockets to launch.
      const n = 1 + Math.random() * (Fireworks.maxNumRockets - this.rockets.length);
      for (let i = 0; i < n; i++) {
        const x = Fireworks.screen.w * Math.random();

        const rocket = new Rocket(x, Fireworks.screen.h);
        rocket.explosionColor = Math.floor(Math.random() * 360 / 10) * 10;
        rocket.dx = 6 * Math.random() - 3;
        rocket.dy = -3 * Math.random() - 4;
        rocket.size = 8;
        rocket.shrink = 0.999;
        rocket.gravity = 0.01;
        this.rockets.push(rocket);
      }
    }
  }

  update() {
    const canvas = this.canvasRef.current;

    // Update the canvas size if the screen dimensions have changed.
    if (Fireworks.screen.w !== window.innerWidth || Fireworks.screen.h !== window.innerHeight) {
      Fireworks.screen.w = window.innerWidth;
      Fireworks.screen.h = window.innerHeight;
      canvas.width = Fireworks.screen.w;
      canvas.height = Fireworks.screen.h;
    }

    const context = canvas.getContext("2d");
    context.fillStyle = "rgba(0, 0, 0, 0.05)";
    context.fillRect(0, 0, Fireworks.screen.w, Fireworks.screen.h);

    // Update the rockets
    const newRockets = [];
    for (let i = 0; i < this.rockets.length; i++) {
      const rocket = this.rockets[i];

      rocket.update();
      rocket.render(context);

      // Determine if we should add smoke.
      if (Math.random() < 0.6) {
        const smoke = new Smoke(rocket.x, rocket.y);
        smoke.dx = Math.random() - 0.5;
        this.particles.push(smoke);
      }

      // Slow the rocket down as it gets near the top of the screen.
      if (rocket.y < Fireworks.screen.h/3) {
        rocket.resistance = 0.98;
      }

      // Determine whether or not this rocket should explode.  A rocket explodes
      // if it gets too close to the top of the screen, or its close to its
      // apex.
      if (rocket.y < Fireworks.screen.h/6 || Math.abs(rocket.dy) <= 1) {
        const explosionParticles = rocket.explode();
        for (let j = 0; j < explosionParticles.length; j++) {
          this.particles.push(explosionParticles[j]);
        }
        continue;
      }

      newRockets.push(rocket);
    }
    this.rockets = newRockets;

    // Update the fireworks
    const newParticles = [];
    for (let i = 0; i < this.particles.length; i++) {
      const particle = this.particles[i];
      particle.update();

      if (particle.exists()) {
        particle.render(context);
        newParticles.push(particle);
      }
    }
    this.particles = newParticles;
  }

  render() {
    return (<canvas className="fireworks" ref={this.canvasRef}/>);
  }
}

class Particle {
  constructor(x, y) {
    this.x = x || 0;
    this.y = y || 0;
    this.dx = 0;
    this.dy = 0;

    this.shrink = 0.97;
    this.size = 2;

    this.resistance = 1;
    this.gravity = 0;

    this.flick = false;

    this.alpha = 1;
    this.fade = 0;
    this.color = 0;
  }

  update() {
    // apply resistance
    this.dx *= this.resistance;
    this.dy *= this.resistance;

    // gravity down
    this.dy += this.gravity;

    // update position based on speed
    this.x += this.dx;
    this.y += this.dy;

    // shrink
    this.size *= this.shrink;

    // fade out
    this.alpha -= this.fade;
  }

  render(c) {
    if (!this.exists()) {
      return;
    }

    const x = this.x;
    const y = this.y;
    const r = this.size / 2;
    const radius = this.size * (this.flick ? Math.random() : 1);

    c.save();
    c.globalCompositeOperation = "lighter";

    const gradient = c.createRadialGradient(x, y, 0.1, x, y, r);
    gradient.addColorStop(0.1, `rgba(255, 255, 255, ${this.alpha})`);
    gradient.addColorStop(0.8, `hsla(${this.color}, 100%, 50%, ${this.alpha})`);
    gradient.addColorStop(1, `hsla(${this.color}, 100%, 50%, 0.1)`);
    c.fillStyle = gradient;

    c.beginPath();
    c.arc(x, y, radius, 0, 2 * Math.PI, true);
    c.closePath();
    c.fill();

    c.restore();
  }

  exists() {
    return this.alpha >= 0.1 && this.size >= 1;
  }
}

class Rocket extends Particle {
  constructor(x, y) {
    super(x, y);
    this.explosionColor = 0;
  }

  explode() {
    // decide the explosion shape for this rocket
    const fn = [this.sphere, this.star, this.heart][Math.floor(Math.random()*3)];

    // number of particles to be generated
    const count = Math.random() * 10 + 70;

    // create particles
    const particles = [];
    for (let i = 0; i < count; i++) {
      const particle = new Particle(this.x, this.y);
      const {dx, dy} = fn();
      particle.dx = dx;
      particle.dy = dy;
      particle.size = 10;
      particle.gravity = 0.2;
      particle.resistance = 0.92;
      particle.shrink = Math.random() * 0.05 + 0.93;
      particle.flick = true;
      particle.color = this.explosionColor;

      particles.push(particle);
    }

    return particles;
  }

  render(c) {
    if (!this.exists()) {
      return;
    }

    c.save();
    c.globalCompositeOperation = "lighter";
    c.fillStyle = "rgb(255, 200, 0)";
    c.beginPath();

    // draw several particles for each rocket position
    for (let i = 0; i < 5; i++) {
      const angle = Math.random() * 2 * Math.PI;
      const r = Math.random() * this.size / 2;
      const x = this.x + r * Math.cos(angle);
      const y = this.y + r * Math.sin(angle);

      // draw several 1px particles
      c.arc(x, y, 1.2, 0, 2 * Math.PI, true);
    }
    c.closePath();
    c.fill();
    c.restore();
  }

  sphere() {
    // emulate 3D effect by using cosine and put more particles in the middle
    const angle = Math.random() * Math.PI * 2;
    const speed = Math.cos(Math.random() * Math.PI / 2) * 11;
    return {
      dx: Math.cos(angle) * speed,
      dy: Math.sin(angle) * speed
    };
  }

  star() {
    // sin(5*r) creates a star, need to add PI to rotate 180 degrees
    const angle = Math.random() * Math.PI * 2;
    const speed = Math.sin(5 * angle + Math.PI) * 9 + Math.random() * 3;
    return {
      dx: Math.cos(angle) * speed,
      dy: Math.sin(angle) * speed
    };
  }

  heart() {
    // invert y speed to display heart in the right orientation
    const angle = Math.random() * Math.PI * 2;
    const speed = Math.random() * 0.2 + 0.5;
    return {
      dx: (16 * Math.pow(Math.sin(angle), 3)) * speed,
      dy: -(13 * Math.cos(angle) - 5 * Math.cos(2 * angle) - 2 * Math.cos(3 * angle) - Math.cos(4 * angle)) * speed
    };
  }
}

class Smoke extends Particle {
  constructor(x, y) {
    super(x, y);
    this.size = 1;
    this.dx = Math.random() * 0.01;
    this.dy = Math.random() * 0.01;
    this.gravity = -0.2;
    this.resistance = 0.01;
    this.shrink = 1.03;
    this.fade = Math.random() * 0.03 + 0.02;
    this.alpha = 1;
    this.start = 0;
  }

  render(c) {
    if (!this.exists()) {
      return;
    }

    const x = this.x;
    const y = this.y;
    const r = this.size / 2;

    c.save();
    c.globalCompositionOperation = "lighter";

    const gradient = c.createRadialGradient(x, y, 0.1, x, y, r);
    gradient.addColorStop(0.1, `rgba(200, 200, 200, ${this.alpha})`);
    gradient.addColorStop(1, `rgba(150, 150, 150, ${this.alpha})`);
    c.fillStyle = gradient;

    c.beginPath();
    c.arc(this.x, this.y, this.size, 0, Math.PI * 2, true);
    c.lineTo(this.x, this.y);
    c.closePath();
    c.fill();
    c.restore();
  }

  exists() {
    return this.alpha >= 0.01;
  }
}
