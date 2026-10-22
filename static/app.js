// ---------- Mission Control interactivity ----------
// This file only runs meaningful logic on /mission - on other pages the
// guarded lookups below simply find nothing and the script exits quietly.

(function typedIntro() {
  const el = document.getElementById('mc-typed-text');
  if (!el) return;
  const full = el.getAttribute('data-full-text') || '';
  el.textContent = '';
  const cursor = document.createElement('span');
  cursor.className = 'mc-cursor';
  cursor.innerHTML = '&nbsp;';

  let i = 0;
  const speed = 14; // ms per character - fast enough not to feel slow, slow enough to read as "typed"
  function tick() {
    if (i <= full.length) {
      el.textContent = full.slice(0, i);
      el.appendChild(cursor);
      i++;
      requestAnimationFrame(() => setTimeout(tick, speed));
    } else {
      cursor.remove();
    }
  }
  tick();
})();

(function missionControl() {
  const cards = document.querySelectorAll('.mc-card');
  if (!cards.length) return;

  const captured = window.MC_CAPTURED || {};
  try {
    if (window.parent && window.parent !== window) {
      const solvedCount = Object.keys(captured).filter(function(k) { return captured[k]; }).length;
      window.parent.postMessage({
        type: 'CTF_SCORE_UPDATE',
        captured: solvedCount,
        total: cards.length,
        scorePercentage: Math.round((solvedCount / cards.length) * 100)
      }, '*');
    }
  } catch (e) {}

  cards.forEach((card) => {
    const id = card.getAttribute('data-challenge-id');
    const head = card.querySelector('.mc-card-head');
    const body = document.getElementById('body-' + id);
    const status = document.getElementById('status-' + id);
    const hintBtn = card.querySelector('[data-hint-btn="' + id + '"]');
    const hintsWrap = document.getElementById('hints-' + id);
    const submitBtn = card.querySelector('[data-submit-btn="' + id + '"]');
    const flagInput = document.getElementById('flaginput-' + id);
    const feedback = document.getElementById('feedback-' + id);
    const reveal = document.getElementById('reveal-' + id);

    if (captured[id]) markSolved(card, status);

    head.addEventListener('click', () => {
      const open = body.classList.toggle('is-open');
      if (open) {
        document.querySelectorAll('.mc-card-body.is-open').forEach((b) => {
          if (b !== body) b.classList.remove('is-open');
        });
      }
    });

    hintBtn.addEventListener('click', async () => {
      hintBtn.disabled = true;
      hintBtn.textContent = 'Thinking...';
      try {
        const res = await fetch('/api/hint', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id }),
        });
        const data = await res.json();
        hintsWrap.innerHTML = '';
        (data.hints || []).forEach((h, idx) => {
          const div = document.createElement('div');
          div.className = 'mc-hint';
          div.textContent = 'Hint ' + (idx + 1) + ': ' + h;
          hintsWrap.appendChild(div);
        });
        if (data.done) {
          hintBtn.textContent = 'No more hints for this one';
        } else {
          hintBtn.disabled = false;
          hintBtn.textContent = 'Ask the Security Consultant for another hint';
        }
      } catch (e) {
        hintBtn.disabled = false;
        hintBtn.textContent = 'Ask the Security Consultant for a hint';
      }
    });

    async function doSubmit() {
      const flag = flagInput.value.trim();
      if (!flag) return;
      submitBtn.disabled = true;
      try {
        const res = await fetch('/api/submit-flag', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id, flag }),
        });
        const data = await res.json();
        feedback.textContent = data.message || '';
        feedback.className = 'mc-submit-feedback ' + (data.correct ? 'is-correct' : 'is-wrong');

        if (data.correct) {
          if (data.reveal) reveal.textContent = data.reveal;
          markSolved(card, status);
          updateProgress(data.captured, data.total);
          burstConfetti();
          try {
            if (window.parent && window.parent !== window) {
              window.parent.postMessage({
                type: 'CTF_SCORE_UPDATE',
                challengeId: id,
                captured: data.captured,
                total: data.total,
                scorePercentage: Math.round((data.captured / data.total) * 100),
                allDone: data.all_done
              }, '*');
            }
          } catch (e) {}
          if (data.all_done) {
            setTimeout(showCelebration, 700);
          }
        }
      } catch (e) {
        feedback.textContent = 'Something went wrong reaching Mission Control - try again.';
        feedback.className = 'mc-submit-feedback is-wrong';
      } finally {
        submitBtn.disabled = false;
      }
    }

    submitBtn.addEventListener('click', doSubmit);
    flagInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') doSubmit();
    });
  });

  function markSolved(card, status) {
    card.classList.add('is-solved');
    status.textContent = 'Solved';
  }

  function updateProgress(capturedCount, total) {
    const fill = document.getElementById('mc-progress-fill');
    const label = document.getElementById('mc-progress-count');
    if (fill) fill.style.width = Math.round((capturedCount / total) * 100) + '%';
    if (label) label.textContent = capturedCount + ' / ' + total;
  }

  function showCelebration() {
    const overlay = document.getElementById('mc-celebrate-overlay');
    if (!overlay) return;
    overlay.classList.add('is-visible');
    burstConfetti(140);
  }
  const closeBtn = document.getElementById('mc-celebrate-close');
  if (closeBtn) {
    closeBtn.addEventListener('click', () => {
      document.getElementById('mc-celebrate-overlay').classList.remove('is-visible');
    });
  }

  // ---------- tiny, self-contained confetti burst (no external deps) ----------
  const canvas = document.getElementById('mc-confetti-canvas');
  let confettiParticles = [];
  let confettiRunning = false;

  function resizeCanvas() {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  window.addEventListener('resize', resizeCanvas);
  resizeCanvas();

  function burstConfetti(count) {
    if (!canvas) return;
    const colors = ['#E8A33D', '#4FA88A', '#E7E2D3', '#D9695F'];
    const n = count || 60;
    for (let i = 0; i < n; i++) {
      confettiParticles.push({
        x: Math.random() * canvas.width,
        y: -20 - Math.random() * 100,
        w: 5 + Math.random() * 4,
        h: 8 + Math.random() * 6,
        color: colors[Math.floor(Math.random() * colors.length)],
        vy: 2 + Math.random() * 3,
        vx: -1.5 + Math.random() * 3,
        rot: Math.random() * 360,
        vr: -6 + Math.random() * 12,
        life: 0,
      });
    }
    if (!confettiRunning) {
      confettiRunning = true;
      requestAnimationFrame(animateConfetti);
    }
  }

  function animateConfetti() {
    const ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    confettiParticles.forEach((p) => {
      p.x += p.vx;
      p.y += p.vy;
      p.rot += p.vr;
      p.life++;
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.rotate((p.rot * Math.PI) / 180);
      ctx.fillStyle = p.color;
      ctx.fillRect(-p.w / 2, -p.h / 2, p.w, p.h);
      ctx.restore();
    });
    confettiParticles = confettiParticles.filter((p) => p.y < canvas.height + 40 && p.life < 480);
    if (confettiParticles.length > 0) {
      requestAnimationFrame(animateConfetti);
    } else {
      confettiRunning = false;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
    }
  }
})();
