/**
 * @type {HTMLCanvasElement}
 */
// @ts-ignore
const canvas = document.getElementById('starfield');
/**
 * @type {CanvasRenderingContext2D}
 */
const ctx = canvas.getContext('2d');

let stars = [];
const numStars = 200;
const shootingStarFrequency = 100; // Lower number means more frequent

function resizeCanvas() {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
}

function createStars() {
  stars = [];
  for (let i = 0; i < numStars; i++) {
    stars.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      radius: Math.random() * 1.5,
      alpha: Math.random(),
      speed: Math.random() * 0.5,
    });
  }
}

function drawStars() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  for (let star of stars) {
    ctx.fillStyle = `rgba(255, 255, 255, ${star.alpha})`;
    ctx.beginPath();
    ctx.arc(star.x, star.y, star.radius, 0, Math.PI * 2);
    ctx.fill();

    star.y -= star.speed;
    if (star.y < 0) {
      star.y = canvas.height;
    }
  }

  if (Math.random() * shootingStarFrequency < 1) {
    drawShootingStar();
  }

  requestAnimationFrame(drawStars);
}

function drawShootingStar() {
  const xStart = Math.random() * canvas.width;
  const yStart = Math.random() * canvas.height / 2;
  const length = Math.random() * 400 + 200;
  const speed = Math.random() * 10 + 6;
  const angle = Math.PI / 4;

  const xEnd = xStart + length * Math.cos(angle);
  const yEnd = yStart + length * Math.sin(angle);

  ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(xStart, yStart);
  ctx.lineTo(xEnd, yEnd);
  ctx.stroke();
  ctx.lineWidth = 1;
}

resizeCanvas();
createStars();
drawStars();
window.addEventListener('resize', () => {
  resizeCanvas();
  createStars();
});
