// Linglow — App Shell with routing + theme + responsive layout
const { useState, useEffect } = React;

function App() {
  const saved = (() => { try { return localStorage.getItem('linglow_theme') || 'light'; } catch { return 'light'; } })();
  const [theme, setTheme] = useState(saved);
  const [screen, setScreen] = useState(() => {
    try { return localStorage.getItem('linglow_screen') || 'home'; } catch { return 'home'; }
  });
  const [isDesktop, setIsDesktop] = useState(() => window.innerWidth >= 900);

  // Persist theme
  useEffect(() => {
    try { localStorage.setItem('linglow_theme', theme); } catch {}
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  // Persist screen
  useEffect(() => {
    try { localStorage.setItem('linglow_screen', screen); } catch {}
  }, [screen]);

  // Track viewport width
  useEffect(() => {
    const check = () => setIsDesktop(window.innerWidth >= 900);
    window.addEventListener('resize', check);
    return () => window.removeEventListener('resize', check);
  }, []);

  const go = (s) => setScreen(s);
  const toggleTheme = () => setTheme(t => t === 'light' ? 'dark' : 'light');

  const [screenId, param] = screen.split(':');
  // Exercise goes full-screen — no nav on either layout
  const showNav = !['exercise', 'chat'].includes(screenId);

  const renderScreen = () => {
    switch (screenId) {
      case 'home':     return <HomeScreen     go={go} theme={theme} onToggleTheme={toggleTheme} />;
      case 'map':      return <MapScreen       go={go} theme={theme} />;
      case 'district': return <DistrictScreen  go={go} theme={theme} districtId={param || 'viajes'} />;
      case 'practice': return <PracticeScreen  go={go} />;
      case 'grammar':  return <GrammarScreen   go={go} />;
      case 'lesson':   return <LessonScreen    go={go} lessonId={param || 'ar'} />;
      case 'exercise':     return <ExerciseScreen       go={go} startIndex={parseInt(param || '0', 10)} />;
      case 'reading':        return <ReadingListScreen    go={go} />;
      case 'reading-text':   return <ReadingTextScreen    go={go} textId={param || 'dialogo-estacion'} />;
      case 'chat':           return <ChatScreen           go={go} />;
      case 'progress': return <ProgressScreen  go={go} />;
      case 'profile':  return <ProfileScreen   go={go} theme={theme} onToggleTheme={toggleTheme} />;
      default:         return <HomeScreen      go={go} theme={theme} onToggleTheme={toggleTheme} />;
    }
  };

  return (
    <ThemeCtx.Provider value={{ theme, toggle: toggleTheme }}>
    <LayoutCtx.Provider value={{ isDesktop }}>
    <div style={{
      position: 'fixed', inset: 0,
      display: 'flex',
      background: isDesktop
        ? 'var(--bg)'
        : (theme === 'dark' ? '#000' : '#d8d0bc'),
    }}>
      {isDesktop ? (
        // ── Desktop layout: sidebar + content ──────────────────
        <>
          {showNav && (
            <SideNav
              active={screenId}
              go={go}
              theme={theme}
              onToggleTheme={toggleTheme}
            />
          )}
          {/* Content area — centered, max-width except for map */}
          <div style={{
            flex: 1, height: '100%', overflow: 'hidden',
            display: 'flex', justifyContent: 'center',
            position: 'relative',
          }}>
            <div style={{
              flex: 1, height: '100%', overflow: 'hidden',
              maxWidth: screenId === 'map' ? '760px' : '880px',
              width: '100%', position: 'relative',
            }}>
              {renderScreen()}
            </div>
          </div>
        </>
      ) : (
        // ── Mobile layout: full-width ────────────────────────────
        <div style={{
          width: '100%', height: '100%',
          background: 'var(--bg)', position: 'relative',
          overflow: 'hidden',
        }}>
          <div style={{ height: '100%', position: 'relative' }}>
            {renderScreen()}
          </div>
          {showNav && <BottomNav active={screenId} go={go} />}
        </div>
      )}
    </div>
    </LayoutCtx.Provider>
    </ThemeCtx.Provider>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
