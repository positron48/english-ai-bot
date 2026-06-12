// Linglow — City Map Screen (Ciudad Luminaria — canvas edition)
const { useState, useEffect, useRef } = React;

// ─── Ciudad Luminaria district definitions ────────────────────────────────────
// lv: current mastery level (1=locked … 5=master)
// lp: [x%, y%] label anchor on canvas
// poly: [[x%,y%]…] mask polygon vertices
const CITY_DISTS = [
{ id: 'alto', name: 'Distrito Alto', lv: 1, lp: [23.9, 13.7], wrap: false,
  poly: [[0, 0], [57, 0], [57.29, 13.59], [36.17, 19.67], [34.58, 23], [21.78, 23.72], [0.37, 23.96]],
  acts: [{ i: '📖', s: 'started' }, { i: '🌿', s: 'started' }, { i: '📚', s: 'started' }, { i: '💬', s: 'started' }] },
{ id: 'puentes', name: 'Puentes del Relato', lv: 1, lp: [80.3, 21.8], wrap: true,
  poly: [[57, 0], [99.72, 0], [99.91, 40.88], [62.34, 26.94], [53.08, 28.13], [34.58, 23], [36.17, 19.67], [57.29, 13.59]],
  acts: [{ i: '📖', s: 'started' }, { i: '🌿', s: 'started' }, { i: '📚', s: 'started' }, { i: '💬', s: 'started' }] },
{ id: 'barrio', name: 'Barrio Vivo', lv: 3, lp: [18, 34.4], wrap: false,
  poly: [[0.37, 23.96], [21.78, 23.72], [34.58, 23], [53.08, 28.13], [52.99, 31.7], [50.65, 34.21], [36.07, 39.69], [23.55, 45.17], [20.47, 54.47], [21.59, 57.09], [0, 57]],
  acts: [{ i: '📖', s: 'stable' }, { i: '🌿', s: 'growing' }, { i: '📚', s: 'weak' }, { i: '💬', s: 'started' }] },
{ id: 'plaza', name: 'Plaza Clara', lv: 5, lp: [65.3, 51.8], wrap: false,
  poly: [[99.91, 40.88], [99.91, 58.05], [60, 73.3], [52.9, 69.13], [21.59, 57.09], [20.47, 54.47], [23.55, 45.17], [36.07, 39.69], [50.65, 34.21], [52.99, 31.7], [53.08, 28.13], [62.34, 26.94]],
  acts: [{ i: '📖', s: 'stable' }, { i: '🌿', s: 'stable' }, { i: '📚', s: 'growing' }, { i: '💬', s: 'growing' }] },
{ id: 'puerta', name: 'Puerta de la Chispa', lv: 2, lp: [22.8, 73.5], wrap: true,
  poly: [[0, 57], [21.59, 57.09], [52.9, 69.13], [60, 73.3], [60.09, 78.19], [52.15, 83.08], [41.4, 90.11], [45.79, 99.88], [0.37, 99.4], [0, 86]],
  acts: [{ i: '📖', s: 'started' }, { i: '🌿', s: 'growing' }, { i: '📚', s: 'started' }, { i: '💬', s: 'started' }] },
{ id: 'campus', name: 'Campus de Maestría', lv: 1, lp: [76.5, 91.1], wrap: true,
  poly: [[99.91, 58.05], [100, 100], [45.79, 99.88], [41.4, 90.11], [52.15, 83.08], [60.09, 78.19], [60, 73.3]],
  acts: [{ i: '📖', s: 'started' }, { i: '🌿', s: 'started' }, { i: '📚', s: 'started' }, { i: '💬', s: 'started' }] }];


// ─── Canvas rendering (pure functions, no React deps) ─────────────────────────
function _oc(w, h) {
  const el = document.createElement('canvas');
  el.width = w;el.height = h;
  return { el, c: el.getContext('2d') };
}
function _poly(ctx, pts, W, H) {
  ctx.beginPath();
  pts.forEach(([x, y], i) => i ? ctx.lineTo(x * W / 100, y * H / 100) : ctx.moveTo(x * W / 100, y * H / 100));
  ctx.closePath();ctx.fill();
}
function _renderCity(cvs, imgs, dists) {
  if (!cvs || !imgs[0]) return;
  const tw = cvs.width,th = cvs.height;
  const ctx = cvs.getContext('2d');
  ctx.clearRect(0, 0, tw, th);
  ctx.drawImage(imgs[0], 0, 0, tw, th);
  for (let t = 2; t <= 5; t++) {
    const elig = dists.filter((d) => d.lv >= t);
    if (!elig.length || !imgs[t - 1]) continue;
    const lc = _oc(tw, th);
    lc.c.drawImage(imgs[t - 1], 0, 0, tw, th);
    const mc = _oc(tw, th);
    mc.c.filter = `blur(${Math.max(10, Math.round(tw * 0.032))}px)`;
    mc.c.fillStyle = '#fff';
    elig.forEach((d) => _poly(mc.c, d.poly, tw, th));
    mc.c.filter = 'none';
    lc.c.globalCompositeOperation = 'destination-in';
    lc.c.drawImage(mc.el, 0, 0);
    lc.c.globalCompositeOperation = 'source-over';
    ctx.drawImage(lc.el, 0, 0);
  }
}

// ─── Activity icon colours ─────────────────────────────────────────────────────
const ACT_BG = {
  stable: 'rgba(36,108,56,0.9)',
  growing: 'rgba(70,155,85,0.85)',
  weak: 'rgba(196,144,28,0.9)',
  started: 'rgba(142,132,120,0.82)'
};

// ─── Map Screen ───────────────────────────────────────────────────────────────
function MapScreen({ go, theme }) {
  const { user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const cvsRef = useRef(null);
  const wrapRef = useRef(null);
  const imgsRef = useRef(new Array(5).fill(null));
  const readyRef = useRef(0);

  const resize = () => {
    const c = cvsRef.current;
    if (!c) return;
    const r = c.getBoundingClientRect();
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    c.width = r.width * dpr;
    c.height = r.height * dpr;
    _renderCity(c, imgsRef.current, CITY_DISTS);
  };

  useEffect(() => {
    // Small delay so layout settles before measuring
    const t = setTimeout(resize, 50);

    ['uploads/1.jpg', 'uploads/2.jpg', 'uploads/3.jpg', 'uploads/4.jpg', 'uploads/5.jpg'].
    forEach((src, i) => {
      const img = new Image();
      img.onload = () => {
        imgsRef.current[i] = img;
        if (++readyRef.current === 5)
        _renderCity(cvsRef.current, imgsRef.current, CITY_DISTS);
      };
      img.onerror = () => {if (++readyRef.current === 5) _renderCity(cvsRef.current, imgsRef.current, CITY_DISTS);};
      img.src = src;
    });

    window.addEventListener('resize', resize);
    return () => {clearTimeout(t);window.removeEventListener('resize', resize);};
  }, []);

  const openDistrict = (d) => {
    if (d.lv > 1) go(`district:${d.id}`);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', background: 'var(--bg)' }}>

      {/* ── Map area (full height, title floats over image) ── */}
      <div ref={wrapRef} style={{
        flex: 1, position: 'relative', overflow: 'hidden',
        display: 'flex', flexDirection: 'column'
      }}>
        {/* Title overlay — floats on top of the map */}
        <div style={{
          position: 'absolute', top: 0, left: 0, right: 0,
          zIndex: 20,
          padding: '18px 20px 32px',
          background: 'linear-gradient(to bottom, rgba(0,0,0,0.28) 0%, transparent 100%)',
          pointerEvents: 'none'
        }}>
          <div style={{
            fontFamily: 'Lora,serif', fontWeight: 700, fontSize: 22,
            color: '#fff',
            textShadow: '0 1px 8px rgba(0,0,0,0.5)',
            lineHeight: 1, textAlign: 'center'
          }}>
            Ciudad Luminaria
          </div>
          <div style={{
            fontFamily: 'Inter,sans-serif', fontSize: 11,
            color: 'rgba(255,255,255,0.78)',
            textShadow: '0 1px 4px rgba(0,0,0,0.4)',
            marginTop: 4, textAlign: 'center', letterSpacing: 0.1
          }}>
            Исследуй город и учи испанский каждый день.
          </div>
        </div>
        <canvas ref={cvsRef} style={{ display: 'block', width: '100%', flex: 1 }} />

        {/* Nav spacer — keeps canvas from hiding behind bottom nav */}
        <div style={{ height: isDesktop ? 0 : 72, flexShrink: 0 }} />

        {/* District label overlays — positioned over canvas only */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: isDesktop ? 0 : 72, pointerEvents: 'none' }}>
          {CITY_DISTS.map((d) => {
            const locked = d.lv <= 1;
            return (
              <div key={d.id}
              onClick={() => openDistrict(d)}
              style={{
                position: 'absolute',
                left: d.lp[0] + '%', top: d.lp[1] + '%',
                transform: 'translate(-50%,-50%)',
                textAlign: 'center',
                maxWidth: 115,
                pointerEvents: 'auto',
                cursor: locked ? 'default' : 'pointer'
              }}>

                {/* Name */}
                <span style={{
                    display: 'block',
                    fontFamily: 'Lora, Georgia, serif',
                    fontSize: 13, fontWeight: 700,
                    lineHeight: 1.25,
                    whiteSpace: d.wrap ? 'normal' : 'nowrap',
                    color: locked ? 'rgba(15,13,10,0.68)' : '#1e1208',
                    textShadow: locked
                      ? '0 0 8px rgba(255,255,255,0.9), 1px 1px 2px rgba(0,0,0,0.45), -1px -1px 2px rgba(0,0,0,0.35)'
                      : '0 0 18px rgba(255,248,210,0.99), 0 0 8px rgba(255,236,160,0.88), 0 1px 3px rgba(255,255,255,0.7)',
                  }}>
                  {d.name}
                </span>

                {/* Lock or activity icons */}
                {locked ?
                <div style={{ display: 'flex', justifyContent: 'center', marginTop: 4 }}>
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                  stroke="rgba(44,31,14,0.38)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="3" y="11" width="18" height="11" rx="2" />
                      <path d="M7 11V7a5 5 0 0110 0v4" />
                    </svg>
                  </div> :

                <div style={{ display: 'flex', gap: 3, marginTop: 5, justifyContent: 'center' }}>
                    {d.acts.map((a, ai) =>
                  <div key={ai} style={{
                    width: 24, height: 24, borderRadius: 7,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 11,
                    background: ACT_BG[a.s] || ACT_BG.started,
                    border: '1px solid rgba(255,255,255,0.28)',
                    boxShadow: '0 1px 4px rgba(0,0,0,0.2)'
                  }}>{a.i}</div>
                  )}
                  </div>
                }
              </div>);

          })}
        </div>
      </div>
    </div>);

}

// ─── District Detail Screen ───────────────────────────────────────────────────
function DistrictScreen({ districtId, go, theme }) {
  const { districtDetail, user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const d = districtDetail[districtId] || districtDetail.plaza || Object.values(districtDetail)[0];
  if (!d) return null;

  const areas = [
    { icon: '🏛', label: 'Грамматика', level: 2, done: 4, total: 10, pct: 75, color: '#2d6b3a',  cta: 'Продолжить', action: () => go('grammar') },
    { icon: '🌿', label: 'Слова',      level: 3, done: 26, total: 50, pct: 60, color: '#2d6b3a', cta: 'Продолжить', action: () => go('exercise:0') },
    { icon: '📚', label: 'Чтение',     level: 2, done: 3,  total: 8,  pct: 40, color: '#c8a84b', cta: 'Читать',     action: () => go('reading') },
    { icon: '💬', label: 'Общение',    level: 2, done: 2,  total: 8,  pct: 20, color: '#2d6b3a', cta: 'Практиковать', action: () => go('chat') },
  ];

  const buildings = [
    { name: 'Jardín de Frases',       sub: 'грамматика', x: '18%', y: '22%' },
    { name: 'Mercado de Palabras',    sub: 'слова',      x: '72%', y: '18%' },
    { name: 'Quiosco de Lectura',     sub: 'чтение',     x: '12%', y: '55%' },
    { name: 'Cabinas de Conversación',sub: 'общение',    x: '76%', y: '52%' },
    { name: 'Estación de Repaso',     sub: 'повторение', x: '32%', y: '78%' },
    { name: 'Taller de Errores',      sub: 'ошибки',     x: '68%', y: '78%' },
  ];

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? 32 : 80 }}>

      {/* ── Centered header ── */}
      <CenteredHeader title={d.name} onBack={() => go('map')} streakN={user.streak} />

      {/* ── Subtitle + status chip ── */}
      <div style={{ padding: '0 16px 14px', textAlign: 'center' }}>
        <p style={{ margin: '0 0 10px', fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', lineHeight: 1.5 }}>
          {d.desc}
        </p>
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: '5px',
          padding: '5px 14px', borderRadius: '20px',
          background: 'rgba(45,107,58,0.1)', border: '1px solid rgba(45,107,58,0.25)',
          fontFamily: 'Inter,sans-serif', fontSize: '12px', fontWeight: 600, color: '#2d6b3a',
        }}>🌿 растёт уверенность</span>
      </div>

      {/* ── District illustration with building labels ── */}
      <div style={{
        margin: '0 16px', borderRadius: '18px', overflow: 'hidden',
        position: 'relative', height: 240, background: 'var(--chip-bg)',
        border: '1px solid var(--border)',
      }}>
        <img src="uploads/img2.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
          onError={e => { e.target.style.display = 'none'; }} />
        {/* Building labels */}
        {buildings.map((b, i) => (
          <div key={i} style={{
            position: 'absolute', left: b.x, top: b.y,
            transform: 'translate(-50%, -50%)',
            background: 'rgba(255,251,240,0.92)',
            border: '1px solid rgba(200,168,75,0.4)',
            borderRadius: '10px', padding: '5px 9px',
            boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
            maxWidth: 110,
          }}>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '10px', color: '#2d3a1e', lineHeight: 1.2 }}>{b.name}</div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '9px', color: '#7a6a50', marginTop: '1px' }}>{b.sub}</div>
          </div>
        ))}
      </div>

      {/* ── 4 area rows ── */}
      <div style={{ margin: '14px 16px 0' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', overflow: 'hidden' }}>
          {areas.map((a, i) => (
            <div key={i} style={{ borderTop: i > 0 ? '1px solid var(--border)' : 'none' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '12px 16px' }}>
                <div style={{
                  width: 36, height: 36, borderRadius: '10px', flexShrink: 0,
                  background: 'var(--chip-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '18px',
                }}>{a.icon}</div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)', marginBottom: '1px' }}>{a.label}</div>
                  <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>Уровень {a.level} · {a.done} из {a.total}</div>
                  <div style={{ marginTop: '5px' }}><ProgressBar pct={a.pct} h={3} color={a.color} /></div>
                </div>
                <button onClick={a.action} style={{
                  padding: '7px 14px', borderRadius: '20px', border: 'none',
                  background: 'transparent', cursor: 'pointer', flexShrink: 0,
                  fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '13px',
                  color: 'var(--salvia)',
                  display: 'flex', alignItems: 'center', gap: '2px',
                }}>
                  {a.cta} <ChevronRight s={13} c="var(--salvia)" />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ── Новые открытия ── */}
      <div style={{ margin: '16px 16px 0' }}>
        <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '18px', color: 'var(--text)', marginBottom: '10px' }}>Новые открытия</div>
        <button onClick={() => go('reading')} style={{
          width: '100%', background: 'var(--card-bg)', border: '1px solid var(--border)',
          borderRadius: '16px', padding: '14px 16px', cursor: 'pointer', textAlign: 'left',
          display: 'flex', alignItems: 'center', gap: '12px',
        }}>
          <div style={{ width: 44, height: 44, borderRadius: '12px', background: 'var(--chip-bg)', overflow: 'hidden', flexShrink: 0 }}>
            <img src="uploads/img4.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              onError={e => { e.target.style.display = 'none'; }} />
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)', marginBottom: '2px' }}>Фраза дня разблокирована</div>
            <div style={{ fontFamily: 'Lora,serif', fontWeight: 600, fontSize: '15px', color: 'var(--text)' }}>¿Cómo llego al mercado?</div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', marginTop: '1px' }}>Узнай, как уверенно спросить дорогу.</div>
          </div>
          <ChevronRight s={16} c="var(--subtext)" />
        </button>
      </div>

      {/* ── Lumi tip ── */}
      <div style={{ margin: '12px 16px 0' }}>
        <div style={{ background: 'var(--chip-bg)', border: '1px solid var(--border)', borderRadius: '16px', padding: '12px 14px', display: 'flex', gap: '10px', alignItems: 'flex-start' }}>
          <LumiSVG size={44} />
          <div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '12px', color: 'var(--text)', marginBottom: '3px' }}>Lumi советует ✦</div>
            <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', lineHeight: 1.5 }}>{d.lumiTip || '¡Sigue practicando cada día!'}</p>
          </div>
        </div>
      </div>

    </div>
  );
}

Object.assign(window, { MapScreen, DistrictScreen });