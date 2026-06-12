// Linglow — Home, Progress, Profile Screens
const { useState, useEffect } = React;

// ─── Home Screen ──────────────────────────────────────────────────────────────
function HomeScreen({ go }) {
  const { user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const { theme } = React.useContext(ThemeCtx);
  const isDark = theme === 'dark';



  const [langOpen, setLangOpen] = useState(false);
  const [lang, setLang]         = useState('Испанский');
  const langRef = React.useRef(null);

  useEffect(() => {
    if (!langOpen) return;
    const close = (e) => { if (langRef.current && !langRef.current.contains(e.target)) setLangOpen(false); };
    document.addEventListener('mousedown', close);
    return () => document.removeEventListener('mousedown', close);
  }, [langOpen]);

    // ── CSS-var text aliases (React resolves these at runtime) ──
  const T  = 'var(--text)';
  const S  = 'var(--subtext)';
  const BD = 'var(--border)';
  const CB = 'var(--card-bg)';

  // ── Theme-specific values that can't be a simple CSS var ──
  const SHADOW = isDark
    ? '0 16px 40px rgba(0,0,0,0.32), inset 0 1px 0 rgba(255,255,255,0.035)'
    : '0 14px 34px rgba(86,57,22,0.09), inset 0 1px 0 rgba(255,255,255,0.75)';

  // City card: text overlay gradient (left → transparent)
  const cityOverlay = isDark
    ? 'linear-gradient(90deg,rgba(16,28,21,0.97) 0%,rgba(16,28,21,0.78) 50%,rgba(16,28,21,0.12) 100%)'
    : 'linear-gradient(90deg,rgba(255,249,237,0.98) 0%,rgba(255,249,237,0.80) 50%,rgba(255,249,237,0.05) 100%)';

  // Illustration placeholders
  const cityPhBg = isDark
    ? 'linear-gradient(150deg,#182818 0%,#0d190f 60%,#162216 100%)'
    : 'linear-gradient(150deg,#EAE2CA 0%,#C8BCA0 60%,#D4C8AA 100%)';

  const districtPhBg = isDark
    ? 'linear-gradient(170deg,#1A261C 0%,#0F1C12 60%,#152015 100%)'
    : 'linear-gradient(170deg,#D4CDB0 0%,#C0B490 60%,#CABEA0 100%)';

  const districtOverlay = isDark
    ? 'linear-gradient(to right,rgba(16,28,21,0.0) 0%,rgba(16,28,21,0.82) 48%,rgba(16,28,21,0.97) 100%)'
    : 'linear-gradient(to right,rgba(255,249,237,0.0) 0%,rgba(255,249,237,0.80) 48%,rgba(255,249,237,0.98) 100%)';

  // Gold accent colour (dorado)
  const GOLD = isDark ? '#D9B25C' : '#D9A83F';

  // ── Inline SVG icons ──────────────────────────────────────────────────────
  const ChevronDown = ({ s = 11, c = 'var(--subtext)' }) => (
    <svg width={s} height={s} viewBox="0 0 12 12" fill="none">
      <path d="M2 4L6 8L10 4" stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
  const ChatDots = ({ s = 20, c = 'var(--hoja)' }) => (
    <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
      <path d="M21 15C21 15.5 20.79 16.04 20.41 16.41C20.04 16.79 19.53 17 19 17H7L3 21V5C3 4.47 3.21 3.96 3.59 3.59C3.96 3.21 4.47 3 5 3H19C19.53 3 20.04 3.21 20.41 3.59C20.79 3.96 21 4.47 21 5V15Z"
        stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <circle cx="8"  cy="10" r="1.2" fill={c} />
      <circle cx="12" cy="10" r="1.2" fill={c} />
      <circle cx="16" cy="10" r="1.2" fill={c} />
    </svg>
  );
  const OpenBook = ({ s = 20, c = 'var(--hoja)' }) => (
    <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
      <path d="M2 3H8C9.06 3 10.08 3.42 10.83 4.17C11.58 4.92 12 5.94 12 7V21C12 20.21 11.68 19.45 11.12 18.88C10.55 18.32 9.79 18 9 18H2V3Z"
        stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M22 3H16C14.94 3 13.92 3.42 13.17 4.17C12.42 4.92 12 5.94 12 7V21C12 20.21 12.32 19.45 12.88 18.88C13.45 18.32 14.21 18 15 18H22V3Z"
        stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );

  // ── Universal icon box ──────────────────────────────────────────────────
  const IconBox = ({ children, circle = false, size = 42 }) => (
    <div style={{
      width: size, height: size, flexShrink: 0,
      borderRadius: circle ? '50%' : 11,
      background: 'var(--icon-bg)',
      border: '1px solid var(--icon-border)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
    }}>{children}</div>
  );

  // ── Task rows ────────────────────────────────────────────────────────────
  const tasks = [
    {
      icon: (
        <div style={{ position: 'relative', width: 42, height: 42, flexShrink: 0 }}>
          <IconBox><OpenBook s={20} /></IconBox>
          <div style={{
            position: 'absolute', top: -5, left: -5, width: 20, height: 20,
            borderRadius: '50%', background: GOLD,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: `0 1px 4px rgba(0,0,0,.22)`,
          }}>
            <span style={{ color: isDark ? '#07110D' : '#20352A', fontSize: 10, fontWeight: 800, fontFamily: 'Inter,sans-serif', lineHeight: 1 }}>12</span>
          </div>
        </div>
      ),
      title: 'Повтори 12 слов', sub: 'Укрепи свою память', fn: () => go('exercise:0'),
    },
    {
      icon: (
        <IconBox circle>
          <MapPinIcon s={18} c="var(--hoja)" />
        </IconBox>
      ),
      title: 'Продолжи Plaza Clara', sub: 'Осваивай район и его истории', fn: () => go('district:plaza'),
    },
    {
      icon: (
        <IconBox><OpenBook s={20} /></IconBox>
      ),
      title: 'Прочитай короткий текст', sub: 'Погрузись в язык', fn: () => go('reading'),
    },
    {
      icon: (
        <IconBox><ChatDots s={20} /></IconBox>
      ),
      title: 'Попрактикуй диалог с ИИ', sub: 'Говори уверенно', fn: () => go('chat'),
    },
  ];

  // ── Area bars ────────────────────────────────────────────────────────────
  const areas = [
    { icon: '🏛', label: 'Грамматика', pct: 75, bar: isDark ? '#A8D07F' : '#3F6F3F' },
    { icon: '🌿', label: 'Слова',      pct: 60, bar: isDark ? '#7FAE6A' : '#7FAE6A' },
    { icon: '📖', label: 'Чтение',     pct: 40, bar: GOLD },
    { icon: '💬', label: 'Общение',    pct: 20, bar: isDark ? '#A8D07F' : '#3F6F3F' },
  ];


  const langs = ['Испанский', 'Английский'];

    return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '90px', background: 'var(--bg)' }}>

      {/* ─── Header ─────────────────────────────────────────────────────── */}
      <div style={{ padding: '22px 20px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div style={{ position: 'relative' }} ref={langRef}>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: '4px' }}>
            <span style={{ fontFamily: 'Lora,serif', fontSize: '36px', color: T, letterSpacing: '-0.02em', lineHeight: 1 }}>Linglow</span>
            <span style={{ color: GOLD, fontSize: '16px', lineHeight: 1, paddingBottom: '2px' }}>✦</span>
          </div>
          <button
            onClick={() => setLangOpen(o => !o)}
            style={{ display: 'flex', alignItems: 'center', gap: '4px', background: 'none', border: 'none', cursor: 'pointer', padding: '3px 0 0' }}
          >
            <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '14px', color: S }}>{lang}</span>
            <ChevronDown s={11} c={S} />
          </button>
          {langOpen && (
            <div style={{
              position: 'absolute', top: '110%', left: 0,
              background: CB, border: `1px solid ${BD}`, borderRadius: '12px',
              overflow: 'hidden', zIndex: 300, minWidth: '150px',
              boxShadow: SHADOW,
            }}>
              {langs.map((l, i) => (
                <button key={l} onClick={() => { setLang(l); setLangOpen(false); }} style={{
                  width: '100%', padding: '11px 16px', border: 'none', background: 'none',
                  cursor: 'pointer', textAlign: 'left', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                  fontFamily: 'Inter,sans-serif', fontSize: '14px',
                  color: l === lang ? T : S,
                  fontWeight: l === lang ? 600 : 400,
                  borderBottom: i < langs.length - 1 ? `1px solid ${BD}` : 'none',
                }}>
                  {l}
                  {l === lang && <span style={{ color: GOLD, fontSize: '12px' }}>✓</span>}
                </button>
              ))}
            </div>
          )}
        </div>
        {/* Streak badge */}
        <StreakBadge n={user.streak} />
      </div>

      {/* ─── Ciudad Luminaria ────────────────────────────────────────────── */}
      <div style={{ margin: '16px 16px 0' }}>
        <button onClick={() => go('map')} style={{
          width: '100%', background: CB, border: `1px solid ${BD}`, borderRadius: '22px',
          overflow: 'hidden', cursor: 'pointer', textAlign: 'left',
          display: 'block', padding: 0, position: 'relative', minHeight: 118,
          boxShadow: SHADOW,
        }}>
          {/* Placeholder city illustration — full card */}
          <div style={{ position: 'absolute', inset: 0, background: cityPhBg }} />
          {/* Overlay gradient — makes left readable */}
          <div style={{ position: 'absolute', inset: 0, background: cityOverlay }} />
          {/* Text content */}
          <div style={{ position: 'relative', zIndex: 2, padding: '18px 18px 18px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%', minHeight: 118 }}>
            <div>
              <div style={{ fontFamily: 'Lora,serif', fontSize: '24px', color: T, lineHeight: 1.15, marginBottom: '5px' }}>Ciudad Luminaria</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                <span style={{ fontSize: '11px', color: S }}>🔒</span>
                <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: S }}>Distrito Alto</span>
              </div>
            </div>
            <div style={{ marginTop: '14px', maxWidth: '55%' }}>
              <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: S, marginBottom: '5px' }}>3 из 6 районов</div>
              <div style={{ height: 4, borderRadius: 3, background: 'var(--progress-track)', overflow: 'hidden' }}>
                <div style={{ height: '100%', width: '50%', background: isDark ? 'linear-gradient(90deg,#76A765,#A8D07F)' : 'linear-gradient(90deg,#3F6F3F,#7FAE6A)', borderRadius: 3, boxShadow: isDark ? '0 0 8px rgba(168,208,127,0.28)' : 'none' }} />
              </div>
            </div>
          </div>
        </button>
      </div>

      {/* ─── Твой путь сегодня ──────────────────────────────────────────── */}
      <div style={{ margin: '10px 16px 0', position: 'relative' }}>
        {/* Lumi floats above card top-right */}
        <div style={{ position: 'absolute', top: -22, right: 8, zIndex: 10, pointerEvents: 'none' }}>
          <LumiSVG size={68} />
        </div>

        <div style={{ background: CB, border: `1px solid ${BD}`, borderRadius: '24px', overflow: 'hidden', boxShadow: SHADOW }}>
          {/* Header row */}
          <div style={{ padding: '16px 80px 14px 18px', display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
            <span style={{ fontFamily: 'Lora,serif', fontSize: '20px', color: T, lineHeight: 1, whiteSpace: 'nowrap' }}>Твой путь сегодня</span>
            {/* Decorative gold arrow + dotted trail */}
            <span style={{ color: GOLD, fontSize: '16px', lineHeight: 1, flexShrink: 0 }}>←</span>
            <svg width="52" height="10" viewBox="0 0 52 10" fill="none" style={{ flexShrink: 0 }}>
              <circle cx="3"  cy="5" r="1.6" fill={GOLD} opacity="0.85" />
              <circle cx="13" cy="4" r="1.4" fill={GOLD} opacity="0.68" />
              <circle cx="23" cy="3" r="1.3" fill={GOLD} opacity="0.52" />
              <circle cx="33" cy="3" r="1.1" fill={GOLD} opacity="0.38" />
              <circle cx="43" cy="3" r="0.9" fill={GOLD} opacity="0.24" />
            </svg>
          </div>
          {/* Task rows */}
          {tasks.map((t, i) => (
            <button key={i} onClick={t.fn} style={{
              width: '100%', display: 'flex', alignItems: 'center', gap: '14px',
              padding: '13px 18px',
              background: 'none', border: 'none', cursor: 'pointer', textAlign: 'left',
              borderTop: `1px solid ${BD}`,
            }}>
              {t.icon}
              <div style={{ flex: 1 }}>
                <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '15px', color: T, lineHeight: 1.25, marginBottom: '2px' }}>{t.title}</div>
                <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: S }}>{t.sub}</div>
              </div>
              <ChevronRight s={16} c={isDark ? 'rgba(200,191,169,0.4)' : 'rgba(111,104,92,0.4)'} />
            </button>
          ))}
        </div>
      </div>

      {/* ─── Plaza Clara ────────────────────────────────────────────────── */}
      <div style={{ margin: '10px 16px 0' }}>
        <button onClick={() => go('district:plaza')} style={{
          width: '100%', background: CB, border: `1px solid ${BD}`, borderRadius: '24px',
          overflow: 'hidden', cursor: 'pointer', textAlign: 'left',
          display: 'block', padding: 0, position: 'relative',
          boxShadow: SHADOW,
        }}>
          {/* District illustration — left ~42% */}
          <div style={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: '44%', background: districtPhBg }}>
            {/* Fade overlay to right */}
            <div style={{ position: 'absolute', inset: 0, background: districtOverlay }} />
          </div>
          {/* Content — right side */}
          <div style={{ position: 'relative', zIndex: 2, padding: '14px 14px 14px calc(44% + 12px)', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', minHeight: 160 }}>
            <div>
              <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: S, marginBottom: '2px' }}>Текущий район</div>
              <div style={{ fontFamily: 'Lora,serif', fontSize: '26px', color: T, lineHeight: 1.1, marginBottom: '3px' }}>Plaza Clara</div>
              <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: S }}>Сердце города встречает тебя.</div>
            </div>
            {/* Area icons */}
            <div style={{ display: 'flex', gap: '5px', marginTop: '14px' }}>
              {areas.map((a, i) => (
                <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '2px' }}>
                  <div style={{
                    width: '100%', maxWidth: 44, height: 40, borderRadius: 10,
                    background: 'var(--icon-bg)', border: '1px solid var(--icon-border)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '17px',
                  }}>{a.icon}</div>
                  <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '9.5px', color: T, fontWeight: 500, textAlign: 'center', lineHeight: 1.2, marginTop: '1px' }}>{a.label}</div>
                  <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '9.5px', color: S, textAlign: 'center' }}>{a.pct}%</div>
                  <div style={{ width: '100%', height: 3, borderRadius: 2, background: 'var(--progress-track)', overflow: 'hidden' }}>
                    <div style={{ height: '100%', width: `${a.pct}%`, background: a.bar, borderRadius: 2 }} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </button>
      </div>

      {/* ─── 24 слова ──────────────────────────────────────────────────── */}
      <div style={{ margin: '10px 16px 0' }}>
        <div style={{
          background: CB, border: `1px solid ${BD}`, borderRadius: '22px',
          padding: '16px', display: 'flex', alignItems: 'center', gap: '14px',
          boxShadow: SHADOW,
        }}>
          {/* Counter box */}
          <div style={{
            width: 58, height: 60, borderRadius: '14px', flexShrink: 0,
            background: 'var(--surface-2)', border: `1px solid ${BD}`,
            display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
          }}>
            <span style={{ fontFamily: 'Lora,serif', fontSize: '26px', color: T, lineHeight: 1 }}>24</span>
            <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '9px', color: S, marginTop: '2px' }}>слова</span>
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '15px', color: T, marginBottom: '4px', lineHeight: 1.3 }}>
              24 слова пора повторить
            </div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: S, lineHeight: 1.5 }}>
              Регулярное повторение —{'\n'}ключ к запоминанию.
            </div>
          </div>
          <button onClick={() => go('exercise:0')} style={{
            padding: '12px 20px', borderRadius: '14px', border: '1px solid var(--btn-border)',
            background: 'var(--btn-gradient)', color: 'white', flexShrink: 0,
            fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '15px', cursor: 'pointer',
            whiteSpace: 'nowrap', boxShadow: 'var(--btn-shadow)',
          }}>Начать</button>
        </div>
      </div>

      {/* ─── Совет от Lumi ──────────────────────────────────────────────── */}
      <div style={{ margin: '10px 16px 0' }}>
        <div style={{
          background: CB, border: `1px solid ${BD}`, borderRadius: '20px',
          padding: '14px 14px 14px 12px', display: 'flex', alignItems: 'center', gap: '6px',
          position: 'relative', overflow: 'hidden',
          boxShadow: SHADOW,
        }}>
          {/* Faint city silhouette bg */}
          <div style={{
            position: 'absolute', right: 0, top: 0, bottom: 0, width: '45%',
            background: isDark
              ? 'linear-gradient(to left,rgba(25,40,28,0.7) 0%,transparent 100%)'
              : 'linear-gradient(to left,rgba(245,235,210,0.6) 0%,transparent 100%)',
            pointerEvents: 'none',
          }}>
            <span style={{ position: 'absolute', right: 14, bottom: 8, fontSize: '46px', opacity: isDark ? 0.08 : 0.10 }}>🏙</span>
          </div>
          {/* Lumi */}
          <LumiSVG size={74} />
          <div style={{ flex: 1, position: 'relative', zIndex: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '5px' }}>
              <span style={{ color: GOLD, fontSize: '13px' }}>✦</span>
              <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '13px', color: GOLD }}>Lumi знает</span>
            </div>
            <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '14px', color: T, lineHeight: 1.55 }}>
              {LUMI_FACTS[Math.floor(Date.now() / 86400000) % LUMI_FACTS.length]}
            </p>
          </div>
        </div>
      </div>

    </div>
  );
}

// ─── Progress Screen — moved to linglow-screens-progress.jsx ────────────────
// (loaded after this file; its Object.assign overwrites this stub)
function ProgressScreen({ go }) {
  const { user, progressStats } = window.LINGLOW_DATA;
  const days = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
  const { isDesktop } = React.useContext(LayoutCtx);
  const { theme } = React.useContext(ThemeCtx);
  const isDark = theme === 'dark';

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '80px' }}>
      <div style={{ padding: '20px 16px 12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ fontFamily: 'Lora,serif', fontSize: '26px', color: 'var(--text)', margin: 0 }}>Прогресс</h1>
        <StreakBadge n={user.streak} />
      </div>
      <div style={{ margin: '0 16px 14px' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px', display: 'flex', alignItems: 'center', gap: '16px', boxShadow: isDark ? '0 8px 24px rgba(0,0,0,0.25)' : '0 4px 14px rgba(86,57,22,0.07)' }}>
          <CircleRing val={user.progress} max={100} size={72} stroke={6}>
            <span style={{ fontFamily: 'Lora,serif', fontSize: '17px', color: 'var(--text)' }}>{user.progress}%</span>
          </CircleRing>
          <div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '15px', color: 'var(--text)' }}>Уровень {user.level}</div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', margin: '2px 0 8px' }}>Общий прогресс в испанском</div>
            <div style={{ display: 'flex', gap: '10px' }}>
              {[['🔥', user.streak, 'дней'], ['🌿', user.words, 'слов']].map(([ic, v, u], i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px', fontFamily: 'Inter,sans-serif' }}>
                  <span>{ic}</span>
                  <span style={{ fontWeight: 700, color: 'var(--text)' }}>{v}</span>
                  <span style={{ color: 'var(--subtext)' }}>{u}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
      {[
        { title: 'Эта неделя', content: (
          <div style={{ display: 'flex', gap: '8px', alignItems: 'flex-end', height: '80px' }}>
            {progressStats.weekly.map((v, i) => (
              <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px', height: '100%', justifyContent: 'flex-end' }}>
                <div style={{ width: '100%', borderRadius: '5px 5px 0 0', height: `${v}%`, background: i === 6 ? 'var(--salvia)' : 'var(--progress-track)' }} />
                <span style={{ fontSize: '10px', fontFamily: 'Inter,sans-serif', color: 'var(--subtext)' }}>{days[i]}</span>
              </div>
            ))}
          </div>
        )},
        { title: 'По районам', content: progressStats.byDistrict.map((d, i) => (
          <div key={i} style={{ marginBottom: '10px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
              <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)' }}>{d.name}</span>
              <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', fontWeight: 600, color: 'var(--hoja)' }}>{d.pct}%</span>
            </div>
            <ProgressBar pct={d.pct} h={5} />
          </div>
        ))},
        { title: 'По навыкам', content: progressStats.bySkill.map((s, i) => (
          <div key={i} style={{ marginBottom: '10px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
              <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)' }}>{s.name}</span>
              <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', fontWeight: 600, color: 'var(--hoja)' }}>{s.pct}%</span>
            </div>
            <ProgressBar pct={s.pct} h={5} />
          </div>
        ))},
      ].map(({ title, content }, idx) => (
        <div key={idx} style={{ margin: '0 16px 14px' }}>
          <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px', boxShadow: isDark ? '0 8px 24px rgba(0,0,0,0.22)' : '0 4px 14px rgba(86,57,22,0.07)' }}>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)', marginBottom: '14px' }}>{title}</div>
            {content}
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Profile Screen ───────────────────────────────────────────────────────────
function ProfileScreen({ go, theme, onToggleTheme }) {
  const { user } = window.LINGLOW_DATA;
  const badges = ['🌿 Первый день', '🔥 7 дней подряд', '⭐ Уровень A1', '✈️ Путешественник', '📖 Грамматик'];
  const { isDesktop } = React.useContext(LayoutCtx);
  const isDark = theme === 'dark';

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '80px' }}>
      <div style={{ padding: '20px 16px 12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ fontFamily: 'Lora,serif', fontSize: '26px', color: 'var(--text)', margin: 0 }}>Профиль</h1>
        <button onClick={onToggleTheme} style={{ width: 38, height: 38, borderRadius: '50%', border: '1px solid var(--border)', background: 'var(--card-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
          {isDark ? <SunIcon s={18} c="var(--text)" /> : <MoonIcon s={18} c="var(--text)" />}
        </button>
      </div>
      <div style={{ padding: '0 16px 16px', display: 'flex', alignItems: 'center', gap: '16px' }}>
        <div style={{ width: 72, height: 72, borderRadius: '50%', background: 'var(--chip-bg)', border: '2px solid var(--salvia)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <LumiSVG size={50} />
        </div>
        <div>
          <div style={{ fontFamily: 'Lora,serif', fontSize: '20px', color: 'var(--text)' }}>{user.name}</div>
          <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', margin: '2px 0' }}>Уровень {user.level} · Испанский</div>
          <div style={{ display: 'flex', gap: '8px', marginTop: '4px' }}>
            <LevelChip label={`🔥 ${user.streak} дней`} active />
            <LevelChip label={`📚 ${user.words} слов`} />
          </div>
        </div>
      </div>
      <div style={{ margin: '0 16px 14px' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px' }}>
          <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)', marginBottom: '12px' }}>Значки</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
            {badges.map((b, i) => (
              <span key={i} style={{ padding: '6px 12px', borderRadius: '20px', background: 'var(--chip-bg)', border: '1px solid var(--border)', fontFamily: 'Inter,sans-serif', fontSize: '12px', fontWeight: 500, color: 'var(--text)' }}>{b}</span>
            ))}
          </div>
        </div>
      </div>
      {[['🌐', 'Язык интерфейса', 'Русский'], ['🎯', 'Цель на день', '20 минут'], ['🔔', 'Уведомления', 'Включены'], ['🎨', 'Тема', isDark ? 'Тёмная' : 'Светлая']].map(([ic, label, val], i) => (
        <div key={i} style={{ margin: '0 16px 8px' }}>
          <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '14px', padding: '14px 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <span style={{ fontSize: '18px' }}>{ic}</span>
              <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)' }}>{label}</span>
            </div>
            <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)' }}>{val}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

Object.assign(window, { HomeScreen, ProfileScreen });
