// Linglow — Reading Screens
const { useState } = React;

// ─── Reading List Screen ───────────────────────────────────────────────────────
function ReadingListScreen({ go }) {
  const { readingTexts, user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const [activeFilter, setActiveFilter] = useState('plaza');

  const filters = [
    { id: 'plaza',  label: 'Plaza Clara',         color: '#2d6b3a' },
    { id: 'puerta', label: 'Puerta de la Chispa',  color: '#b07830' },
    { id: 'travel', label: 'Путешествия',           color: '#2d5a27' },
    { id: 'cafe',   label: 'Кафе',                  color: '#8a4a30' },
  ];

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '80px' }}>

      {/* Header */}
      <CenteredHeader title="Чтение" streakN={user.streak} />
      <div style={{ textAlign: 'center', padding: '0 32px 12px' }}>
        <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', lineHeight: 1.5 }}>
          Читайте истории, учите слова и открывайте новые уголки Ciudad Luminaria.
        </p>
      </div>

      {/* Filter chips */}
      <div style={{ display: 'flex', gap: '8px', paddingLeft: '16px', overflowX: 'auto', paddingBottom: '4px', scrollbarWidth: 'none', marginBottom: '14px' }}>
        {filters.map(f => {
          const on = activeFilter === f.id;
          return (
            <button key={f.id} onClick={() => setActiveFilter(f.id)} style={{
              flexShrink: 0, display: 'flex', alignItems: 'center', gap: '5px',
              padding: '7px 14px', borderRadius: '20px', cursor: 'pointer',
              border: on ? 'none' : '1.5px solid var(--border)',
              background: on ? f.color : 'var(--card-bg)',
              fontFamily: 'Inter,sans-serif', fontSize: '12px', fontWeight: on ? 700 : 500,
              color: on ? 'white' : 'var(--text)', transition: 'all .15s',
            }}>
              {on && <BookIcon s={12} c="white" />}
              {f.label}
            </button>
          );
        })}
        <div style={{ width: '16px', flexShrink: 0 }} />
      </div>

      {/* Lumi promo card */}
      <div style={{ margin: '0 16px 14px' }}>
        <div style={{
          background: 'var(--chip-bg)', border: '1px solid var(--border)',
          borderRadius: '18px', padding: '16px', overflow: 'hidden',
          display: 'flex', alignItems: 'center', gap: '12px',
        }}>
          <LumiSVG size={64} pose="book" />
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '16px', color: 'var(--text)', marginBottom: '4px' }}>
              Начни с короткого диалога
            </div>
            <p style={{ margin: '0 0 10px', fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', lineHeight: 1.5 }}>
              Лёгкий диалог, чтобы войти в ритм испанского языка.
            </p>
            <button onClick={() => go('reading-text:dialogo-estacion')} style={{
              display: 'inline-flex', alignItems: 'center', gap: '5px',
              padding: '8px 16px', borderRadius: '20px', border: 'none',
              background: 'var(--salvia)', color: 'white',
              fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '13px', cursor: 'pointer',
            }}>Читать <ChevronRight s={13} c="white" /></button>
          </div>
        </div>
      </div>

      {/* Reading cards */}
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
        {readingTexts.map(txt => (
          <button key={txt.id} onClick={() => go(`reading-text:${txt.id}`)} style={{
            width: '100%', background: 'var(--card-bg)', border: '1px solid var(--border)',
            borderRadius: '18px', overflow: 'hidden', cursor: 'pointer', textAlign: 'left',
            display: 'flex', padding: 0, minHeight: 110,
          }}>
            <div style={{ flex: 1, padding: '14px 14px 14px 16px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '4px', marginBottom: '6px' }}>
                  <BookIcon s={12} c={txt.districtColor} />
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', fontWeight: 600, color: txt.districtColor }}>{txt.districtName}</span>
                </div>
                <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '17px', color: 'var(--text)', lineHeight: 1.25, marginBottom: '8px' }}>
                  {txt.title}
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '3px' }}>
                  <ClockIcon s={12} c="var(--subtext)" />
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>{txt.duration}</span>
                </div>
                <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>
                  {txt.known} знакомых слов · {txt.newWords} новых
                </span>
              </div>
            </div>
            <div style={{ width: 100, flexShrink: 0, overflow: 'hidden', background: 'var(--chip-bg)' }}>
              <img src={txt.img} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                onError={e => { e.target.style.display = 'none'; }} />
            </div>
          </button>
        ))}
      </div>

    </div>
  );
}

// ─── Reading Text Screen ───────────────────────────────────────────────────────
function ReadingTextScreen({ textId, go }) {
  const { readingTextDetail, user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const [showTranslation, setShowTranslation] = useState({});
  const [tooltip, setTooltip] = useState(null); // {word, meaning, lineIdx}
  const [quizAnswer, setQuizAnswer] = useState(null);

  const data = readingTextDetail[textId] || readingTextDetail['dialogo-estacion'];
  if (!data) return null;

  const toggleTranslation = (i) => setShowTranslation(prev => ({ ...prev, [i]: !prev[i] }));

  // Render a line with bold highlights
  const renderLine = (html) => {
    const parts = html.split(/(<b>.*?<\/b>)/g);
    return parts.map((p, i) => {
      if (p.startsWith('<b>')) {
        const word = p.replace(/<\/?b>/g, '');
        return (
          <span key={i}
            onClick={(e) => { e.stopPropagation(); setTooltip(tooltip?.word === word ? null : { word, meaning: '...' }); }}
            style={{
              background: 'rgba(200,168,75,0.25)', borderRadius: '4px',
              padding: '0 3px', cursor: 'pointer', fontWeight: 700,
              color: 'var(--text)', borderBottom: '2px solid rgba(200,168,75,0.6)',
            }}>{word}</span>
        );
      }
      return p;
    });
  };

  return (
    <div style={{ overflowY: 'auto', height: '100%', paddingBottom: isDesktop ? '32px' : '80px' }}
      onClick={() => setTooltip(null)}>

      {/* Header */}
      <CenteredHeader title={data.title} onBack={() => go('reading')} streakN={user.streak} />
      <div style={{ textAlign: 'center', paddingBottom: '8px' }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)' }}>
          <PinIcon s={12} c="var(--subtext)" /> {data.districtName}
        </span>
      </div>

      {/* Hero image */}
      <div style={{ width: '100%', height: 180, position: 'relative', overflow: 'hidden', background: 'var(--chip-bg)' }}>
        <img src={data.img || 'uploads/img3.png'} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
          onError={e => { e.target.style.display = 'none'; }} />
        {/* Lumi bubble */}
        <div style={{ position: 'absolute', bottom: 12, right: 12, display: 'flex', alignItems: 'flex-end', gap: '8px' }}>
          <div style={{
            background: 'rgba(255,251,240,0.95)', border: '1px solid var(--border)',
            borderRadius: '12px', padding: '8px 12px', maxWidth: 160,
            fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--text)',
            boxShadow: '0 2px 8px rgba(0,0,0,0.12)',
          }}>
            {data.lumiTip} <SoundIcon s={13} c="var(--salvia)" />
          </div>
          <LumiSVG size={56} />
        </div>
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 50, background: 'linear-gradient(transparent, var(--bg))' }} />
      </div>

      {/* Dialogue */}
      <div style={{ margin: '12px 16px 0' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', overflow: 'hidden', position: 'relative' }}>
          {data.lines.map((line, i) => (
            <div key={i} style={{ padding: '12px 16px', borderBottom: i < data.lines.length - 1 ? '1px solid var(--border)' : 'none' }}>
              <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: '8px' }}>
                <div style={{ flex: 1 }}>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '13px', color: 'var(--salvia)', marginRight: '6px' }}>{line.speaker}:</span>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)', lineHeight: 1.5 }}>
                    {renderLine(line.es)}
                  </span>
                  {showTranslation[i] && (
                    <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', marginTop: '4px', paddingLeft: '4px' }}>
                      {line.ru}
                    </div>
                  )}
                </div>
                <button onClick={() => toggleTranslation(i)} style={{
                  background: 'none', border: 'none', cursor: 'pointer', flexShrink: 0, padding: '2px',
                }}>
                  <SoundIcon s={16} c={showTranslation[i] ? 'var(--salvia)' : 'var(--subtext)'} />
                </button>
              </div>
            </div>
          ))}

          {/* Word tooltip */}
          {tooltip && (
            <div style={{
              position: 'sticky', bottom: 0, left: 0, right: 0,
              background: 'var(--card-bg)', borderTop: '1.5px solid var(--salvia)',
              padding: '12px 16px',
            }} onClick={e => e.stopPropagation()}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '6px' }}>
                <button style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}>
                  <SoundIcon s={16} c="var(--salvia)" />
                </button>
                <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '17px', color: 'var(--text)' }}>andén</span>
                <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>м.р.</span>
              </div>
              <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', marginBottom: '8px' }}>платформа (на вокзале)</div>
              <button style={{
                display: 'flex', alignItems: 'center', gap: '5px', background: 'none', border: 'none', cursor: 'pointer', padding: 0,
                fontFamily: 'Inter,sans-serif', fontSize: '13px', fontWeight: 600, color: 'var(--salvia)',
              }}>
                🔖 Сохранить слово
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Comprehension quiz */}
      <div style={{ margin: '12px 16px 0' }}>
        <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '18px', padding: '16px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '10px' }}>
            <BookIcon s={16} c="var(--salvia)" />
            <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)' }}>Проверим понимание</span>
          </div>
          <p style={{ margin: '0 0 12px', fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)' }}>{data.quiz.q}</p>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {data.quiz.options.map((opt, i) => {
              const selected = quizAnswer === i;
              const correct = i === data.quiz.correct;
              let bg = 'var(--card-bg)', border = '1.5px solid var(--border)', color = 'var(--text)';
              if (quizAnswer !== null && correct) { bg = 'rgba(45,107,58,0.1)'; border = '2px solid var(--salvia)'; color = 'var(--salvia)'; }
              else if (selected && !correct) { bg = 'rgba(220,60,60,0.08)'; border = '2px solid #E05A5A'; }
              return (
                <button key={i} onClick={() => setQuizAnswer(i)} style={{
                  padding: '8px 14px', borderRadius: '20px', cursor: 'pointer',
                  background: bg, border, color,
                  fontFamily: 'Inter,sans-serif', fontSize: '12px', fontWeight: 500, transition: 'all .15s',
                }}>{opt}</button>
              );
            })}
          </div>
        </div>
      </div>

      {/* CTA: practice in dialogue */}
      <div style={{ margin: '12px 16px 0' }}>
        <button onClick={() => go('chat')} style={{
          width: '100%', background: 'var(--card-bg)', border: '1px solid var(--border)',
          borderRadius: '18px', padding: '14px 16px', cursor: 'pointer', textAlign: 'left',
          display: 'flex', alignItems: 'center', gap: '12px', overflow: 'hidden',
        }}>
          <div style={{ width: 40, height: 40, borderRadius: '12px', background: 'var(--chip-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <MicIcon s={20} c="var(--salvia)" />
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--salvia)', marginBottom: '2px' }}>Потренировать в диалоге</div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)' }}>Поговори с Луми и закрепи новые выражения.</div>
          </div>
          <ChevronRight s={16} c="var(--subtext)" />
        </button>
      </div>

      {/* Lumi fact */}
      <div style={{ margin: '12px 16px 0' }}>
        <LumiFactCard lumiSize={44} />
      </div>

    </div>
  );
}

Object.assign(window, { ReadingListScreen, ReadingTextScreen });
