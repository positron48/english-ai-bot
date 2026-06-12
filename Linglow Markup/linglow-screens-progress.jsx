// Linglow — Progress / Statistics Screen (full redesign)

// ─── Skill icon: grammar / words / reading / speaking ────────────────────────
function PrgSkillIcon({ name, size = 18, c = '#3F6F3F' }) {
  if (name === 'Грамматика') return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M12 2v4M8 6h8M9 10h6M10.5 14h3" stroke={c} strokeWidth="1.8" strokeLinecap="round"/>
      <rect x="6" y="17" width="12" height="3.5" rx="1.8" fill={c} opacity="0.85"/>
    </svg>
  );
  if (name === 'Слова') return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M12 4L4 8V13C4 17.4 7.6 21 12 22C16.4 21 20 17.4 20 13V8L12 4Z"
        stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M9 12L11 14L15 10" stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  );
  if (name === 'Чтение') return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M2 4H8C9.06 4 10.08 4.42 10.83 5.17C11.58 5.92 12 6.94 12 8V20C12 19.21 11.68 18.45 11.12 17.88C10.55 17.32 9.79 17 9 17H2V4Z"
        stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
      <path d="M22 4H16C14.94 4 13.92 4.42 13.17 5.17C12.42 5.92 12 6.94 12 8V20C12 19.21 12.32 18.45 12.88 17.88C13.45 17.32 14.21 17 15 17H22V4Z"
        stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  );
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <path d="M21 15C21 15.5 20.79 16.04 20.41 16.41C20.04 16.79 19.53 17 19 17H7L3 21V5C3 4.47 3.21 3.96 3.59 3.59C3.96 3.21 4.47 3 5 3H19C19.53 3 20.04 3.21 20.41 3.59C20.79 3.96 21 4.47 21 5V15Z"
        stroke={c} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
      <circle cx="8" cy="10" r="1" fill={c}/>
      <circle cx="12" cy="10" r="1" fill={c}/>
      <circle cx="16" cy="10" r="1" fill={c}/>
    </svg>
  );
}

// ─── Progress Screen ──────────────────────────────────────────────────────────
function ProgressScreen({ go }) {
  const { user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const { theme } = React.useContext(ThemeCtx);
  const isDark = theme === 'dark';

  const GOLD    = isDark ? '#D9B25C' : '#D9A83F';
  const GREEN   = '#3F6F3F';
  const HOJA    = '#7FAE6A';
  const T       = 'var(--text)';
  const S       = 'var(--subtext)';
  const BD      = 'var(--border)';
  const CB      = 'var(--card-bg)';
  const BORDER  = 'rgba(129,93,42,0.14)';
  const TRACK   = isDark ? 'rgba(255,255,255,0.10)' : 'rgba(32,53,42,0.12)';
  const SHADOW  = isDark
    ? '0 16px 40px rgba(0,0,0,0.32), inset 0 1px 0 rgba(255,255,255,0.035)'
    : '0 14px 34px rgba(86,57,22,0.09), inset 0 1px 0 rgba(255,255,255,0.75)';
  const CARD_BG = isDark ? 'rgba(16,28,21,0.94)' : 'rgba(255,249,237,0.94)';
  const IC      = isDark ? '#A8D07F' : '#3F6F3F';

  // ── static data ──────────────────────────────────────────────────────────
  const WEEK_DAYS = ['Пн','Вт','Ср','Чт','Пт','Сб','Вс'];
  const WEEK_ACT  = ['done','done','done','done','done','today','empty'];

  const DISTRICTS = [
    { name:'Plaza Clara',    status:'Отлично',      pct:88, fill:GREEN },
    { name:'Distrito Alto',  status:'Хорошо',       pct:62, fill:HOJA  },
    { name:'Barrio del Mar', status:'В процессе',   pct:38, fill:GOLD  },
    { name:'El Mercado',     status:'Только начал', pct:14, fill:'#E3D8C6' },
    { name:'Colina Verde',   status:'Скоро',        pct:0,  fill:'#E3D8C6', locked:true },
  ];

  const SKILLS = [
    { name:'Грамматика', pct:75, fill:GREEN },
    { name:'Слова',      pct:72, fill:GREEN },
    { name:'Чтение',     pct:55, fill:GOLD  },
    { name:'Общение',    pct:60, fill:HOJA  },
  ];

  const IMPROVEMENTS = [
    { skill:'Грамматика', tip:'Пора повторить времена' },
    { skill:'Чтение',     tip:'Больше коротких текстов' },
    { skill:'Слова',      tip:'Расширяй активный словарь' },
  ];

  const ACHIEVEMENTS = [
    { icon:'🔥', val:'12',  title:'Огненный старт', sub:'12 дней подряд' },
    { icon:'📚', val:'35',  title:'Любитель чтения', sub:'35 текстов'    },
    { icon:'🌿', val:'250', title:'Собиратель слов', sub:'250 слов'       },
    { icon:'🏙', val:'4',   title:'Исследователь',  sub:'4 района'       },
    { icon:'👑', val:'',    title:'Эксперт',        sub:'Скоро!', locked:true },
  ];

  const METRICS = [
    { icon:'⏱', value:'1 245', label:'минут',    sub:'учил испанский' },
    { icon:'🌿', value:'386',   label:'слов',     sub:'закрепил'       },
    { icon:'📖', value:'23',    label:'текста',   sub:'прочитал'       },
    { icon:'💬', value:'18',    label:'диалогов', sub:'с ИИ'           },
  ];

  // ── atoms ─────────────────────────────────────────────────────────────────
  // Card — radius 16, padding 14
  const Card = ({ children, xStyle = {} }) => (
    <div style={{
      position:'relative', padding:'14px', borderRadius:16,
      background:CARD_BG, border:`1px solid ${BORDER}`,
      boxShadow:SHADOW, overflow:'hidden', ...xStyle,
    }}>
      {children}
    </div>
  );

  // Card title — sparkle embedded inline so it never escapes to its own line
  const CT = ({ children, fs = 15, spk = false }) => (
    <div style={{ fontFamily:'Lora,serif', fontSize:fs, fontWeight:600, lineHeight:1.2, color:T }}>
      {children}
      {spk && <span style={{ color:GOLD, fontSize:10, marginLeft:4 }}>✦</span>}
    </div>
  );

  const CS = ({ children }) => (
    <div style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:S, marginTop:3, lineHeight:1.35 }}>
      {children}
    </div>
  );

  const Bar = ({ pct, fill, h = 5 }) => (
    <div style={{ height:h, borderRadius:999, background:TRACK, overflow:'hidden' }}>
      <div style={{ height:'100%', width:`${pct}%`, borderRadius:999, background:fill }}/>
    </div>
  );

  const ChevDown = () => (
    <svg width="10" height="10" viewBox="0 0 12 12" fill="none">
      <path d="M2 4L6 8L10 4" stroke={S} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  );

  // side padding: 16px (tighter than before to match mockup density)
  const P = '0 16px';

  return (
    <div style={{ overflowY:'auto', height:'100%', paddingBottom: isDesktop ? 32 : 96, background:'var(--bg)' }}>

      {/* ── HEADER ─────────────────────────────────────────────────────── */}
      <div style={{ padding:'16px 20px 0', display:'flex', alignItems:'baseline', gap:4 }}>
        <span style={{ fontFamily:'Lora,serif', fontSize:34, color:T, letterSpacing:'-0.02em', lineHeight:1 }}>Linglow</span>
        <span style={{ color:GOLD, fontSize:14, marginLeft:2 }}>✦</span>
      </div>

            {/* ── MONTH SUMMARY CARD ────────────────────────────────────────── */}
      <div style={{
        margin:'12px 16px 12px', position:'relative', borderRadius:18,
        overflow:'hidden', border:`1px solid ${BORDER}`, boxShadow:SHADOW,
      }}>
        <div style={{ position:'absolute', inset:0, background: isDark
          ? 'linear-gradient(90deg,rgba(11,23,17,0.98) 0%,rgba(16,28,21,0.88) 100%)'
          : 'linear-gradient(90deg,rgba(255,249,237,0.98) 0%,rgba(255,244,226,0.90) 100%)',
        }}/>
        <div style={{ position:'absolute', inset:0,
          background:'radial-gradient(circle at 84% 14%, rgba(217,168,63,0.18), transparent 32%)',
        }}/>
        <img src="assets/map_city.jpg" alt="" aria-hidden="true" style={{
          position:'absolute', right:0, top:0, width:'54%', height:'52%',
          objectFit:'cover', opacity: isDark ? 0.38 : 0.70,
          maskImage:'linear-gradient(90deg, transparent 0%, black 36%)',
          WebkitMaskImage:'linear-gradient(90deg, transparent 0%, black 36%)',
          pointerEvents:'none',
        }}/>
        <div style={{ position:'absolute', right:10, top:6, zIndex:3, pointerEvents:'none' }}>
          <img src="assets/lumi.png" alt="Lumi" style={{
            width:66, height:'auto',
            filter:'drop-shadow(0 4px 10px rgba(86,57,22,0.14)) drop-shadow(0 0 14px rgba(255,190,80,0.22))',
          }}/>
        </div>

        <div style={{ position:'relative', zIndex:2, padding:'16px 16px 14px' }}>
          <div style={{ fontFamily:'Lora,serif', fontSize:18, color:T, fontWeight:600, lineHeight:1.15, marginBottom:2 }}>
            Мой испанский за месяц <span style={{ color:GOLD, fontSize:12 }}>✦</span>
          </div>
          <div style={{ fontFamily:'Inter,sans-serif', fontSize:12, color:S, marginBottom:14 }}>
            Отличный прогресс! Продолжай в том же духе!
          </div>

          <div style={{ display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:7 }}>
            {METRICS.map((m, i) => (
              <div key={i} style={{
                padding:'9px 4px', borderRadius:10,
                background: isDark ? 'rgba(32,53,42,0.42)' : 'rgba(255,249,237,0.74)',
                border:`1px solid ${isDark ? 'rgba(255,255,255,0.06)' : 'rgba(129,93,42,0.11)'}`,
                textAlign:'center', display:'flex', flexDirection:'column', alignItems:'center',
              }}>
                <span style={{ fontSize:15, lineHeight:1 }}>{m.icon}</span>
                <div style={{ marginTop:4, fontFamily:'Lora,serif', fontSize:19, lineHeight:1, fontWeight:600, color:T }}>{m.value}</div>
                <div style={{ marginTop:2, fontFamily:'Inter,sans-serif', fontSize:9, lineHeight:1.3, color:S }}>{m.label}</div>
                <div style={{ fontFamily:'Inter,sans-serif', fontSize:9, lineHeight:1.3, color:S }}>{m.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── ROW 2: RHYTHM + DISTRICTS ─────────────────────────────────── */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, padding:P }}>

        <Card>
          <CT fs={15} spk>Текущий ритм</CT>
          <CS>Отличная неделя!</CS>

          <div style={{ display:'flex', justifyContent:'space-between', marginTop:12 }}>
            {WEEK_DAYS.map((d, i) => {
              const st = WEEK_ACT[i];
              return (
                <div key={i} style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:3 }}>
                  <div style={{
                    width:22, height:22, borderRadius:'50%',
                    display:'flex', alignItems:'center', justifyContent:'center', flexShrink:0,
                    background: st==='done' ? GREEN : st==='today' ? '#FFF0C7' : (isDark ? 'rgba(32,53,42,0.35)' : '#F5E9D4'),
                    border: st==='today' ? `1.5px dashed ${GOLD}` : 'none',
                  }}>
                    {st==='done'
                      ? <svg width="9" height="9" viewBox="0 0 24 24" fill="none">
                          <path d="M5 12L10 17L19 7" stroke="white" strokeWidth="2.8" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                      : <span style={{ fontSize:7, color: st==='today' ? GOLD : (isDark ? '#555' : '#B9B0A1') }}>○</span>
                    }
                  </div>
                  <span style={{ fontFamily:'Inter,sans-serif', fontSize:8, color:S }}>{d}</span>
                </div>
              );
            })}
          </div>

          <div style={{ marginTop:12, display:'flex', alignItems:'baseline', gap:3 }}>
            <span style={{ fontFamily:'Lora,serif', fontSize:30, color:T, fontWeight:600 }}>6</span>
            <span style={{ fontFamily:'Inter,sans-serif', fontSize:12, color:S }}>дней подряд</span>
            <span style={{ color:GOLD, fontSize:10, marginLeft:2 }}>✦</span>
          </div>
          <div style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:S, marginTop:1 }}>
            Продолжай в том же духе!
          </div>
        </Card>

        <Card>
          <CT fs={15}>Районы города</CT>
          <CS>Твой прогресс по районам</CS>
          <div style={{ marginTop:12, display:'flex', flexDirection:'column', gap:8 }}>
            {DISTRICTS.map((d, i) => (
              <div key={i} style={{ opacity: d.locked ? 0.5 : 1 }}>
                <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:3 }}>
                  <span style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:T, fontWeight:500 }}>{d.name}</span>
                  <span style={{ fontFamily:'Inter,sans-serif', fontSize:9.5, color:S, flexShrink:0, marginLeft:4 }}>
                    {d.locked ? '🔒' : d.status}
                  </span>
                </div>
                <Bar pct={d.pct} fill={d.fill}/>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* ── ROW 3: FAVORITE ZONE + STRONGEST SKILL ────────────────────── */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, padding:'8px 16px 0' }}>

        <Card xStyle={{ minHeight:180 }}>
          <CT fs={14} spk>Любимая зона</CT>
          <div style={{ fontFamily:'Lora,serif', fontSize:13, color:T, fontWeight:600, marginTop:4 }}>Plaza Clara</div>
          <CS>Твоё место силы!</CS>

          <div style={{ marginTop:12 }}>
            <div style={{
              width:66, height:66, borderRadius:'50%',
              background:`conic-gradient(${GOLD} 0deg 270deg, ${isDark ? '#2C2010' : '#EFE2CC'} 270deg 360deg)`,
              display:'flex', alignItems:'center', justifyContent:'center',
            }}>
              <div style={{
                width:50, height:50, borderRadius:'50%',
                background: isDark ? '#101C15' : '#FFF9ED',
                display:'flex', flexDirection:'column', alignItems:'center', justifyContent:'center',
              }}>
                <span style={{ fontFamily:'Lora,serif', fontSize:14, color:T, fontWeight:600, lineHeight:1 }}>75%</span>
                <span style={{ fontFamily:'Inter,sans-serif', fontSize:8, color:S }}>прогресс</span>
              </div>
            </div>
          </div>

          <img src="assets/map_city.jpg" alt="" aria-hidden="true" style={{
            position:'absolute', right:0, bottom:0, width:'68%', height:'50%',
            objectFit:'cover', opacity: isDark ? 0.24 : 0.48,
            maskImage:'linear-gradient(135deg, transparent 0%, black 42%)',
            WebkitMaskImage:'linear-gradient(135deg, transparent 0%, black 42%)',
            pointerEvents:'none',
          }}/>
        </Card>

        <Card xStyle={{ minHeight:180 }}>
          <CT fs={13} spk>Самый сильный навык</CT>

          {/* icon + name in one row */}
          <div style={{ display:'flex', alignItems:'center', gap:9, marginTop:12 }}>
            <div style={{
              width:40, height:40, flexShrink:0,
              background: isDark ? 'rgba(63,111,63,0.18)' : 'rgba(63,111,63,0.08)',
              border:`1px solid ${isDark ? 'rgba(63,111,63,0.3)' : 'rgba(63,111,63,0.14)'}`,
              borderRadius:10, display:'flex', alignItems:'center', justifyContent:'center',
            }}>
              <PrgSkillIcon name="Грамматика" size={20} c={IC}/>
            </div>
            <div>
              <div style={{ fontFamily:'Lora,serif', fontSize:15, color:T, fontWeight:600, lineHeight:1.15 }}>Грамматика</div>
              <div style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:S, marginTop:2 }}>Твоя суперсила!</div>
            </div>
          </div>

          <div style={{ marginTop:14, display:'flex', alignItems:'baseline', gap:1 }}>
            <span style={{ fontFamily:'Lora,serif', fontSize:32, color:T, fontWeight:600 }}>75</span>
            <span style={{ fontFamily:'Lora,serif', fontSize:20, color:T, fontWeight:600 }}>%</span>
            <span style={{ color:GOLD, fontSize:10, marginLeft:3 }}>✦</span>
          </div>
          <div style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:S }}>уровень</div>
        </Card>
      </div>

      {/* ── ROW 4: STRENGTHS + IMPROVEMENTS ──────────────────────────── */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:10, padding:'8px 16px 0' }}>

        <Card>
          <CT fs={14}>Твои сильные стороны</CT>
          <div style={{ marginTop:12, display:'flex', flexDirection:'column', gap:10 }}>
            {SKILLS.map((sk, i) => (
              <div key={i}>
                <div style={{ display:'flex', alignItems:'center', gap:5, marginBottom:3 }}>
                  <PrgSkillIcon name={sk.name} size={12} c={IC}/>
                  <span style={{ flex:1, fontFamily:'Inter,sans-serif', fontSize:11, color:T }}>{sk.name}</span>
                  <span style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:T, fontWeight:600 }}>{sk.pct}%</span>
                </div>
                <Bar pct={sk.pct} fill={sk.fill}/>
              </div>
            ))}
          </div>
        </Card>

        <Card>
          <CT fs={14}>Нужно подтянуть</CT>
          <div style={{ marginTop:9 }}>
            {IMPROVEMENTS.map((imp, i) => (
              <div key={i} style={{
                display:'flex', gap:7, alignItems:'flex-start',
                padding: i === 0 ? '0 0 9px' : '9px 0',
                borderBottom: i < IMPROVEMENTS.length - 1
                  ? `1px solid ${isDark ? 'rgba(255,255,255,0.07)' : 'rgba(129,93,42,0.11)'}`
                  : 'none',
              }}>
                <div style={{ paddingTop:1, flexShrink:0 }}>
                  <PrgSkillIcon name={imp.skill} size={12} c={IC}/>
                </div>
                <div>
                  <div style={{ fontFamily:'Inter,sans-serif', fontSize:11, color:T, fontWeight:500 }}>{imp.skill}</div>
                  <div style={{ fontFamily:'Inter,sans-serif', fontSize:10, color:S, lineHeight:1.35, marginTop:2 }}>{imp.tip}</div>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* ── ACHIEVEMENTS ──────────────────────────────────────────────── */}
      <div style={{
        margin:'8px 16px 0', padding:'16px', borderRadius:16,
        background:CARD_BG, border:`1px solid ${BORDER}`, boxShadow:SHADOW,
      }}>
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:14 }}>
          <div style={{ fontFamily:'Lora,serif', fontSize:18, color:T, fontWeight:600 }}>
            Достижения <span style={{ color:GOLD, fontSize:11 }}>✦</span>
          </div>
          <button style={{
            background:'none', border:'none', cursor:'pointer',
            display:'flex', alignItems:'center', gap:3,
            fontFamily:'Inter,sans-serif', fontSize:12, color:S,
          }}>
            Смотреть все
            <ChevronRight s={13} c={S}/>
          </button>
        </div>

        <div style={{ display:'grid', gridTemplateColumns:'repeat(5, 1fr)', gap:6 }}>
          {ACHIEVEMENTS.map((a, i) => (
            <div key={i} style={{ textAlign:'center', opacity: a.locked ? 0.52 : 1 }}>
              <div style={{
                width:48, height:48, margin:'0 auto 5px', borderRadius:'50%',
                background: isDark ? 'rgba(32,53,42,0.5)' : '#FFF4E2',
                border:`1px solid ${BORDER}`,
                display:'flex', alignItems:'center', justifyContent:'center', flexDirection:'column',
              }}>
                {a.locked
                  ? <span style={{ fontSize:15 }}>🔒</span>
                  : <>
                      <span style={{ fontSize:13, lineHeight:1 }}>{a.icon}</span>
                      {a.val && <span style={{ fontFamily:'Lora,serif', fontSize:10, fontWeight:700, color:T, lineHeight:1, marginTop:1 }}>{a.val}</span>}
                    </>
                }
              </div>
              <div style={{ fontFamily:'Inter,sans-serif', fontSize:10, color:T, fontWeight:500, lineHeight:1.2 }}>{a.title}</div>
              <div style={{ fontFamily:'Inter,sans-serif', fontSize:9, color:S, marginTop:1, lineHeight:1.2 }}>{a.sub}</div>
            </div>
          ))}
        </div>
      </div>

      {/* ── LUMI FACT ─────────────────────────────────────────────────── */}
      {(() => {
        const FACTS = [
          'Испанский — второй язык мира по числу носителей: более 500 млн человек говорят на нём как на родном.',
          '«Шоколад» (chocolate) — из языка ацтеков náhuatl: «xocolātl» — горький напиток из какао. Испанцы привезли его в Европу в XVI веке.',
          'В испанском 10 000 слов арабского происхождения — наследие 700 лет мавританского присутствия на Пиренейском полуострове.',
          '«Guerrilla» — испанское слово, означающее «маленькая война». Оно стало международным после наполеоновских войн.',
          '«Mariposa» (бабочка) — одно из немногих слов неизвестного происхождения в испанском. Лингвисты до сих пор спорят о его истоках.',
          'Испанский восходит к народной латыни, которую принесли римские солдаты на Пиренеи в III веке до н. э.',
          '«Plaza» (площадь) пришло из греческого «plateía» — широкая улица. Через латынь слово попало в испанский.',
          'В испанском два глагола «быть»: ser — для постоянных качеств, estar — для временных состояний. Эта разница уникальна среди романских языков.',
        ];
        const fact = FACTS[Math.floor(Date.now() / 86400000) % FACTS.length];
        return (
          <div style={{
            margin:'8px 16px 0',
            display:'grid', gridTemplateColumns:'60px 1fr', gap:10,
            alignItems:'center', padding:'12px 14px', borderRadius:16,
            background: isDark
              ? 'linear-gradient(90deg,rgba(11,23,17,0.96),rgba(16,28,21,0.88))'
              : 'linear-gradient(90deg,rgba(255,249,237,0.96),rgba(255,244,226,0.84))',
            border:`1px solid ${BORDER}`,
            boxShadow:'0 12px 28px rgba(86,57,22,0.08), inset 0 1px 0 rgba(255,255,255,0.75)',
          }}>
            <img src="assets/lumi_map.png" alt="Lumi" style={{
              width:56, height:'auto',
              filter:'drop-shadow(0 4px 10px rgba(86,57,22,0.14)) drop-shadow(0 0 14px rgba(255,190,80,0.22))',
            }}/>
            <div>
              <div style={{ display:'flex', alignItems:'center', gap:4, fontFamily:'Inter,sans-serif', fontSize:11, fontWeight:600, color: isDark ? GOLD : '#8A6427' }}>
                <span style={{ color:GOLD }}>✦</span> Lumi знает
              </div>
              <p style={{ margin:'3px 0 0', fontFamily:'Inter,sans-serif', fontSize:12, lineHeight:1.5, color:T }}>
                {fact}
              </p>
            </div>
          </div>
        );
      })()}

    </div>
  );
}

Object.assign(window, { ProgressScreen });
