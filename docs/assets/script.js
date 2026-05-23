(() => {
  // ---- theme ----
  const root = document.documentElement;
  const stored = localStorage.getItem('mcpsync-theme');
  if (stored === 'light' || stored === 'dark') {
    root.setAttribute('data-theme', stored);
  } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    root.setAttribute('data-theme', 'light');
  }

  const toggle = document.getElementById('themeToggle');
  if (toggle) {
    toggle.addEventListener('click', () => {
      const next = root.getAttribute('data-theme') === 'light' ? 'dark' : 'light';
      root.setAttribute('data-theme', next);
      localStorage.setItem('mcpsync-theme', next);
    });
  }

  // ---- copy buttons ----
  document.querySelectorAll('.copy[data-copy]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const text = btn.getAttribute('data-copy') || '';
      try {
        await navigator.clipboard.writeText(text);
      } catch {
        const ta = document.createElement('textarea');
        ta.value = text;
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        ta.remove();
      }
      const original = btn.textContent;
      btn.textContent = 'copied';
      btn.classList.add('copied');
      setTimeout(() => {
        btn.textContent = original;
        btn.classList.remove('copied');
      }, 1400);
    });
  });

  // ---- install tabs ----
  const tabs = document.querySelectorAll('.tab[data-tab]');
  const panels = document.querySelectorAll('.panel[data-panel]');
  tabs.forEach((tab) => {
    tab.addEventListener('click', () => {
      const target = tab.getAttribute('data-tab');
      tabs.forEach((t) => {
        const active = t === tab;
        t.classList.toggle('is-active', active);
        t.setAttribute('aria-selected', active ? 'true' : 'false');
      });
      panels.forEach((p) => {
        p.classList.toggle('is-active', p.getAttribute('data-panel') === target);
      });
    });
  });

  // ---- commands slider + typing animation ----
  const slider = document.getElementById('cmdSlider');
  if (slider) {
    const slides = slider.querySelectorAll('.slide');
    const tabBtns = slider.querySelectorAll('.slider-tab');
    const dotBtns = slider.querySelectorAll('.dot');
    const arrows = slider.querySelectorAll('.slider-arrow');

    const finalCmd = new WeakMap();
    slides.forEach((slide) => {
      slide.querySelectorAll('.line[data-type]').forEach((line) => {
        const cmdEl = line.querySelector('.cmd');
        finalCmd.set(line, line.getAttribute('data-type') || '');
        if (cmdEl) cmdEl.textContent = '';
      });
    });

    let activeIdx = 0;
    let timers = [];
    let autoTimer = null;
    const AUTO_MS = 9000;
    const prefersReduced = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    function clearTimers() {
      timers.forEach((t) => clearTimeout(t));
      timers = [];
    }

    function animateSlide(slide) {
      const lines = Array.from(slide.querySelectorAll('.line'));
      if (prefersReduced) {
        lines.forEach((l) => {
          l.classList.remove('hidden');
          const cmdEl = l.querySelector('.cmd');
          if (cmdEl && l.hasAttribute('data-type')) cmdEl.textContent = l.getAttribute('data-type');
        });
        return;
      }
      lines.forEach((l) => {
        l.classList.add('hidden');
        const cmdEl = l.querySelector('.cmd');
        if (cmdEl && l.hasAttribute('data-type')) cmdEl.textContent = '';
      });

      let elapsed = 0;
      const step = (idx) => {
        if (idx >= lines.length) return;
        const line = lines[idx];
        const cmdEl = line.querySelector('.cmd');
        const isType = line.hasAttribute('data-type') && cmdEl;

        line.classList.remove('hidden');

        if (isType) {
          const text = line.getAttribute('data-type') || '';
          cmdEl.classList.add('typing');
          let i = 0;
          const typeChar = () => {
            if (!slide.classList.contains('is-active')) { cmdEl.classList.remove('typing'); return; }
            if (i <= text.length) {
              cmdEl.textContent = text.slice(0, i);
              i++;
              timers.push(setTimeout(typeChar, 28 + Math.random() * 22));
            } else {
              cmdEl.classList.remove('typing');
              timers.push(setTimeout(() => step(idx + 1), 280));
            }
          };
          typeChar();
        } else {
          timers.push(setTimeout(() => step(idx + 1), 110));
        }
      };

      step(0);
    }

    function go(idx, opts) {
      opts = opts || {};
      const n = slides.length;
      activeIdx = ((idx % n) + n) % n;
      slides.forEach((s, i) => s.classList.toggle('is-active', i === activeIdx));
      tabBtns.forEach((t, i) => {
        const active = i === activeIdx;
        t.classList.toggle('is-active', active);
        t.setAttribute('aria-selected', active ? 'true' : 'false');
      });
      dotBtns.forEach((d, i) => d.classList.toggle('is-active', i === activeIdx));
      clearTimers();
      animateSlide(slides[activeIdx]);
      if (opts.userAction) resetAuto();
    }

    function resetAuto() {
      if (autoTimer) clearInterval(autoTimer);
      if (prefersReduced) return;
      autoTimer = setInterval(() => go(activeIdx + 1), AUTO_MS);
    }

    tabBtns.forEach((t, i) => t.addEventListener('click', () => go(i, { userAction: true })));
    dotBtns.forEach((d, i) => d.addEventListener('click', () => go(i, { userAction: true })));
    arrows.forEach((a) => a.addEventListener('click', () => {
      const dir = parseInt(a.getAttribute('data-dir'), 10) || 1;
      go(activeIdx + dir, { userAction: true });
    }));

    slider.addEventListener('mouseenter', () => { if (autoTimer) clearInterval(autoTimer); });
    slider.addEventListener('mouseleave', resetAuto);

    // Kick off when commands section enters the viewport (so the typing isn't wasted off-screen).
    const start = () => { go(0); };
    if ('IntersectionObserver' in window) {
      const io = new IntersectionObserver((entries) => {
        entries.forEach((e) => {
          if (e.isIntersecting) { start(); io.disconnect(); }
        });
      }, { threshold: 0.25 });
      io.observe(slider);
    } else {
      start();
    }
  }
})();
