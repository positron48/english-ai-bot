// Linglow — Grammar Academy + Lesson Screens
const { useState } = React;

// ─── Status Badge ─────────────────────────────────────────────────────────────
function StatusBadge({ status, color }) {
  const cfg = {
    green:  { bg: 'rgba(45,107,58,0.12)', border: 'rgba(45,107,58,0.3)', text: '#2d6b3a', prefix: '🌿 ' },
    yellow: { bg: 'rgba(200,168,75,0.15)', border: 'rgba(200,168,75,0.4)', text: '#8a6800', prefix: '' },
    gray:   { bg: 'var(--chip-bg)', border: 'var(--border)', text: 'var(--subtext)', prefix: '' },
  }[color] || { bg: 'var(--chip-bg)', border: 'var(--border)', text: 'var(--subtext)', prefix: '' };

  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center',
      padding: '3px 10px', borderRadius: '20px', fontSize: '11px',
      fontFamily: 'Inter,sans-serif', fontWeight: 600,
      background: cfg.bg, border: `1px solid ${cfg.border}`, color: cfg.text,
      whiteSpace: 'nowrap',
    }}>{cfg.prefix}{status}</span>
  );
}

// ─── Chapter Icon ─────────────────────────────────────────────────────────────
function ChapterIcon({ icon }) {
  const isEmoji = icon.length > 2;
  return (
    <div style={{
      width: 48, height: 48, borderRadius: '50%', flexShrink: 0,
      background: 'var(--chip-bg)', border: '1.5px solid var(--border)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontSize: isEmoji ? '22px' : undefined,
    }}>
      {isEmoji ? icon : (
        <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '14px', color: 'var(--salvia)' }}>{icon}</span>
      )}
    </div>
  );
}

// ─── Grammar Screen ───────────────────────────────────────────────────────────
function GrammarScreen({ go }) {
  const { grammarCategories, grammarChapters } = window.LINGLOW_DATA;
  const [activeCategory, setActiveCategory] = useState('presente');
  const [activeDistrict, setActiveDistrict] = useState('plaza');
  const chapters = grammarChapters[activeCategory] || [];
  const { isDesktop } = React.useContext(LayoutCtx);

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '80px' }}>

      {/* ── Header ── */}
      <CenteredHeader title="Грамматика" streakN={window.LINGLOW_DATA.user.streak} />

      {/* ── District filter ── */}
      <div style={{ padding: '6px 16px 0' }}>
        <button style={{
          display: 'inline-flex', alignItems: 'center', gap: '5px',
          padding: '7px 14px', borderRadius: '20px',
          border: '1.5px solid var(--border)', background: 'var(--card-bg)',
          fontFamily: 'Inter,sans-serif', fontSize: '13px', fontWeight: 600,
          color: 'var(--text)', cursor: 'pointer',
        }}>
          <PinIcon s={13} c="var(--salvia)" />
          <span>Plaza Clara</span>
          <span style={{ fontSize: '10px', color: 'var(--subtext)' }}>▾</span>
        </button>
      </div>

      {/* ── Category chips ── */}
      <div style={{ margin: '12px 0 0' }}>
        <div style={{ display: 'flex', gap: '8px', paddingLeft: '16px', overflowX: 'auto', paddingBottom: '4px', scrollbarWidth: 'none' }}>
          {grammarCategories.map(cat => {
            const on = activeCategory === cat.id;
            return (
              <button key={cat.id} onClick={() => setActiveCategory(cat.id)} style={{
                flexShrink: 0, padding: '7px 16px', borderRadius: '20px', cursor: 'pointer',
                border: on ? 'none' : '1.5px solid var(--border)',
                background: on ? 'var(--salvia)' : 'var(--card-bg)',
                fontFamily: 'Inter,sans-serif', fontSize: '13px',
                fontWeight: on ? 700 : 500,
                color: on ? 'white' : 'var(--text)',
                transition: 'all .15s',
              }}>{cat.name}</button>
            );
          })}
          <div style={{ width: '16px', flexShrink: 0 }} />
        </div>
      </div>

      {/* ── Lumi intro banner ── */}
      <div style={{ margin: '14px 16px 0' }}>
        <div style={{
          background: 'var(--chip-bg)', border: '1px solid var(--border)',
          borderRadius: '18px', padding: '16px', overflow: 'hidden',
          display: 'flex', alignItems: 'center', gap: '12px', position: 'relative',
        }}>
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '18px', color: 'var(--text)', marginBottom: '6px' }}>
              ¡Hola! Я Lumi.
            </div>
            <p style={{ margin: '0 0 6px', fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', lineHeight: 1.5 }}>
              Грамматика — это ключ к свободе выражения. Немного каждый день — и испанский становится твоим.
            </p>
            <span style={{ color: '#c8a84b' }}>✦</span>
          </div>
          <LumiSVG size={72} pose="book" />
        </div>
      </div>

      {/* ── Chapters ── */}
      <div style={{ margin: '14px 16px 0' }}>
        {chapters.length === 0 ? (
          <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '32px 16px', textAlign: 'center' }}>
            <div style={{ fontSize: '36px', marginBottom: '8px' }}>🔒</div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--subtext)' }}>Скоро появится</div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {chapters.map(ch => (
              <button key={ch.id} onClick={() => go(`lesson:${ch.id}`)} style={{
                background: 'var(--card-bg)', border: '1px solid var(--border)',
                borderRadius: '16px', padding: '14px',
                cursor: 'pointer', textAlign: 'left',
                display: 'flex', alignItems: 'center', gap: '12px',
                position: 'relative', overflow: 'hidden',
              }}>
                {/* Left icon */}
                <ChapterIcon icon={ch.icon} />

                {/* Text */}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '16px', color: 'var(--text)', marginBottom: '3px', lineHeight: 1.2 }}>
                    {ch.title}
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px', marginBottom: '4px' }}>
                    <PinIcon s={11} c="var(--subtext)" />
                    <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>{ch.district}</span>
                  </div>
                  <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', lineHeight: 1.4 }}>{ch.descRu}</div>
                </div>

                {/* Right: status + arrow */}
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '8px', flexShrink: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <StatusBadge status={ch.statusRu} color={ch.statusColor} />
                    <ChevronRight s={14} c="var(--subtext)" />
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* ── Bottom Lumi strip ── */}
      <div style={{ margin: '14px 16px 0' }}>
        <div style={{
          background: 'var(--chip-bg)', border: '1px solid var(--border)',
          borderRadius: '14px', padding: '12px 14px',
          display: 'flex', alignItems: 'center', gap: '10px',
        }}>
          <span style={{ fontSize: '20px' }}>📖</span>
          <div style={{ flex: 1 }}>
            <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '13px', color: 'var(--text)' }}>Регулярность — твой лучший учитель.</span>
            <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)' }}> 5 минут грамматики в день — это прогресс!</span>
          </div>
          <span style={{ color: '#c8a84b', fontSize: '14px' }}>✦</span>
        </div>
      </div>
    </div>
  );
}

// ─── Lesson Screen ────────────────────────────────────────────────────────────
function LessonScreen({ lessonId, go }) {
  const { lesson, user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '0' }}>

      {/* ── Centered header ── */}
      <CenteredHeader
        title="Глаголы на -ar"
        onBack={() => go('grammar')}
        streakN={user.streak}
      />
      {/* District tag */}
      <div style={{ textAlign: 'center', paddingBottom: '8px' }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)' }}>
          <PinIcon s={12} c="var(--subtext)" /> Plaza Clara
        </span>
      </div>

      {/* ── Hero image ── */}
      <div style={{
        width: '100%', height: 200, position: 'relative', overflow: 'hidden',
        background: 'var(--chip-bg)',
      }}>
        <img src="uploads/img1.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
          onError={e => { e.target.style.display = 'none'; }} />
        {/* Lumi overlay */}
        <div style={{ position: 'absolute', bottom: 12, right: 16 }}>
          <LumiSVG size={80} pose="book" />
        </div>
        {/* Bottom gradient */}
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 60, background: 'linear-gradient(transparent, var(--bg))' }} />
      </div>

      {/* ── Теория ── */}
      <div style={{ margin: '14px 16px 0' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '10px' }}>
            <div style={{ width: 32, height: 32, borderRadius: '10px', background: 'var(--chip-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px' }}>📖</div>
            <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '16px', color: 'var(--text)' }}>Теория</span>
          </div>
          <p style={{ margin: '0 0 8px', fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)', lineHeight: 1.6 }}>
            Глаголы на <strong style={{ color: 'var(--salvia)' }}>-ar</strong> — это регулярные глаголы первого спряжения в испанском языке.
          </p>
          <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--subtext)', lineHeight: 1.6 }}>
            Они образуют свои формы по определённому шаблону, прибавляя личные окончания к основе глагола.
          </p>
        </div>
      </div>

      {/* ── Divider ornament ── */}
      <div style={{ textAlign: 'center', padding: '12px 0 4px', color: '#c8a84b', fontSize: '12px' }}>◆</div>

      {/* ── Conjugation table ── */}
      <div style={{ margin: '0 16px' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', overflow: 'hidden', padding: '16px' }}>
          <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)', marginBottom: '12px' }}>
            Спряжение: <span style={{ color: 'var(--salvia)' }}>hablar</span> <span style={{ color: 'var(--subtext)', fontWeight: 400 }}>(говорить)</span>
          </div>
          <div style={{ borderRadius: '10px', overflow: 'hidden', border: '1px solid var(--border)' }}>
            {lesson.conjugation.rows.map(([p1, f1, p2, f2], i) => (
              <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', borderTop: i > 0 ? '1px solid var(--border)' : 'none' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '11px 14px', borderRight: '1px solid var(--border)' }}>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)' }}>{p1}</span>
                  <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '15px', color: 'var(--salvia)' }}>{f1}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '11px 14px' }}>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)' }}>{p2}</span>
                  <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '15px', color: 'var(--salvia)' }}>{f2}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── Примеры ── */}
      <div style={{ margin: '12px 16px 0' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
            <div style={{ width: 32, height: 32, borderRadius: '10px', background: 'var(--chip-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px' }}>💬</div>
            <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '16px', color: 'var(--text)' }}>Примеры</span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {lesson.examples.slice(0, 3).map(([pre, verb, post], i) => (
              <div key={i}>
                <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span style={{ color: 'var(--salvia)', fontSize: '10px' }}>🌿</span>
                  <span>{pre}<strong style={{ color: 'var(--salvia)' }}>{verb}</strong>{post}</span>
                </div>
                <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', marginTop: '2px', paddingLeft: '18px' }}>
                  {['Я говорю по-испански каждый день.', 'Ты говоришь по-английски?', 'Они разговаривают с учителем.'][i]}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── Lumi fact ── */}
      <div style={{ margin: '12px 16px 0' }}>
        <LumiFactCard lumiSize={44} />
      </div>

      {/* ── CTA ── */}
      <div style={{ margin: '16px 16px 0' }}>
        <PrimaryBtn onClick={() => go('exercise:0')}>
          Практика  ›
        </PrimaryBtn>
      </div>

      {/* ── Ornament footer ── */}
      <div style={{ textAlign: 'center', padding: '16px 0 20px', color: '#c8a84b', fontSize: '13px', letterSpacing: '4px' }}>❧ ✦ ❧</div>
    </div>
  );
}

Object.assign(window, { GrammarScreen, LessonScreen });
