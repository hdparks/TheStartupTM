const logo = document.getElementById('logo');
const clickCounter = document.getElementById('click-counter');
const footerLogo = document.getElementById('footer-logo');
const contentDiv = document.getElementById('content');
let clickCount = Number(localStorage.getItem('clickCount')) || 0;

function updateClickCounter() {
  clickCounter.textContent = `Clicks: ${clickCount}`;
}

function handleLogoClick(event) {
  clickCount++;
  localStorage.setItem('clickCount', clickCount.toString());
  updateClickCounter();

  const clickEffect = document.createElement('div');
  clickEffect.classList.add('click-effect');
  clickEffect.style.left = `${event.clientX - 50}px`;
  clickEffect.style.top = `${event.clientY - 50}px`;
  document.body.appendChild(clickEffect);

  setTimeout(() => {
    clickEffect.remove();
  }, 500);

  if (clickCount >= 10) {
    document.getElementById('hero').style.display = 'none';
    document.getElementById('content').style.display = 'block';
    loadContent();
  }
}
let loaded = false
function loadContent() {
  if (loaded) { return }
  contentDiv.innerHTML = ''
  fetch('data/content.json')
    .then(response => response.json())
    .then(data => {
      data.layers.forEach(layer => {
        const layerDiv = document.createElement('div');
        layerDiv.classList.add('layer', layer.className);
        layerDiv.innerHTML = `<div class="content-box">${layer.content}</div>`;
        contentDiv.appendChild(layerDiv);
      });
      loaded = true
    })
    .catch(error => console.error('Error loading content:', error));
}

logo.addEventListener('click', handleLogoClick);
footerLogo.addEventListener('click', handleLogoClick);
updateClickCounter();

// if (clickCount >= 10) {
//   document.getElementById('hero').style.display = 'none';
//   document.getElementById('content').style.display = 'block';
//   loadContent();
// }
