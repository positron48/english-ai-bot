// Linglow — Shared UI Components
const { useState, useEffect, useContext, createContext, useRef } = React;

// ─── Theme Context ────────────────────────────────────────────────────────────
const ThemeCtx = createContext({ theme: 'light', toggle: () => {} });

// ─── Layout Context ──────────────────────────────────────────────────────────
const LayoutCtx = createContext({ isDesktop: false });

// ─── Lumi Image Character ─────────────────────────────────────────────────────
function LumiSVG({ size = 80, pose = 'default' }) {
  const { theme } = React.useContext(ThemeCtx);
  // lumi.png has transparent bg — works on both light and dark backgrounds
  const src = pose === 'book'   ? 'assets/lumi_book.png'
            : pose === 'pencil' ? 'assets/lumi_pencil.png'
            : pose === 'map'    ? 'assets/lumi_map.png'
            :                    'assets/lumi.png';
  return (
    <img src={src} width={size} alt="Lumi"
      style={{ height: 'auto', flexShrink: 0, display: 'block', objectFit: 'contain' }} />
  );
}

// ─── App Top Bar (shared across Home / Practice / Progress) ─────────────────
// full=true → [Linglow | Испанский ▾ | 🔥 N]   full=false → centered logo only
function AppTopBar({ full = false }) {
  const { user } = window.LINGLOW_DATA;
  const [langOpen, setLangOpen] = React.useState(false);
  const [lang, setLang]         = React.useState('Испанский');
  const langRef                 = React.useRef(null);
  React.useEffect(() => {
    if (!langOpen) return;
    const close = (e) => { if (langRef.current && !langRef.current.contains(e.target)) setLangOpen(false); };
    document.addEventListener('mousedown', close);
    return () => document.removeEventListener('mousedown', close);
  }, [langOpen]);
  return (
    <div style={{
      height: 50, display: 'grid',
      gridTemplateColumns: full ? '1fr auto 1fr' : '1fr',
      alignItems: 'center', padding: '0 14px',
      background: 'var(--nav-bg)', border: '1px solid var(--border)',
      borderRadius: 28, boxShadow: '0 6px 18px rgba(86,57,22,0.08)',
    }}>
      <span style={{ justifySelf: full ? 'start' : 'center', fontFamily: 'Lora,serif', fontSize: 20, fontWeight: 600, color: 'var(--salvia)' }}>
        Linglow
      </span>
      {full && <>
        <div ref={langRef} style={{ justifySelf: 'center', position: 'relative' }}>
          <button onClick={() => setLangOpen(o => !o)} style={{ display: 'flex', alignItems: 'center', gap: 5, background: 'none', border: 'none', cursor: 'pointer', fontFamily: 'Inter,sans-serif', fontSize: 15, fontWeight: 500, color: 'var(--text)' }}>
            {lang} <span style={{ fontSize: 10 }}>▾</span>
          </button>
          {langOpen && (
            <div style={{ position: 'absolute', top: '130%', left: '50%', transform: 'translateX(-50%)', background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: 12, overflow: 'hidden', zIndex: 300, minWidth: 140, boxShadow: '0 14px 34px rgba(86,57,22,0.09)' }}>
              {['Испанский', 'Английский'].map((l, i, arr) => (
                <button key={l} onClick={() => { setLang(l); setLangOpen(false); }} style={{ width: '100%', padding: '10px 14px', border: 'none', background: 'none', cursor: 'pointer', textAlign: 'left', display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontFamily: 'Inter,sans-serif', fontSize: 13, color: l === lang ? 'var(--text)' : 'var(--subtext)', fontWeight: l === lang ? 600 : 400, borderBottom: i < arr.length - 1 ? '1px solid var(--border)' : 'none' }}>
                  {l}{l === lang && <span style={{ color: 'var(--dorado)', fontSize: 11 }}>✓</span>}
                </button>
              ))}
            </div>
          )}
        </div>
        <div style={{ justifySelf: 'end', display: 'inline-flex', alignItems: 'center', gap: 6, height: 34, padding: '0 12px', borderRadius: 999, background: 'var(--streak-bg)', border: '1px solid var(--streak-border)' }}>
          <span style={{ fontSize: 15 }}>🔥</span>
          <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: 13, color: 'var(--text)' }}>{user.streak}</span>
        </div>
      </>}
    </div>
  );
}

// ─── Speech Bubble ────────────────────────────────────────────────────────────
function SpeechBubble({ text, style: extra = {} }) {
  return (
    <div style={{
      position: 'relative', background: 'var(--card-bg)',
      border: '1.5px solid var(--border)', borderRadius: '14px',
      padding: '9px 13px', fontSize: '13px', fontFamily: 'Inter,sans-serif',
      fontWeight: 500, color: 'var(--text)', lineHeight: 1.45,
      whiteSpace: 'pre-line', maxWidth: '150px',
      boxShadow: '0 2px 8px rgba(0,0,0,0.08)', ...extra,
    }}>
      {text}
    </div>
  );
}

// ─── Icons ────────────────────────────────────────────────────────────────────
function HomeIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M3 12L12 4L21 12V20C21 20.6 20.6 21 20 21H15V16H9V21H4C3.4 21 3 20.6 3 20V12Z" fill={c} />
  </svg>;
}
function MapPinIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="10" r="3" stroke={c} strokeWidth="2" />
    <path d="M12 2C8.1 2 5 5.1 5 9c0 5.3 7 13 7 13s7-7.7 7-13c0-3.9-3.1-7-7-7z"
      stroke={c} strokeWidth="2" fill="none" />
  </svg>;
}
function BookIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <rect x="4" y="3" width="16" height="18" rx="2" stroke={c} strokeWidth="2" fill="none" />
    <path d="M8 7H16M8 11H16M8 15H12" stroke={c} strokeWidth="2" strokeLinecap="round" />
  </svg>;
}
function RefreshIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M1 4v6h6" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M23 20v-6h-6" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M20.5 9A9 9 0 005.6 5.6L1 10M23 14l-4.6 4.4A9 9 0 013.5 15" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>;
}
function SendIcon({ s = 20, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M22 2L11 13" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>;
}
function MicIcon({ s = 20, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <rect x="9" y="2" width="6" height="11" rx="3" stroke={c} strokeWidth="2" fill="none"/>
    <path d="M19 10a7 7 0 01-14 0M12 19v3M8 22h8" stroke={c} strokeWidth="2" strokeLinecap="round"/>
  </svg>;
}
function PinIcon({ s = 14, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="10" r="3" stroke={c} strokeWidth="2" />
    <path d="M12 2C8.1 2 5 5.1 5 9c0 5.3 7 13 7 13s7-7.7 7-13c0-3.9-3.1-7-7-7z" stroke={c} strokeWidth="2" fill="none" />
  </svg>;
}
function HamburgerIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M3 6H21M3 12H21M3 18H21" stroke={c} strokeWidth="2" strokeLinecap="round"/>
  </svg>;
}
function ClockIcon({ s = 14, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="12" r="9" stroke={c} strokeWidth="2" fill="none"/>
    <path d="M12 7v5l3 3" stroke={c} strokeWidth="2" strokeLinecap="round"/>
  </svg>;
}
function DumbbellIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <rect x="1" y="10" width="3" height="4" rx="1" fill={c}/>
    <rect x="4" y="8" width="3" height="8" rx="1.5" fill={c}/>
    <rect x="7" y="11" width="10" height="2" rx="1" fill={c}/>
    <rect x="17" y="8" width="3" height="8" rx="1.5" fill={c}/>
    <rect x="20" y="10" width="3" height="4" rx="1" fill={c}/>
  </svg>;
}
function BarIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <rect x="3" y="12" width="4" height="9" rx="1" fill={c} />
    <rect x="10" y="7" width="4" height="14" rx="1" fill={c} />
    <rect x="17" y="3" width="4" height="18" rx="1" fill={c} />
  </svg>;
}
function UserIcon({ s = 22, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="8" r="4" stroke={c} strokeWidth="2" fill="none" />
    <path d="M4 20C4 16.7 7.6 14 12 14C16.4 14 20 16.7 20 20"
      stroke={c} strokeWidth="2" strokeLinecap="round" fill="none" />
  </svg>;
}
function ChevronRight({ s = 16, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M9 6L15 12L9 18" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>;
}
function ChevronLeft({ s = 16, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M15 6L9 12L15 18" stroke={c} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>;
}
function XIcon({ s = 16, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M6 6L18 18M6 18L18 6" stroke={c} strokeWidth="2" strokeLinecap="round" />
  </svg>;
}
function CheckIcon({ s = 16, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M5 12L10 17L19 7" stroke={c} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>;
}
function SoundIcon({ s = 20, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M11 5L6 9H2V15H6L11 19V5Z" fill={c} />
    <path d="M19.1 4.9A10 10 0 0 1 19.1 19.1M15.5 8.5A5 5 0 0 1 15.5 15.5"
      stroke={c} strokeWidth="2" fill="none" strokeLinecap="round" />
  </svg>;
}
function MoonIcon({ s = 18, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" fill={c} />
  </svg>;
}
function SunIcon({ s = 18, c = 'currentColor' }) {
  return <svg width={s} height={s} viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="12" r="5" stroke={c} strokeWidth="2" fill="none" />
    <path d="M12 2V4M12 20V22M2 12H4M20 12H22M5.6 5.6L7 7M17 17L18.4 18.4M18.4 5.6L17 7M7 17L5.6 18.4"
      stroke={c} strokeWidth="2" strokeLinecap="round" />
  </svg>;
}

// ─── Progress Bar ─────────────────────────────────────────────────────────────
function ProgressBar({ pct, h = 4, color, style: ex = {} }) {
  const w = Math.min(100, Math.max(0, pct));
  return (
    <div style={{ height: h, borderRadius: h, background: 'var(--progress-track)', overflow: 'hidden', ...ex }}>
      <div style={{
        height: '100%', width: `${w}%`, borderRadius: h,
        background: color || 'var(--salvia)', transition: 'width .5s ease',
      }} />
    </div>
  );
}

// ─── Streak Badge ─────────────────────────────────────────────────────────────
function StreakBadge({ n }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', gap: '5px',
      padding: '7px 14px 7px 10px', borderRadius: '999px',
      background: 'var(--streak-bg)', border: '1px solid var(--streak-border)',
      fontFamily: 'Inter,sans-serif', fontWeight: 700,
      color: 'var(--text)', flexShrink: 0,
    }}>
      <span style={{ fontSize: '18px', lineHeight: 1 }}>🔥</span>
      <span style={{ fontSize: '14px' }}>{n} дней</span>
    </div>
  );
}

// ─── Level Chip ───────────────────────────────────────────────────────────────
function LevelChip({ label, active = false }) {
  return (
    <span style={{
      padding: '3px 10px', borderRadius: '8px', fontSize: '12px',
      fontFamily: 'Inter,sans-serif', fontWeight: 700, display: 'inline-block',
      background: active ? 'var(--salvia)' : 'var(--chip-bg)',
      color: active ? 'white' : 'var(--chip-text)',
    }}>{label}</span>
  );
}

// ─── New Chip ─────────────────────────────────────────────────────────────────
function NewChip() {
  return (
    <span style={{
      padding: '2px 8px', borderRadius: '8px', fontSize: '11px',
      fontFamily: 'Inter,sans-serif', fontWeight: 700,
      background: '#E05A30', color: 'white',
    }}>Nuevo</span>
  );
}

// ─── Lumi Tip Card ────────────────────────────────────────────────────────────
function LumiTip({ text }) {
  return (
    <div style={{
      display: 'flex', gap: '12px', alignItems: 'center',
      background: 'var(--card-bg)', border: '1px solid var(--border)',
      borderRadius: '16px', padding: '14px 16px',
    }}>
      <LumiSVG size={52} />
      <div>
        <div style={{ display: 'flex', gap: '5px', alignItems: 'center', marginBottom: '3px' }}>
          <span style={{ fontSize: '13px', fontWeight: 600, fontFamily: 'Inter,sans-serif', color: 'var(--text)' }}>
            Consejo de Lumi
          </span>
          <span>✨</span>
        </div>
        <p style={{ margin: 0, fontSize: '13px', lineHeight: 1.5, fontFamily: 'Inter,sans-serif', color: 'var(--subtext)' }}>
          {text}
        </p>
      </div>
    </div>
  );
}

// ─── Lumi Facts (rotating, day-based) ────────────────────────────────────────
const LUMI_FACTS = [
  'Испанский — второй язык мира по числу носителей: более 500 млн человек говорят на нём как на родном.',
  '«Шоколад» (chocolate) пришёл из языка ацтеков náhuatl: «xocolātl» — горький напиток из какао. Испанцы привезли его в Европу в XVI веке.',
  'В испанском около 10 000 слов арабского происхождения — наследие 700 лет мавританского присутствия на Пиренейском полуострове.',
  '«Guerrilla» — испанское слово, означающее «маленькая война». Оно стало международным термином после наполеоновских войн.',
  '«Mariposa» (бабочка) — одно из немногих слов неизвестного происхождения в испанском. Лингвисты до сих пор спорят о его истоках.',
  'Испанский восходит к народной латыни, которую принесли римские солдаты на Пиренеи в III веке до н. э.',
  '«Plaza» (площадь) пришло из греческого «plateía» — широкая улица. Через латынь оно попало в испанский.',
  'В испанском два глагола «быть»: ser — для постоянных качеств, estar — для временных. Такое разграничение редко встречается среди романских языков.',
  '«Tornado» — от испанского tornar (вращать). Слово вошло в английский через испанских моряков Атлантики.',
  '«Bonanza» по-испански означает «хорошая погода». В США оно прижилось в значении «удача» благодаря горнякам XIX века.',
];

function LumiFactCard({ lumiSize = 52, style: extra = {} }) {
  const fact = LUMI_FACTS[Math.floor(Date.now() / 86400000) % LUMI_FACTS.length];
  return (
    <div style={{
      display: 'flex', gap: '12px', alignItems: 'center',
      background: 'var(--card-bg)', border: '1px solid var(--border)',
      borderRadius: '16px', padding: '14px 16px', ...extra,
    }}>
      <LumiSVG size={lumiSize} />
      <div style={{ flex: 1 }}>
        <div style={{ display: 'flex', gap: '4px', alignItems: 'center', marginBottom: '4px' }}>
          <span style={{ fontSize: '11px', color: 'var(--dorado)' }}>✦</span>
          <span style={{ fontSize: '12px', fontWeight: 600, fontFamily: 'Inter,sans-serif', color: 'var(--dorado)' }}>Lumi знает</span>
        </div>
        <p style={{ margin: 0, fontSize: '13px', lineHeight: 1.5, fontFamily: 'Inter,sans-serif', color: 'var(--text)' }}>
          {fact}
        </p>
      </div>
    </div>
  );
}

// ─── Centered Page Header ────────────────────────────────────────────────────
function CenteredHeader({ title, subtitle, onBack, right, streakN }) {
  return (
    <div style={{ padding: '16px 16px 8px' }}>
      <div style={{ display: 'flex', alignItems: 'center', position: 'relative', minHeight: 40 }}>
        {onBack && (
          <button onClick={onBack} style={{
            width: 38, height: 38, borderRadius: '50%',
            border: '1px solid var(--border)', background: 'var(--card-bg)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer', flexShrink: 0, position: 'absolute', left: 0, zIndex: 1,
          }}>
            <ChevronLeft s={20} c="var(--text)" />
          </button>
        )}
        {!onBack && (
          <button style={{
            width: 38, height: 38, borderRadius: '50%',
            border: '1px solid var(--border)', background: 'var(--card-bg)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer', flexShrink: 0, position: 'absolute', left: 0, zIndex: 1,
          }}>
            <HamburgerIcon s={18} c="var(--text)" />
          </button>
        )}
        <div style={{ flex: 1, textAlign: 'center' }}>
          <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '20px', color: 'var(--text)', lineHeight: 1.2 }}>{title}</div>
        </div>
        <div style={{ position: 'absolute', right: 0, zIndex: 1 }}>
          {right || (streakN !== undefined && <StreakBadge n={streakN} />)}
        </div>
      </div>
      {subtitle && (
        <div style={{ textAlign: 'center', fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', marginTop: '4px' }}>{subtitle}</div>
      )}
    </div>
  );
}

// ─── Side Navigation (Desktop) ──────────────────────────────────────────────
function SideNav({ active, go, theme, onToggleTheme }) {
  const { user } = window.LINGLOW_DATA;
  const tabs = [
    { id: 'home',     label: 'Главная',    Icon: HomeIcon    },
    { id: 'map',      label: 'Город',      Icon: MapPinIcon  },
    { id: 'practice', label: 'Практика',  Icon: DumbbellIcon },
    { id: 'progress', label: 'Прогресс',   Icon: BarIcon     },
    { id: 'profile',  label: 'Профиль',    Icon: UserIcon    },
  ];
  const rootId = active ? active.split(':')[0] : 'home';
  return (
    <nav style={{
      width: 220, flexShrink: 0, height: '100%',
      background: 'var(--card-bg)', borderRight: '1px solid var(--border)',
      display: 'flex', flexDirection: 'column', overflow: 'hidden',
    }}>
      {/* Logo */}
      <div style={{
        padding: '28px 20px 22px',
        display: 'flex', alignItems: 'center', gap: '8px',
        borderBottom: '1px solid var(--border)',
      }}>
        <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '20px', color: 'var(--text)' }}>Linglow</span>
        <span style={{ fontSize: '16px' }}>🌿</span>
      </div>
      {/* Nav items */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '2px', padding: '14px 10px', overflowY: 'auto' }}>
        {tabs.map(({ id, label, Icon }) => {
          const practiceIds = ['practice','grammar','lesson','reading','reading-text'];
          const on = id === 'practice' ? practiceIds.includes(rootId) : rootId === id;
          return (
            <button key={id} onClick={() => go(id)} style={{
              display: 'flex', alignItems: 'center', gap: '11px',
              padding: '10px 14px', borderRadius: '12px', border: 'none',
              background: on ? 'var(--chip-bg)' : 'transparent',
              cursor: 'pointer', textAlign: 'left', width: '100%',
              position: 'relative',
            }}>
              {on && <div style={{
                position: 'absolute', left: 0, top: '18%', bottom: '18%',
                width: 3, borderRadius: '0 3px 3px 0',
                background: 'var(--salvia)',
              }} />}
              <Icon s={20} c={on ? 'var(--salvia)' : 'var(--subtext)'} />
              <span style={{
                fontFamily: 'Inter,sans-serif', fontSize: '14px',
                fontWeight: on ? 700 : 400,
                color: on ? 'var(--text)' : 'var(--subtext)',
              }}>{label}</span>
            </button>
          );
        })}
      </div>
      {/* Bottom: streak + theme toggle */}
      <div style={{
        padding: '16px 20px',
        borderTop: '1px solid var(--border)',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      }}>
        <StreakBadge n={user.streak} />
        <button onClick={onToggleTheme} style={{
          width: 36, height: 36, borderRadius: '50%',
          border: '1px solid var(--border)', background: 'var(--bg)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          cursor: 'pointer',
        }}>
          {theme === 'light' ? <MoonIcon s={16} c="var(--text)" /> : <SunIcon s={16} c="var(--text)" />}
        </button>
      </div>
    </nav>
  );
}

// ─── Bottom Navigation ────────────────────────────────────────────────────────
function BottomNav({ active, go }) {
  const [showBorder, setShowBorder] = useState(false);

  useEffect(() => {
    // Capture-phase scroll listener — catches any scrollable child
    const onScroll = (e) => {
      const el = e.target;
      if (!el || typeof el.scrollTop === 'undefined') return;
      if (el.scrollHeight <= el.clientHeight) return;
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 8;
      setShowBorder(el.scrollTop > 8 && !atBottom);
    };
    window.addEventListener('scroll', onScroll, true);
    return () => window.removeEventListener('scroll', onScroll, true);
  }, []);

  const screenId = active ? active.split(':')[0] : 'home';
  const tabs = [
    { id: 'home',     label: 'Главная',  Icon: HomeIcon     },
    { id: 'map',      label: 'Город',    Icon: MapPinIcon   },
    { id: 'practice', label: 'Практика', Icon: DumbbellIcon },
    { id: 'progress', label: 'Прогресс', Icon: BarIcon      },
    { id: 'profile',  label: 'Профиль',  Icon: UserIcon     },
  ];
  const rootId = active ? active.split(':')[0] : 'home';

  return (
    <nav style={{
      position: 'fixed', bottom: 0, left: 0, right: 0,
      width: '100%',
      background: 'var(--nav-bg)',
      backdropFilter: 'blur(18px)', WebkitBackdropFilter: 'blur(18px)',
      borderTop: showBorder ? '1px solid var(--border)' : '1px solid transparent',
      boxShadow: showBorder ? '0 -8px 28px rgba(0,0,0,0.18)' : '0 -4px 20px rgba(0,0,0,0.10)',
      display: 'flex', zIndex: 200, alignItems: 'center',
      paddingBottom: 'env(safe-area-inset-bottom,0px)',
      transition: 'border-color .2s ease',
    }}>
      {tabs.map(({ id, label, Icon }) => {
        const practiceIds = ['practice','grammar','lesson','reading','reading-text'];
        const on = id === 'practice' ? practiceIds.includes(rootId) : rootId === id;
        return (
          <button key={id} onClick={() => go(id)} style={{
            flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center',
            padding: '8px 4px 6px', border: 'none', background: 'none', cursor: 'pointer',
          }}>
            {on ? (
              <div style={{
                display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '2px',
                padding: '5px 12px 4px', borderRadius: '18px',
                background: 'var(--nav-active-bg)',
                border: '1px solid var(--nav-active-border)',
              }}>
                <Icon s={20} c="var(--nav-active-color)" />
                <span style={{ fontSize: '10px', fontFamily: 'Inter,sans-serif', color: 'var(--nav-active-color)', fontWeight: 600 }}>{label}</span>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '2px', padding: '5px 6px 4px' }}>
                <Icon s={20} c="var(--nav-inactive)" />
                <span style={{ fontSize: '10px', fontFamily: 'Inter,sans-serif', color: 'var(--nav-inactive)', fontWeight: 400 }}>{label}</span>
              </div>
            )}
          </button>
        );
      })}
    </nav>
  );
}

// ─── Back Header ──────────────────────────────────────────────────────────────
function BackHeader({ title, onBack, streak }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '16px 16px 8px',
    }}>
      <button onClick={onBack} style={{
        width: 38, height: 38, borderRadius: '50%',
        border: '1px solid var(--border)', background: 'var(--card-bg)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        cursor: 'pointer',
      }}>
        <ChevronLeft s={20} c="var(--text)" />
      </button>
      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <span style={{ fontSize: '16px', fontFamily: 'Inter,sans-serif', fontWeight: 600, color: 'var(--text)' }}>
          {title}
        </span>
        <span style={{ fontSize: '16px' }}>🌿</span>
      </div>
      <StreakBadge n={streak || 7} />
    </div>
  );
}

// ─── Primary Button ───────────────────────────────────────────────────────────
function PrimaryBtn({ children, onClick, disabled = false, style: ex = {} }) {
  return (
    <button onClick={onClick} disabled={disabled} style={{
      width: '100%', padding: '15px', borderRadius: '50px',
      border: disabled ? 'none' : '1px solid var(--btn-border)',
      cursor: disabled ? 'default' : 'pointer',
      background: disabled ? 'var(--chip-bg)' : 'var(--btn-gradient)',
      color: disabled ? 'var(--subtext)' : 'white',
      fontSize: '16px', fontFamily: 'Inter,sans-serif', fontWeight: 600,
      boxShadow: disabled ? 'none' : 'var(--btn-shadow)',
      transition: 'opacity .15s',
      ...ex,
    }}>
      {children}
    </button>
  );
}

// ─── Circle Progress Ring ─────────────────────────────────────────────────────
function CircleRing({ val, max, size = 72, stroke = 6, children }) {
  const r = (size - stroke) / 2;
  const circ = 2 * Math.PI * r;
  const offset = circ * (1 - Math.min(1, val / max));
  return (
    <div style={{ position: 'relative', width: size, height: size, flexShrink: 0 }}>
      <svg width={size} height={size} style={{ transform: 'rotate(-90deg)' }}>
        <circle cx={size / 2} cy={size / 2} r={r}
          stroke="var(--progress-track)" strokeWidth={stroke} fill="none" />
        <circle cx={size / 2} cy={size / 2} r={r}
          stroke="var(--salvia)" strokeWidth={stroke} fill="none"
          strokeDasharray={circ} strokeDashoffset={offset}
          strokeLinecap="round" />
      </svg>
      <div style={{
        position: 'absolute', inset: 0, display: 'flex',
        alignItems: 'center', justifyContent: 'center',
        flexDirection: 'column',
      }}>{children}</div>
    </div>
  );
}

// Export all
Object.assign(window, {
  ThemeCtx, LayoutCtx, LumiSVG, SpeechBubble, AppTopBar,
  HomeIcon, MapPinIcon, BookIcon, RefreshIcon, BarIcon, UserIcon, DumbbellIcon,
  ChevronRight, ChevronLeft, XIcon, CheckIcon, SoundIcon, MoonIcon, SunIcon,
  SendIcon, MicIcon, PinIcon, HamburgerIcon, ClockIcon,
  ProgressBar, StreakBadge, LevelChip, NewChip, LumiTip, LumiFactCard, LUMI_FACTS,
  CenteredHeader, SideNav, BottomNav, BackHeader, PrimaryBtn, CircleRing,
});
