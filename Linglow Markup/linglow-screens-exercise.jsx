// Linglow — Exercise Screens (Choice, Tiles, Write)
const { useState, useEffect, useRef } = React;

// ─── Exercise Progress Header ─────────────────────────────────────────────────
function ExerciseHeader({ qNum, total, onClose, streak }) {
  const pct = (qNum / total) * 100;
  return (
    <div style={{ padding: '14px 16px 10px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
        <button onClick={onClose} style={{
          width: 36, height: 36, borderRadius: '50%', border: '1.5px solid var(--border)',
          background: 'var(--card-bg)', display: 'flex', alignItems: 'center',
          justifyContent: 'center', cursor: 'pointer', flexShrink: 0,
        }}>
          <XIcon s={16} c="var(--subtext)" />
        </button>
        <div style={{ flex: 1 }}>
          <div style={{
            display: 'flex', gap: '4px', alignItems: 'center', marginBottom: '5px',
          }}>
            {Array.from({ length: Math.min(total, 12) }).map((_, i) => (
              <div key={i} style={{
                flex: 1, height: 5, borderRadius: '3px',
                background: i < qNum ? 'var(--salvia)' : i === qNum ? 'var(--hoja)' : 'var(--progress-track)',
                transition: 'background .3s',
              }} />
            ))}
          </div>
          <div style={{ fontFamily: 'Inter,sans-serif', fontSize: '11px', color: 'var(--subtext)', textAlign: 'center' }}>
            {qNum} de {total}
          </div>
        </div>
        <StreakBadge n={streak} />
      </div>
    </div>
  );
}

// ─── Feedback Banner ──────────────────────────────────────────────────────────
function FeedbackBanner({ correct, message, points }) {
  return (
    <div style={{
      margin: '12px 16px 0',
      padding: '12px 14px',
      borderRadius: '14px',
      background: correct ? 'rgba(106,143,99,0.12)' : 'rgba(220,60,60,0.1)',
      border: `1.5px solid ${correct ? 'var(--salvia)' : '#E05A5A'}`,
      display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '10px',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <span style={{ fontSize: '18px' }}>🌿</span>
        <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)', lineHeight: 1.4 }}>
          {message}
        </span>
      </div>
      {correct && points && (
        <div style={{
          padding: '4px 10px', borderRadius: '20px',
          background: 'var(--salvia)', color: 'white',
          fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '13px',
          flexShrink: 0,
        }}>+{points} ✦</div>
      )}
    </div>
  );
}

// ─── Multiple Choice Exercise ─────────────────────────────────────────────────
function ChoiceExercise({ ex, onNext }) {
  const [selected, setSelected] = useState(null);
  const answered = selected !== null;
  const isCorrect = selected === ex.correct;

  const handleSelect = (i) => { if (!answered) setSelected(i); };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ flex: 1, overflowY: 'auto', padding: '0 16px' }}>

        {/* Question */}
        <h2 style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '22px', color: 'var(--text)', margin: '4px 0 14px', textAlign: 'center' }}>
          Как переводится «{ex.question}»?
        </h2>

        {/* Word card with city bg */}
        <div style={{ borderRadius: '18px', overflow: 'hidden', marginBottom: '16px', position: 'relative', height: 130, background: 'var(--chip-bg)', border: '1px solid var(--border)' }}>
          <img src="uploads/img1.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block', opacity: 0.65 }}
            onError={e => { e.target.style.display = 'none'; }} />
          <div style={{ position: 'absolute', inset: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
            <div style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '30px', color: '#1e1208', textShadow: '0 2px 12px rgba(255,248,220,0.95)' }}>
              {ex.question}
            </div>
            <button style={{ width: 32, height: 32, borderRadius: '50%', background: 'rgba(255,251,240,0.9)', border: '1px solid rgba(200,168,75,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
              <SoundIcon s={14} c="var(--subtext)" />
            </button>
          </div>
        </div>

        {/* Options */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {ex.options.map((opt, i) => {
            let bg = 'var(--card-bg)', border = '1.5px solid var(--border)';
            if (answered) {
              if (i === ex.correct)    { bg = 'rgba(45,107,58,0.1)'; border = '2px solid var(--salvia)'; }
              else if (i === selected) { bg = 'rgba(220,60,60,0.08)'; border = '2px solid #E05A5A'; }
            } else if (selected === i) { bg = 'var(--chip-bg)'; border = '2px solid var(--salvia)'; }
            return (
              <button key={i} onClick={() => handleSelect(i)} style={{
                padding: '15px 20px', borderRadius: '14px',
                background: bg, border, cursor: answered ? 'default' : 'pointer',
                textAlign: 'center', transition: 'all .15s', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              }}>
                <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '16px', color: 'var(--text)', fontWeight: 500, flex: 1, textAlign: 'center' }}>{opt}</span>
                {answered && i === ex.correct && <CheckIcon s={18} c="var(--salvia)" />}
              </button>
            );
          })}
        </div>

        {/* Lumi feedback card */}
        {answered && (
          <div style={{ marginTop: '14px', background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '16px', padding: '14px 16px', display: 'flex', alignItems: 'center', gap: '12px' }}>
            <LumiSVG size={52} />
            <div style={{ flex: 1 }}>
              <div style={{ fontFamily: 'Inter,sans-serif', fontWeight: 700, fontSize: '15px', color: isCorrect ? 'var(--salvia)' : '#E05A5A', marginBottom: '3px' }}>
                {isCorrect ? '¡Correcto! ✦' : 'Не совсем...'}
              </div>
              <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', lineHeight: 1.4 }}>
                {isCorrect ? `«${ex.question}» — это ${ex.options[ex.correct]}. Отличная работа!` : `Правильный ответ: «${ex.options[ex.correct]}».`}
              </p>
            </div>
          </div>
        )}
      </div>

      <div style={{ padding: '12px 16px 16px' }}>
        <PrimaryBtn onClick={onNext} disabled={!answered}>
          {answered ? 'Продолжить →' : 'Проверить'}
        </PrimaryBtn>
      </div>
    </div>
  );
}

// ─── Letter Tiles Exercise ────────────────────────────────────────────────────
function TilesExercise({ ex, onNext }) {
  const word = ex.word;

  // Build shuffled tiles with ids
  const buildTiles = () => {
    const extra = ['o', 'e', 'a']; // decoys depending on word length
    const allChars = [...word.split(''), ...extra.slice(0, Math.max(0, 11 - word.length))];
    // Shuffle
    for (let i = allChars.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [allChars[i], allChars[j]] = [allChars[j], allChars[i]];
    }
    return allChars.map((ch, i) => ({ id: i, ch, used: false }));
  };

  const [tiles, setTiles] = useState(() => buildTiles());
  const [answer, setAnswer] = useState([]); // [{ch, tileId}]
  const [checked, setChecked] = useState(false);

  const isCorrect = checked && answer.map(a => a.ch).join('') === word;
  const isWrong   = checked && !isCorrect;

  const addTile = (tile) => {
    if (tile.used || checked) return;
    setTiles(prev => prev.map(t => t.id === tile.id ? { ...t, used: true } : t));
    setAnswer(prev => [...prev, { ch: tile.ch, tileId: tile.id }]);
  };

  const removeSlot = (idx) => {
    if (checked) return;
    const removed = answer[idx];
    setTiles(prev => prev.map(t => t.id === removed.tileId ? { ...t, used: false } : t));
    setAnswer(prev => prev.filter((_, i) => i !== idx));
  };

  const handleCheck = () => setChecked(true);

  const handleReset = () => {
    setTiles(buildTiles());
    setAnswer([]);
    setChecked(false);
  };

  const slotCount = word.length;
  const rows = tiles.length <= 6 ? [tiles] : [tiles.slice(0, Math.ceil(tiles.length / 2)), tiles.slice(Math.ceil(tiles.length / 2))];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* City illustration bg area */}
      <div style={{ position: 'relative', overflow: 'hidden', background: 'var(--chip-bg)', flexShrink: 0, height: 200 }}>
        <img src="uploads/img1.png" alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block', opacity: 0.8 }}
          onError={e => { e.target.style.display = 'none'; }} />
        {/* Lumi + bubble */}
        <div style={{ position: 'absolute', top: 12, right: 16, display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
          <SpeechBubble text="Я верю
в тебя!" style={{ fontSize: '12px' }} />
          <LumiSVG size={64} pose="pencil" />
        </div>
        {/* Title overlay */}
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, padding: '30px 20px 14px', background: 'linear-gradient(transparent, var(--bg))' }}>
          <h2 style={{ margin: 0, fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '26px', color: 'var(--text)', textAlign: 'center' }}>Собери слово</h2>
          <div style={{ textAlign: 'center', fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--subtext)', marginTop: '2px' }}>{ex.word}</div>
        </div>
      </div>

      <div style={{ flex: 1, overflowY: 'auto', padding: '12px 16px 0' }}>
        {/* Audio pill */}
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '14px' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', padding: '8px 16px', borderRadius: '24px', background: 'var(--card-bg)', border: '1px solid var(--border)' }}>
            <button style={{ width: 28, height: 28, borderRadius: '50%', background: 'var(--salvia)', border: 'none', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
              <SoundIcon s={14} c="white" />
            </button>
            <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--text)', fontWeight: 500 }}>el {ex.word} – {ex.word}</span>
          </div>
        </div>

        {/* Answer slots */}
        <div style={{ display: 'flex', gap: '6px', justifyContent: 'center', flexWrap: 'wrap', marginBottom: '10px' }}>
          {Array.from({ length: slotCount }).map((_, i) => {
            const filled = answer[i];
            let borderColor = 'var(--border)';
            if (checked && filled) borderColor = isCorrect ? 'var(--salvia)' : '#E05A5A';
            else if (i === answer.length && !checked) borderColor = 'var(--salvia)';
            return (
              <button key={i} onClick={() => filled && removeSlot(i)} style={{
                width: Math.min(38, 280 / slotCount), height: 42,
                borderRadius: '10px', border: `2px solid ${borderColor}`,
                background: filled
                  ? (checked ? (isCorrect ? 'rgba(106,143,99,0.12)' : 'rgba(220,60,60,0.1)') : 'var(--chip-bg)')
                  : 'transparent',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                cursor: filled && !checked ? 'pointer' : 'default',
                transition: 'all .15s',
              }}>
                {filled && (
                  <span style={{
                    fontFamily: 'Lora,serif', fontWeight: 700,
                    fontSize: slotCount > 9 ? '13px' : '16px',
                    color: checked ? (isCorrect ? 'var(--salvia)' : '#E05A5A') : 'var(--text)',
                  }}>{filled.ch}</span>
                )}
              </button>
            );
          })}
        </div>
        <div style={{ textAlign: 'center', marginBottom: '16px' }}>
          <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '12px', color: 'var(--subtext)', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '5px' }}>
            <span>🌿</span> Forma la palabra seleccionando las letras
          </span>
        </div>

        {/* Feedback */}
        {checked && (
          <FeedbackBanner
            correct={isCorrect}
            message={isCorrect ? '¡Perfecto! Has escrito bien la palabra. ✨' : `La respuesta correcta es "${word}".`}
            points={isCorrect ? 10 : null}
          />
        )}

        {/* Letter tiles grid */}
        {!checked && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '4px' }}>
            {rows.map((row, ri) => (
              <div key={ri} style={{ display: 'flex', gap: '8px', justifyContent: 'center', flexWrap: 'wrap' }}>
                {row.map(tile => (
                  <button key={tile.id} onClick={() => addTile(tile)} disabled={tile.used} style={{
                    width: 44, height: 52, borderRadius: '12px',
                    border: '1.5px solid var(--border)',
                    background: tile.used ? 'transparent' : 'var(--card-bg)',
                    borderColor: tile.used ? 'transparent' : 'var(--border)',
                    cursor: tile.used ? 'default' : 'pointer',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    transition: 'all .1s',
                    boxShadow: tile.used ? 'none' : '0 2px 6px rgba(0,0,0,0.1)',
                  }}>
                    {!tile.used && (
                      <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '20px', color: 'var(--text)' }}>
                        {tile.ch}
                      </span>
                    )}
                  </button>
                ))}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Buttons */}
      <div style={{ padding: '12px 16px 16px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
        {!checked && (
          <button onClick={handleReset} style={{
            padding: '13px', borderRadius: '24px',
            border: '1.5px solid var(--border)', background: 'var(--card-bg)',
            fontFamily: 'Inter,sans-serif', fontWeight: 600, fontSize: '14px',
            color: 'var(--text)', cursor: 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px',
          }}>
            💡 Подсказка 🪙 1
          </button>
        )}
        <PrimaryBtn onClick={checked ? onNext : handleCheck} disabled={!checked && answer.length < slotCount}>
          {checked ? 'Продолжить →' : 'Проверить'}
        </PrimaryBtn>
      </div>
    </div>
  );
}

// ─── Write / Translation Exercise ────────────────────────────────────────────
function WriteExercise({ ex, onNext }) {
  const [value, setValue] = useState('');
  const [checked, setChecked] = useState(false);
  const inputRef = useRef(null);

  const normalise = s => s.trim().toLowerCase().replace(/[¡!¿?.,]/g, '');
  const isCorrect = checked && normalise(value) === normalise(ex.correct);

  const handleCheck = () => { if (value.trim()) setChecked(true); };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ flex: 1, overflowY: 'auto', padding: '0 16px' }}>
        {/* Lumi + header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
          <div style={{ flex: 1 }}>
            <h2 style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '24px', color: 'var(--text)', margin: '0 0 4px', display: 'flex', alignItems: 'center', gap: '6px' }}>
              Traduce al español <span>🌿</span>
            </h2>
            <p style={{ margin: 0, fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)' }}>
              Escribe la palabra completa.
            </p>
          </div>
          <LumiSVG size={68} pose="pencil" />
        </div>

        {/* Source word card */}
        <div style={{
          background: 'var(--card-bg)', border: '1px solid var(--border)',
          borderRadius: '18px', padding: '20px 16px', marginBottom: '16px',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <span style={{ fontSize: '28px' }}>🧳</span>
            <span style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '26px', color: 'var(--text)' }}>{ex.word_en}</span>
          </div>
          <div style={{ display: 'flex', gap: '5px', marginTop: '10px' }}>
            {Array.from({ length: ex.correct.length }).map((_, i) => (
              <div key={i} style={{ flex: 1, height: '3px', borderRadius: '2px', background: 'var(--progress-track)' }} />
            ))}
          </div>
        </div>

        {/* Input */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: '10px',
          border: `2px solid ${checked ? (isCorrect ? 'var(--salvia)' : '#E05A5A') : 'var(--salvia)'}`,
          borderRadius: '14px', padding: '14px 16px',
          background: 'var(--input-bg)', marginBottom: '10px',
          transition: 'border-color .2s',
        }}>
          <span style={{ fontSize: '16px' }}>🌿</span>
          <input
            ref={inputRef}
            value={value}
            onChange={e => !checked && setValue(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !checked && handleCheck()}
            placeholder="Escribe tu respuesta..."
            style={{
              flex: 1, border: 'none', background: 'transparent', outline: 'none',
              fontFamily: 'Inter,sans-serif', fontSize: '16px', color: 'var(--text)',
            }}
          />
        </div>

        {/* Hint */}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '6px', marginBottom: '14px' }}>
          <span style={{ fontSize: '14px' }}>💡</span>
          <span style={{ fontFamily: 'Inter,sans-serif', fontSize: '13px', color: 'var(--subtext)', lineHeight: 1.5 }}>
            <strong>Pista:</strong> {ex.hint}
          </span>
        </div>

        {/* Feedback */}
        {checked && (
          <FeedbackBanner
            correct={isCorrect}
            message={isCorrect ? '¡Perfecto! La traducción es correcta. ✨' : `La respuesta correcta es "${ex.correct}".`}
            points={isCorrect ? 10 : null}
          />
        )}
      </div>

      {/* Button */}
      <div style={{ padding: '12px 16px 16px' }}>
        <PrimaryBtn onClick={checked ? onNext : handleCheck} disabled={!checked && !value.trim()}>
          {checked ? 'Continuar →' : 'Comprobar'}
        </PrimaryBtn>
      </div>
    </div>
  );
}

// ─── Exercise Shell ───────────────────────────────────────────────────────────
function ExerciseScreen({ startIndex = 0, go }) {
  const { exercises, user } = window.LINGLOW_DATA;
  const [idx, setIdx] = useState(startIndex);
  const [done, setDone] = useState(false);

  const ex = exercises[idx % exercises.length];

  const handleNext = () => {
    if (idx + 1 >= exercises.length) setDone(true);
    else setIdx(i => i + 1);
  };

  if (done) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', padding: '32px', gap: '16px' }}>
        <LumiSVG size={100} pose="wave" />
        <div style={{ textAlign: 'center' }}>
          <h2 style={{ fontFamily: 'Lora,serif', fontWeight: 700, fontSize: '28px', color: 'var(--text)', margin: '0 0 8px' }}>
            ¡Ejercicio completado! ✦
          </h2>
          <p style={{ fontFamily: 'Inter,sans-serif', fontSize: '15px', color: 'var(--subtext)', margin: 0 }}>
            Cada práctica te acerca más al español.
          </p>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <LevelChip label="🔥 +1 racha" active />
          <LevelChip label="⭐ +30 puntos" active />
        </div>
        <PrimaryBtn onClick={() => go('grammar')} style={{ marginTop: '8px' }}>
          Volver a lecciones
        </PrimaryBtn>
        <button onClick={() => { setIdx(0); setDone(false); }} style={{
          background: 'none', border: 'none', cursor: 'pointer',
          fontFamily: 'Inter,sans-serif', fontSize: '14px', color: 'var(--salvia)', fontWeight: 600,
        }}>
          Practicar de nuevo
        </button>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <ExerciseHeader qNum={idx + 1} total={exercises.length * 2} onClose={() => go('grammar')} streak={user.streak} />
      <div style={{ flex: 1, overflow: 'hidden' }}>
        {ex.type === 'choice' && <ChoiceExercise key={idx} ex={ex} onNext={handleNext} />}
        {ex.type === 'tiles'  && <TilesExercise  key={idx} ex={ex} onNext={handleNext} />}
        {ex.type === 'write'  && <WriteExercise  key={idx} ex={ex} onNext={handleNext} />}
      </div>
    </div>
  );
}

Object.assign(window, { ExerciseScreen });
