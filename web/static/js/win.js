//
// This code is completely stolen from: https://jsfiddle.net/dtrooper/AceJJ/
//

var WIN_ANIMATION_RUNNING = false,
  MAX_PARTICLES = 400;

var canvas = document.createElement("canvas"),
  context = canvas.getContext("2d"),
  particles = [],
  rockets = [],
  launchInterval = null,
  loopInterval = null;

function start_win_animation() {
  WIN_ANIMATION_RUNNING = true;
  document.body.appendChild(canvas);

  launchInterval = setInterval(launch, 800);
  loopInterval = setInterval(loop, 1000 / 50);
  setTimeout(stop_win_animation, 20000);
}

function stop_win_animation() {
  WIN_ANIMATION_RUNNING = false;
  document.body.removeChild(canvas);
}

function launch() {
  if (WIN_ANIMATION_RUNNING === false) {
    clearInterval(launchInterval);
    launchInterval = null;
    return;
  }

  if (rockets.length < 10) {
    var x = Math.floor(Math.random() * window.innerWidth / 10) * 10,
      rocket = new Rocket(x);
    rocket.explosionColor = Math.floor(Math.random() * 360 / 10) * 10;
    rocket.vel.y = Math.random() * -3 - 4;
    rocket.vel.x = Math.random() * 6 - 3;
    rocket.size = 8;
    rocket.shrink = 0.999;
    rocket.gravity = 0.01;
    rockets.push(rocket);
  }
}

function loop() {
  if (WIN_ANIMATION_RUNNING === false) {
    clearInterval(loopInterval);
    loopInterval = null;
    return;
  }

  // update screen size
  var width = window.innerWidth,
    height = window.innerHeight;
  if (canvas.width !== width) {
    canvas.width = width
  }
  if (canvas.height !== height) {
    canvas.height = height
  }

  // clear canvas
  context.fillStyle = "rgba(0, 0, 0, 0.05)";
  context.fillRect(0, 0, width, height);

  // render rockets first...
  var existingRockets = [];
  for (var i = 0; i < rockets.length; i++) {
    // update and render
    rockets[i].update();
    rockets[i].render(context);

    // determine if the rocket is going to explode
    var randomChance = rockets[i].pos.y < (height * 2 / 3)
      ? (Math.random() * 100 <= 1)
      : false;
    if (rockets[i].pos.y < height / 5 || rockets[i].vel.y >= 0 || randomChance) {
      rockets[i].explode();
      continue;
    }

    // if it hasn't exploded, then remember it for next time through
    existingRockets.push(rockets[i]);
  }
  rockets = existingRockets;

  // now render particles...
  var existingParticles = [];
  for (var i = 0; i < particles.length; i++) {
    particles[i].update();

    if (!particles[i].exists()) {
      continue;
    }

    // render and save particles for next time through
    particles[i].render(context);
    existingParticles.push(particles[i]);
  }
  particles = existingParticles;

  // if there are too many particles then lose some.
  while (particles.length > MAX_PARTICLES) {
    particles.shift();
  }
}

function Particle(pos) {
  this.pos = {
    x: pos ? pos.x : 0,
    y: pos ? pos.y : 0
  };
  this.vel = {
    x: 0,
    y: 0
  };
  this.shrink = .97;
  this.size = 2;

  this.resistance = 1;
  this.gravity = 0;

  this.flick = false;

  this.alpha = 1;
  this.fade = 0;
  this.color = 0;
}

Particle.prototype.update = function () {
  // apply resistance
  this.vel.x *= this.resistance;
  this.vel.y *= this.resistance;

  // gravity down
  this.vel.y += this.gravity;

  // update position based on speed
  this.pos.x += this.vel.x;
  this.pos.y += this.vel.y;

  // shrink
  this.size *= this.shrink;

  // fade out
  this.alpha -= this.fade;
};

Particle.prototype.render = function (c) {
  if (!this.exists()) {
    return;
  }

  c.save();

  c.globalCompositeOperation = 'lighter';

  var x = this.pos.x,
    y = this.pos.y,
    r = this.size / 2;

  var gradient = c.createRadialGradient(x, y, 0.1, x, y, r);
  gradient.addColorStop(0.1, "rgba(255,255,255," + this.alpha + ")");
  gradient.addColorStop(0.8, "hsla(" + this.color + ", 100%, 50%, " + this.alpha + ")");
  gradient.addColorStop(1, "hsla(" + this.color + ", 100%, 50%, 0.1)");

  c.fillStyle = gradient;

  c.beginPath();
  c.arc(this.pos.x, this.pos.y, this.flick ? Math.random() * this.size : this.size, 0, Math.PI * 2, true);
  c.closePath();
  c.fill();

  c.restore();
};

Particle.prototype.exists = function () {
  return this.alpha >= 0.1 && this.size >= 1;
};

function Rocket(x) {
  Particle.apply(this, [{
    x: x,
    y: window.innerHeight,
  }]);

  this.explosionColor = 0;
}

Rocket.prototype = new Particle();
Rocket.prototype.constructor = Rocket;

Rocket.prototype.explode = function () {
  var count = Math.random() * 10 + 80;

  for (var i = 0; i < count; i++) {
    var particle = new Particle(this.pos);
    var angle = Math.random() * Math.PI * 2;

    // emulate 3D effect by using cosine and put more particles in the middle
    var speed = Math.cos(Math.random() * Math.PI / 2) * 15;

    particle.vel.x = Math.cos(angle) * speed;
    particle.vel.y = Math.sin(angle) * speed;

    particle.size = 10;

    particle.gravity = 0.2;
    particle.resistance = 0.92;
    particle.shrink = Math.random() * 0.05 + 0.93;

    particle.flick = true;
    particle.color = this.explosionColor;

    particles.push(particle);
  }
};

Rocket.prototype.render = function (c) {
  if (!this.exists()) {
    return;
  }

  c.save();

  c.globalCompositeOperation = 'lighter';

  var x = this.pos.x,
    y = this.pos.y,
    r = this.size / 2;

  var gradient = c.createRadialGradient(x, y, 0.1, x, y, r);
  gradient.addColorStop(0.1, "rgba(255, 255, 255 ," + this.alpha + ")");
  gradient.addColorStop(1, "rgba(0, 0, 0, " + this.alpha + ")");

  c.fillStyle = gradient;

  c.beginPath();
  c.arc(this.pos.x, this.pos.y, this.flick ? Math.random() * this.size / 2 + this.size / 2 : this.size, 0, Math.PI * 2, true);
  c.closePath();
  c.fill();

  c.restore();
};