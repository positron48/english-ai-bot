// Linglow — Practice Screen (compact, matches mockup, theme-aware)
const { useState } = React;

function PracticeScreen({ go }) {
  const { user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const { theme }     = React.useContext(ThemeCtx);
  const isDark        = theme === 'dark';

  const GOLD   = isDark ? '#D9B25C' : '#D9A83F';
  const T_card = isDark
    ? '0 16px 40px rgba(0,0,0,0.32), inset 0 1px 0 rgba(255,255,255,0.035)'
    : '0 8px 20px rgba(86,57,22,0.07), inset 0 1px 0 rgba(255,255,255,0.75)';
  const T_bg   = isDark ? 'rgba(16,28,21,0.94)' : 'rgba(255,249,237,0.94)';
  const T_brd  = '1px solid var(--border)';

  const pageBg = isDark
    ? `radial-gradient(circle at 80% 8%, rgba(220,174,74,0.10), transparent 28%),
       radial-gradient(circle at 20% 35%, rgba(92,151,86,0.10), transparent 32%),
       #07110D`
    : `radial-gradient(circle at 80% 6%, rgba(217,178,92,0.12), transparent 26%),
       radial-gradient(circle at 18% 32%, rgba(127,169,104,0.10), transparent 32%),
       linear-gradient(180deg, #FFF8EC 0%, #F8F1E4 100%)`;

  const artMask = 'linear-gradient(to right, transparent 0%, black 28%)';

  const quickLaunches = [
    { emoji:'📝', title:'Тренировка слов — В кафе',              sub:'Набор Plaza Clara · изучено 18 из 40 слов', action:() => go('exercise:0') },
    { emoji:'🎧', title:'Текст + тест — Diálogo en la estación', sub:'2 минуты · 5 вопросов',                    action:() => go('reading-text:dialogo-estacion') },
    { emoji:'☕', title:'AI-практика — Заказать кофе',           sub:'Ситуационный диалог',                      action:() => go('chat') },
  ];

  const modes = [
    { icon:'💬', bg:'rgba(63,111,63,0.14)',   title:'Общение',    desc:'Практикуй разговор в реальных ситуациях.', art:'assets/dist_cafeterias.jpg', action:() => go('chat')    },
    { icon:'📖', bg:'rgba(217,168,63,0.14)',  title:'Слова',      desc:'Расширяй словарный запас шаг за шагом.',   art:'assets/bldg_repaso.jpg',      action:() => go('grammar') },
    { icon:'🧩', bg:'rgba(155,143,212,0.14)', title:'Грамматика', desc:'Понимай правила и применяй легко.',        art:'assets/bldg_grammar.jpg',     action:() => go('grammar') },
    { icon:'📄', bg:'rgba(91,158,212,0.14)',  title:'Тексты',     desc:'Читай и понимай на испанском языке.',      art:'assets/bldg_lectura.jpg',     action:() => go('reading') },
  ];

  const dictStats = [
    { label:'Добавлено', val:824, icon:'⊕', color: isDark ? '#A8D07F' : '#3F6F3F' },
    { label:'Новых',     val:36,  icon:'★', color: GOLD },
    { label:'Изучается', val:148, icon:'◷', color:'#9B8FD4' },
    { label:'Изучено',   val:640, icon:'✓', color: isDark ? '#A8D07F' : '#3F6F3F' },
  ];

  const fact = LUMI_FACTS[Math.floor(Date.now() / 86400000) % LUMI_FACTS.length];

  return (
    <div style={{ overflowY:'auto', height:'100%', paddingBottom: isDesktop ? 32 : 96, background: pageBg }}>

      {/* ─── BRAND + SCREEN TITLE (one line) ────────────────────────────── */}
      <div style={{ padding:'16px 20px 0', display:'flex', justifyContent:'space-between', alignItems:'flex-start' }}>
        <div style={{ display:'flex', alignItems:'baseline', gap:8, flexWrap:'wrap' }}>
          <span style={{ fontFamily:'Lora,serif', fontSize:34, color:'var(--text)', letterSpacing:'-0.02em', lineHeight:1 }}>Linglow.</span>
          <span style={{ fontFamily:'Lora,serif', fontSize:34, fontWeight:600, color:'var(--text)', letterSpacing:'-0.03em', lineHeight:1 }}>Практика</span>
        </div>
        <LumiSVG size={52} />
      </div>
      <div style={{ padding:'3px 20px 0' }}>
        <p style={{ margin:0, fontFamily:'Inter,sans-serif', fontSize:13, lineHeight:1.3, color:'var(--subtext)' }}>
          Быстрый запуск тренировок.
        </p>
      </div>

      {/* ─── HERO CARD ───────────────────────────────────────────────────── */}
      <div style={{ padding:'10px 16px 0' }}>
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, padding:10, borderRadius:20, background:T_bg, border:T_brd, boxShadow:T_card }}>
          <img src="assets/bldg_grammar.jpg" alt="Plaza Clara"
            style={{ width:'100%', aspectRatio:'4/3', objectFit:'cover', borderRadius:14, display:'block', opacity: isDark ? 0.85 : 1 }}
            onError={e => { e.target.style.background = isDark ? 'rgba(63,111,63,0.15)' : 'rgba(63,111,63,0.08)'; }} />
          <div style={{ display:'flex', flexDirection:'column' }}>
            <span style={{ alignSelf:'flex-start', display:'inline-flex', alignItems:'center', height:22, padding:'0 9px', borderRadius:999, background: isDark ? 'rgba(63,111,63,0.22)' : '#E5F0DA', color: isDark ? '#A8D07F' : '#3F6F3F', fontFamily:'Inter,sans-serif', fontSize:11, fontWeight:500 }}>
              Грамматика
            </span>
            <div style={{ marginTop:7, fontFamily:'Lora,serif', fontSize:16, lineHeight:1.05, fontWeight:600, color:'var(--text)' }}>
              Тренировка грамматики
            </div>
            <div style={{ marginTop:2, fontFamily:'Inter,sans-serif', fontSize:11, color:'var(--subtext)' }}>
              Тест по пройденным главам
            </div>
            <div style={{ marginTop:6, display:'flex', alignItems:'center', gap:3, fontFamily:'Inter,sans-serif', fontSize:11, color:'var(--subtext)' }}>
              <PinIcon s={10} c="var(--subtext)" /> Plaza Clara
            </div>
            <div style={{ marginTop:6 }}>
              <div style={{ height:5, borderRadius:999, background:'var(--progress-track)', overflow:'hidden' }}>
                <div style={{ height:'100%', width:'60%', background:'linear-gradient(90deg, var(--salvia), var(--hoja))', borderRadius:999 }} />
              </div>
              <div style={{ marginTop:2, fontFamily:'Inter,sans-serif', fontSize:10, color:'var(--subtext)' }}>6/10 глав</div>
            </div>
            <button onClick={() => go('exercise:0')} style={{ marginTop:8, alignSelf:'flex-start', display:'inline-flex', alignItems:'center', gap:2, background:'none', border:'none', cursor:'pointer', padding:0, fontFamily:'Inter,sans-serif', fontSize:12, fontWeight:600, color:'var(--salvia)' }}>
              Начать тренировку <ChevronRight s={11} c="var(--salvia)" />
            </button>
          </div>
        </div>
      </div>

      {/* ─── QUICK LAUNCHES ──────────────────────────────────────────────── */}
      <div style={{ padding:'6px 16px 0', display:'flex', flexDirection:'column', gap:6 }}>
        {quickLaunches.map((item, i) => (
          <button key={i} onClick={item.action} style={{
            display:'grid', gridTemplateColumns:'44px 1fr 16px',
            alignItems:'center', gap:10, minHeight:52, padding:'7px 12px',
            borderRadius:16, background:T_bg, border:T_brd, boxShadow:T_card,
            textAlign:'left', cursor:'pointer',
          }}>
            <div style={{ width:44, height:44, borderRadius:11, background:'var(--icon-bg)', display:'flex', alignItems:'center', justifyContent:'center', fontSize:20, flexShrink:0 }}>
              {item.emoji}
            </div>
            <div>
              <div style={{ fontFamily:'Inter,sans-serif', fontSize:13, fontWeight:500, color:'var(--text)', lineHeight:1.2 }}>{item.title}</div>
              <div style={{ marginTop:2, fontFamily:'Inter,sans-serif', fontSize:11, color:'var(--subtext)', lineHeight:1.3 }}>{item.sub}</div>
            </div>
            <ChevronRight s={14} c="var(--subtext)" />
          </button>
        ))}
      </div>

      {/* ─── MODE GRID 2×2 ───────────────────────────────────────────────── */}
      <div style={{ padding:'6px 16px 0' }}>
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:7 }}>
          {modes.map((mode, i) => (
            <button key={i} onClick={mode.action} style={{
              position:'relative', minHeight:112, padding:12,
              borderRadius:18, background:T_bg, border:T_brd, boxShadow:T_card,
              overflow:'hidden', cursor:'pointer', textAlign:'left',
            }}>
              {/* Text content — z-index keeps it above image */}
              <div style={{ position:'relative', zIndex:1, width:36, height:36, borderRadius:11, background:mode.bg, display:'flex', alignItems:'center', justifyContent:'center', fontSize:18 }}>
                {mode.icon}
              </div>
              <div style={{ position:'relative', zIndex:1, marginTop:6, fontFamily:'Lora,serif', fontSize:15, lineHeight:1.1, fontWeight:600, color:'var(--text)' }}>
                {mode.title}
              </div>
              <div style={{ position:'relative', zIndex:1, marginTop:2, maxWidth:'60%', fontFamily:'Inter,sans-serif', fontSize:10, lineHeight:1.3, color:'var(--subtext)' }}>
                {mode.desc}
              </div>
              {/* Watercolor: full card height, right side, fades in from left */}
              <img src={mode.art} alt=""
                style={{
                  position:'absolute', right:0, top:0,
                  width:'55%', height:'100%',
                  objectFit:'cover', objectPosition:'top',
                  opacity: isDark ? 0.40 : 0.88,
                  borderRadius:'0 18px 18px 0',
                  maskImage: artMask,
                  WebkitMaskImage: artMask,
                }}
                onError={e => { e.target.style.display='none'; }} />
            </button>
          ))}
        </div>
      </div>

      {/* ─── MY DICTIONARY ───────────────────────────────────────────────── */}
      <div style={{ padding:'6px 16px 0' }}>
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1.9fr', gap:8, padding:'12px 14px', borderRadius:18, background:T_bg, border:T_brd, boxShadow:T_card }}>
          {/* Left: icon + title + subtitle */}
          <div style={{ display:'flex', flexDirection:'column', justifyContent:'center', gap:3 }}>
            <div style={{ fontSize:24, lineHeight:1 }}>📗</div>
            <div style={{ fontFamily:'Lora,serif', fontSize:14, fontWeight:600, color:'var(--text)', lineHeight:1.2 }}>
              Мой словарь
            </div>
            <div style={{ fontFamily:'Inter,sans-serif', fontSize:10, color:'var(--subtext)', lineHeight:1.3 }}>
              Твои слова —{'\n'}твой прогресс
            </div>
          </div>
          {/* Right: 4 stat columns */}
          <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:4, alignItems:'center' }}>
            {dictStats.map((s, i) => (
              <div key={i} style={{ textAlign:'center' }}>
                <div style={{ display:'flex', alignItems:'center', justifyContent:'center', gap:2 }}>
                  <span style={{ color:s.color, fontSize:11 }}>{s.icon}</span>
                  <span style={{ fontFamily:'Inter,sans-serif', fontSize:9, color:'var(--subtext)' }}>{s.label}</span>
                </div>
                <div style={{ marginTop:3, fontFamily:'Lora,serif', fontSize:20, lineHeight:1, fontWeight:600, color:'var(--text)' }}>{s.val}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ─── LUMI FACT ───────────────────────────────────────────────────── */}
      <div style={{ padding:'6px 16px 0' }}>
        <div style={{
          display:'grid', gridTemplateColumns:'50px 1fr', gap:10,
          alignItems:'center', padding:'9px 12px', borderRadius:16,
          background: isDark
            ? 'linear-gradient(90deg, rgba(11,23,17,0.96), rgba(16,28,21,0.88))'
            : 'linear-gradient(90deg, rgba(255,249,237,0.96), rgba(255,244,226,0.84))',
          border: T_brd,
          boxShadow:'0 6px 16px rgba(86,57,22,0.06), inset 0 1px 0 rgba(255,255,255,0.07)',
        }}>
          <LumiSVG size={46} pose="book" />
          <div>
            <div style={{ display:'flex', alignItems:'center', gap:3, fontFamily:'Inter,sans-serif', fontSize:10, fontWeight:600, color: isDark ? GOLD : '#8A6427' }}>
              <span style={{ color:GOLD, fontSize:9 }}>✦</span> Lumi знает
            </div>
            <p style={{ margin:'2px 0 0', fontFamily:'Inter,sans-serif', fontSize:11, lineHeight:1.45, color:'var(--text)' }}>
              {fact}
            </p>
          </div>
        </div>
      </div>

    </div>
  );
}

Object.assign(window, { PracticeScreen });
