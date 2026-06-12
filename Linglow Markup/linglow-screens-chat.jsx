// Linglow — AI Chat Screen
const { useState, useRef, useEffect } = React;

function ChatScreen({ go }) {
  const { chatScenario, user } = window.LINGLOW_DATA;
  const { isDesktop } = React.useContext(LayoutCtx);
  const s = chatScenario;
  const [messages, setMessages] = useState(s.messages);
  const [input, setInput] = useState('');
  const [showCorrection, setShowCorrection] = useState(true);
  const bottomRef = useRef(null);

  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.parentElement.scrollTop = bottomRef.current.offsetTop;
    }
  }, [messages]);

  const send = () => {
    if (!input.trim()) return;
    setMessages(prev => [...prev, { role: 'user', text: input.trim() }]);
    setInput('');
    // Simulate Lumi reply
    setTimeout(() => {
      setMessages(prev => [...prev, { role: 'lumi', text: '¡Muy bien! ¿Y a qué hora quieres salir?' }]);
    }, 900);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>

      {/* Header */}
      <CenteredHeader
        title={s.title}
        onBack={() => go('home')}
        streakN={user.streak}
      />
      <div style={{ textAlign: 'center', padding: '0 16px 6px' }}>
        <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)' }}>{s.subtitle}</div>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)', marginTop: '2px' }}>
          <PinIcon s={11} c="var(--subtext)" /> {s.district}
        </span>
      </div>

      {/* City illustration strip */}
      <div style={{ width: '100%', height: 100, overflow: 'hidden', background: 'var(--chip-bg)', flexShrink: 0 }}>
        <img src="uploads/img2.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', objectPosition: 'center 30%' }}
          onError={e => { e.target.style.display = 'none'; }} />
        <div style={{ position: 'relative', marginTop: -100, height: 100, background: 'linear-gradient(transparent 40%, var(--bg))' }} />
      </div>

      {/* Chat scroll area */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '8px 16px 0' }}>

        {messages.map((m, i) => (
          <div key={i} style={{
            display: 'flex', alignItems: 'flex-end', gap: '8px', marginBottom: '12px',
            flexDirection: m.role === 'user' ? 'row-reverse' : 'row',
          }}>
            {m.role === 'lumi' && <LumiSVG size={36} />}
            <div style={{ maxWidth: '75%' }}>
              {m.role === 'lumi' && (
                <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '10px', fontWeight: 600, color: 'var(--salvia)', marginBottom: '3px', display: 'flex', alignItems: 'center', gap: '3px' }}>
                  <span style={{ color: '#c8a84b' }}>✦</span> Lumi
                </div>
              )}
              <div style={{
                padding: '11px 14px', borderRadius: m.role === 'lumi' ? '4px 16px 16px 16px' : '16px 4px 16px 16px',
                background: m.role === 'lumi' ? 'var(--card-bg)' : 'var(--chip-bg)',
                border: '1px solid var(--border)',
                fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)', lineHeight: 1.5,
              }}>
                {m.text}
              </div>
              {m.role === 'user' && (
                <div style={{ textAlign: 'right', fontFamily: 'Inter,sans-serif', fontSize: '10px', color: 'var(--subtext)', marginTop: '3px' }}>
                  09:41 ✓✓
                </div>
              )}
            </div>
          </div>
        ))}

        {/* Correction card */}
        {showCorrection && (
          <div style={{
            background: 'var(--card-bg)', border: '1.5px solid var(--border)',
            borderRadius: '16px', padding: '14px 16px', marginBottom: '14px',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '7px' }}>
                <div style={{ width: 28, height: 28, borderRadius: '8px', background: 'rgba(200,168,75,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '14px' }}>🌿</div>
                <span style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '14px', color: 'var(--text)' }}>Небольшое исправление</span>
              </div>
              <button onClick={() => setShowCorrection(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px' }}>
                <XIcon s={14} c="var(--subtext)" />
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', marginBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '6px' }}>
                <span style={{ fontSize: '14px', marginTop: '1px' }}>❌</span>
                <div>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>Ваша фраза: </span>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)' }}>{s.correction.original}</span>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '6px' }}>
                <span style={{ fontSize: '14px', marginTop: '1px' }}>✅</span>
                <div>
                  <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)' }}>Лучше сказать: </span>
                  <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '13px', color: 'var(--salvia)' }}>{s.correction.corrected}</span>
                </div>
              </div>
            </div>
            <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', lineHeight: 1.5, paddingTop: '8px', borderTop: '1px solid var(--border)' }}>
              <strong style={{ color: 'var(--text)' }}>Почему: </strong>{s.correction.note}
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Action buttons */}
      <div style={{ padding: '8px 16px 6px', display: 'flex', gap: '8px' }}>
        {[
          { label: '💡 Подсказка', action: () => {} },
          { label: '✏️ Исправить', action: () => setShowCorrection(true) },
          { label: '🔖 Сохранить фразу', action: () => {} },
        ].map((btn, i) => (
          <button key={i} onClick={btn.action} style={{
            flex: 1, padding: '7px 4px', borderRadius: '20px',
            border: '1.5px solid var(--border)', background: 'var(--card-bg)',
            fontFamily: 'Inter,sans-serif', fontSize: '11px', fontWeight: 600,
            color: 'var(--text)', cursor: 'pointer', whiteSpace: 'nowrap',
          }}>{btn.label}</button>
        ))}
      </div>

      {/* Input bar */}
      <div style={{
        padding: '6px 16px 16px',
        display: 'flex', alignItems: 'center', gap: '10px',
      }}>
        <button style={{ width: 40, height: 40, borderRadius: '50%', border: '1.5px solid var(--border)', background: 'var(--card-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}>
          <MicIcon s={18} c="var(--subtext)" />
        </button>
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', background: 'var(--card-bg)', border: '1.5px solid var(--border)', borderRadius: '24px', padding: '0 14px', minHeight: 44 }}>
          <input
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && send()}
            placeholder="Напишите на испанском..."
            style={{ flex: 1, border: 'none', background: 'transparent', outline: 'none', fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--text)' }}
          />
        </div>
        <button onClick={send} style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'var(--salvia)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}>
          <SendIcon s={16} c="white" />
        </button>
      </div>

    </div>
  );
}

Object.assign(window, { ChatScreen });
