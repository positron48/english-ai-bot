const API = '';

let courses = [];
let currentDraftId = null;

async function api(path, opts = {}) {
  const res = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
    ...opts,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = new Error(data.error || res.statusText);
    if (data.log) err.log = data.log;
    throw err;
  }
  return data;
}

function toast(msg) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.classList.remove('hidden');
  setTimeout(() => el.classList.add('hidden'), 4000);
}

function fillCourseSelects() {
  const savedPub = localStorage.getItem('reading-cms-pub-course') || 'es_ru';
  const ids = ['gen-course', 'import-course', 'filter-course', 'pub-course', 'cover-batch-course'];
  for (const id of ids) {
    const sel = document.getElementById(id);
    const keepAll = id === 'filter-course';
    sel.innerHTML = keepAll ? '<option value="">All courses</option>' : '';
    for (const c of courses) {
      const opt = document.createElement('option');
      opt.value = c.code;
      opt.textContent = `${c.title} (${c.code})`;
      sel.appendChild(opt);
    }
    if (!keepAll && savedPub && [...sel.options].some(o => o.value === savedPub)) {
      if (id === 'pub-course' || id === 'cover-batch-course' || id === 'gen-course' || id === 'import-course') {
        sel.value = savedPub;
      }
    }
  }
}

function badgeAudio(st) {
  const cls = st === 'ready' ? 'ready' : st === 'partial' ? 'partial' : 'none';
  return `<span class="badge ${cls}">${st || 'none'}</span>`;
}

function badgeCover(st) {
  const cls = st === 'ready' ? 'ready' : st === 'prompt' ? 'prompt' : 'none';
  return `<span class="badge ${cls}">${st || 'none'}</span>`;
}

function badgeGit(item) {
  const st = item?.git_status || '';
  if (!st) return '';
  if (item.is_new_uncommitted) {
    return ' <span class="badge new-uncommitted">new</span>';
  }
  return ` <span class="badge git-dirty">${escapeHtml(st)}</span>`;
}

function formatCoverDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function parseCoverDateMs(item) {
  const t = item?.cover_generated_at;
  if (!t) return null;
  const ms = Date.parse(t);
  return Number.isNaN(ms) ? null : ms;
}

function compareCoverDate(a, b, newestFirst) {
  const ams = parseCoverDateMs(a);
  const bms = parseCoverDateMs(b);
  if (ams == null && bms == null) return 0;
  if (ams == null) return 1;
  if (bms == null) return -1;
  return newestFirst ? bms - ams : ams - bms;
}

function sortPublishedTexts(texts, sortKey) {
  const items = [...texts];
  const byLevelTitle = (a, b) => (a.level || '').localeCompare(b.level || '') || (a.title || '').localeCompare(b.title || '');
  switch (sortKey) {
    case 'cover_desc':
      return items.sort((a, b) => compareCoverDate(a, b, true) || byLevelTitle(a, b));
    case 'cover_asc':
      return items.sort((a, b) => compareCoverDate(a, b, false) || byLevelTitle(a, b));
    default:
      return items.sort(byLevelTitle);
  }
}

function coverThumbCell(source, textId, courseCode, thumbPath, coverStatus) {
  if (coverStatus === 'ready' && thumbPath) {
    const file = thumbPath.split('/').pop();
    const url = source === 'course'
      ? `/api/course-images/${encodeURIComponent(courseCode)}/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`
      : `/api/images/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`;
    return `<img class="table-thumb" src="${url}?t=${Date.now()}" alt="" loading="lazy" />`;
  }
  if (coverStatus === 'prompt') {
    const kind = source === 'draft' ? 'draft' : 'pub';
    return `<button type="button" class="table-thumb-placeholder" data-kind="${kind}" data-id="${escapeHtml(textId)}" data-course="${escapeHtml(courseCode)}" title="Промпт готов — открыть и сгенерировать картинку">⌁</button>`;
  }
  return '<span class="meta">—</span>';
}

function rowActionsHtml(kind, item) {
  const id = item.text_id;
  const course = item.course_code || document.getElementById(kind === 'pub' ? 'pub-course' : 'filter-course')?.value || '';
  const coverReady = item.cover_status === 'ready';
  const coverPrompt = item.cover_status === 'prompt';
  const coverViewEnabled = coverReady || coverPrompt;
  const coverLabel = coverReady ? 'Перегенерировать' : coverPrompt ? 'Картинка (ComfyUI)' : 'Сгенерировать';
  const audioBtn = kind === 'draft'
    ? `<button type="button" class="btn btn-sm row-audio" data-id="${id}">TTS</button>`
    : '';
  return `<div class="row-actions">
    ${audioBtn}
    <button type="button" class="btn btn-sm row-cover-view" data-kind="${kind}" data-id="${id}" data-course="${course}" ${coverViewEnabled ? '' : 'disabled'}>Обложка</button>
    <button type="button" class="btn btn-sm row-cover-gen" data-kind="${kind}" data-id="${id}" data-course="${course}" data-cover-ready="${coverReady}" data-cover-prompt="${coverPrompt}" data-cover-prompt-text="${encodeURIComponent(item.cover_image_prompt || '')}">${coverLabel}</button>
    ${coverReady ? `<button type="button" class="btn btn-sm danger row-cover-del" data-kind="${kind}" data-id="${id}" data-course="${course}">Удалить обложку</button>` : ''}
    ${kind === 'pub' ? `<button type="button" class="btn btn-sm danger pub-del" data-id="${id}">Удалить</button>` : ''}
  </div>`;
}

function renderDraftsTable(drafts) {
  const wrap = document.getElementById('drafts-table');
  if (!drafts.length) {
    wrap.innerHTML = '<p class="meta">Черновиков нет. Создайте через «+ Новый текст» или «→ В черновики» на вкладке каталога.</p>';
    return;
  }
  wrap.innerHTML = `<table class="texts-table">
    <thead><tr>
      <th>Title</th><th>Course</th><th>Level</th><th>Status</th>
      <th>Audio</th><th>Cover</th><th>Preview</th><th>Actions</th>
    </tr></thead>
    <tbody>${drafts.map(d => `<tr data-id="${d.text_id}" data-course="${escapeHtml(d.course_code)}">
      <td class="title-cell">${escapeHtml(d.title)}</td>
      <td>${d.course_code}</td>
      <td>${d.level}</td>
      <td>${d.status}</td>
      <td>${badgeAudio(d.audio_status)} <span class="meta">${d.segments_with_audio}/${d.segments_total}</span></td>
      <td>${badgeCover(d.cover_status)}</td>
      <td>${coverThumbCell('draft', d.text_id, d.course_code, d.cover_thumb_rel_path, d.cover_status)}</td>
      <td>${rowActionsHtml('draft', d)}</td>
    </tr>`).join('')}</tbody></table>`;
  bindTableRowHandlers(wrap, 'draft');
}

function renderPublishedTable(texts, total) {
  const wrap = document.getElementById('published-table');
  const countEl = document.getElementById('published-count');
  const course = document.getElementById('pub-course').value || courses[0]?.code || '';
  if (countEl) countEl.textContent = total != null ? `(${total})` : texts.length ? `(${texts.length})` : '';
  if (!texts.length) {
    wrap.innerHTML = `<p class="meta">Нет текстов в <code>courses/${course === 'es_ru' ? 'spanish' : 'english'}-grammar/reading/</code> для выбранных фильтров. Проверьте course и submodule, или снимите фильтры уровня, картинки и git-статуса.</p>`;
    return;
  }
  const sortKey = document.getElementById('pub-sort')?.value || 'level';
  const sorted = sortPublishedTexts(texts, sortKey);
  wrap.innerHTML = `<table class="texts-table">
    <thead><tr>
      <th>Title</th><th>Level</th><th>Audio</th><th>Cover</th><th>Preview</th><th>Seg</th><th>Actions</th>
    </tr></thead>
    <tbody>${sorted.map(t => {
      const coverDate = t.cover_generated_at ? `<div class="meta cover-date">${formatCoverDate(t.cover_generated_at)}</div>` : '';
      return `<tr data-id="${t.text_id}" data-course="${escapeHtml(t.course_code || course)}">
      <td class="title-cell">${escapeHtml(t.title)}${badgeGit(t)}${t.in_cms ? ' <span class="badge partial">cms</span>' : ''}</td>
      <td>${t.level}</td>
      <td>${badgeAudio(t.audio_status)} <span class="meta">${t.segments_with_audio}/${t.segments_count}</span></td>
      <td>${badgeCover(t.cover_status)}${coverDate}</td>
      <td>${coverThumbCell('course', t.text_id, t.course_code || course, t.cover_thumb_rel_path, t.cover_status)}</td>
      <td>${t.segments_count}</td>
      <td>${rowActionsHtml('pub', { ...t, course_code: t.course_code || course })}</td>
    </tr>`;
    }).join('')}</tbody></table>`;
  bindTableRowHandlers(wrap, 'pub');
}

function bindTableRowHandlers(wrap, kind) {
  wrap.querySelectorAll('tbody tr').forEach(row => {
    row.addEventListener('click', (e) => {
      if (e.target.closest('.row-actions, .table-thumb, button, audio')) return;
      openTextReader(kind, row.dataset.id, row.dataset.course || '');
    });
  });
  wrap.querySelectorAll('.row-audio').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      const id = btn.dataset.id;
      toast('TTS...');
      try {
        await api(`/api/drafts/${encodeURIComponent(id)}/audio`, { method: 'POST' });
        toast('Audio OK');
        await loadDrafts();
      } catch (err) {
        toast(err.message);
      }
    });
  });
  wrap.querySelectorAll('.row-cover-view').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      if (btn.disabled) return;
      openCoverModal(btn.dataset.kind, btn.dataset.id, btn.dataset.course);
    });
  });
  wrap.querySelectorAll('.row-cover-gen').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      const { kind, id, course, coverReady, coverPrompt } = btn.dataset;
      const regen = coverReady === 'true';
      const promptOnly = coverPrompt === 'true';
      if (promptOnly) {
        const prompt = decodeURIComponent(btn.dataset.coverPromptText || '');
        if (!prompt) {
          openCoverModal(kind, id, course);
          return;
        }
        if (!confirm('Сгенерировать картинку по сохранённому промпту? (ComfyUI)')) return;
        await generateCoverFromPrompt(kind, id, course, prompt);
        return;
      }
      const msg = regen
        ? 'Перегенерировать обложку? (LLM + ComfyUI)'
        : 'Сгенерировать обложку? (LLM + ComfyUI)';
      if (!confirm(msg)) return;
      await generateCover(kind, id, course, regen);
    });
  });
  wrap.querySelectorAll('.row-cover-del').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      const { kind, id, course } = btn.dataset;
      if (!confirm('Удалить обложку (файлы и поля в JSON)?')) return;
      await deleteCover(kind, id, course);
    });
  });
  wrap.querySelectorAll('.table-thumb').forEach(img => {
    img.addEventListener('click', (e) => {
      e.stopPropagation();
      const viewBtn = img.closest('tr')?.querySelector('.row-cover-view');
      if (viewBtn && !viewBtn.disabled) viewBtn.click();
    });
  });
  wrap.querySelectorAll('.table-thumb-placeholder').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      openCoverModal(btn.dataset.kind, btn.dataset.id, btn.dataset.course);
    });
  });
  if (kind === 'pub') {
    wrap.querySelectorAll('.pub-del').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        if (!confirm('Удалить из course/bundle?')) return;
        const scrollState = captureListScrollState('pub');
        const course = document.getElementById('pub-course').value;
        await api(`/api/published?course_code=${encodeURIComponent(course)}&text_id=${encodeURIComponent(btn.dataset.id)}`, { method: 'DELETE' });
        toast('Deleted from course');
        await reloadListPreservingScroll('pub', scrollState);
      });
    });
  }
}

async function deleteCover(kind, textId, courseCode) {
  const scrollState = captureListScrollState(kind);
  try {
    if (kind === 'pub' || kind === 'course') {
      await api('/api/published/cover', {
        method: 'DELETE',
        body: JSON.stringify({ course_code: courseCode, text_id: textId }),
      });
      toast('Обложка удалена');
      await reloadListPreservingScroll('pub', scrollState);
    } else {
      await api(`/api/drafts/${encodeURIComponent(textId)}/cover`, { method: 'DELETE' });
      toast('Обложка удалена');
      await reloadListPreservingScroll('draft', scrollState);
      if (currentDraftId === textId) await openDraft(textId);
    }
  } catch (err) {
    toast(err.message);
  }
}

async function generateCover(kind, textId, courseCode, force) {
  const listKind = kind === 'draft' ? 'draft' : 'pub';
  const scrollState = captureListScrollState(listKind);
  const ui = openCoverProgressDialog(textId);
  const poll = startCoverProgressPoll(courseCode, textId, ui);
  try {
    let data;
    if (kind === 'pub') {
      data = await api('/api/published/cover', {
        method: 'POST',
        body: JSON.stringify({ course_code: courseCode, text_id: textId, force }),
      });
      await finishCoverProgressDialog(ui, data.log, false, '', data, courseCode, textId);
      toast(`Cover: ${data.text?.cover_status || 'ok'}`);
      await reloadListPreservingScroll('pub', scrollState);
      if (data.text?.cover_status === 'ready') {
        openCoverModal('pub', textId, courseCode, data.text);
      }
    } else {
      data = await api(`/api/drafts/${encodeURIComponent(textId)}/cover`, {
        method: 'POST',
        body: JSON.stringify({ force }),
      });
      await finishCoverProgressDialog(ui, data.log || data.draft?.last_job_log, false, '', data, courseCode, textId);
      toast(`Cover: ${data.draft?.cover_status || 'ok'}`);
      await reloadListPreservingScroll('draft', scrollState);
      if (data.draft?.cover_status === 'ready') {
        const doc = (await api('/api/drafts/' + encodeURIComponent(textId))).document;
        openCoverModal('draft', textId, courseCode, { ...data.draft, ...pickCoverFields(doc) });
      }
    }
  } catch (err) {
    await finishCoverProgressDialog(ui, err.log || '', true, err.message, null, courseCode, textId);
    toast(err.message);
  } finally {
    clearInterval(poll);
    ui.closeBtn.disabled = false;
  }
}

async function generateCoverFromPrompt(kind, textId, courseCode, prompt) {
  const p = String(prompt || '').trim();
  if (!p) {
    toast('Введите промпт для картинки');
    return null;
  }
  const listKind = kind === 'draft' ? 'draft' : 'pub';
  const scrollState = captureListScrollState(listKind);
  const ui = openCoverProgressDialog(textId);
  const poll = startCoverProgressPoll(courseCode, textId, ui);
  try {
    const body = { skip_llm: true, force: true, prompt: p };
    let data;
    if (kind === 'pub') {
      data = await api('/api/published/cover', {
        method: 'POST',
        body: JSON.stringify({ ...body, course_code: courseCode, text_id: textId }),
      });
      await finishCoverProgressDialog(ui, data.log, false, '', data, courseCode, textId);
      toast(`Cover: ${data.text?.cover_status || 'ok'}`);
      await reloadListPreservingScroll('pub', scrollState);
      return data;
    }
    data = await api(`/api/drafts/${encodeURIComponent(textId)}/cover`, {
      method: 'POST',
      body: JSON.stringify(body),
    });
    await finishCoverProgressDialog(ui, data.log || data.draft?.last_job_log, false, '', data, courseCode, textId);
    toast(`Cover: ${data.draft?.cover_status || 'ok'}`);
    await reloadListPreservingScroll('draft', scrollState);
    return data;
  } catch (err) {
    await finishCoverProgressDialog(ui, err.log || '', true, err.message, null, courseCode, textId);
    toast(err.message);
    return null;
  } finally {
    clearInterval(poll);
    ui.closeBtn.disabled = false;
  }
}

function readingTextRu(doc) {
  const segs = doc?.reading_passage?.segments;
  if (!Array.isArray(segs)) return '';
  return segs.map(s => String(s?.text_translation_ru || '').trim()).filter(Boolean).join(' ');
}

async function loadCoverModalData(kind, textId, courseCode) {
  if (kind === 'draft') {
    const data = await api('/api/drafts/' + encodeURIComponent(textId));
    return {
      item: { ...data.draft, ...pickCoverFields(data.document) },
      doc: data.document,
    };
  }
  const course = courseCode || document.getElementById('pub-course').value;
  const params = new URLSearchParams({ course_code: course, text_id: textId });
  const data = await api('/api/published/detail?' + params.toString());
  return {
    item: { ...data.text, ...pickCoverFields(data.document) },
    doc: data.document,
  };
}

function openCoverProgressDialog(opts) {
  let textId = '';
  let batchMode = false;
  let batchId = '';
  let total = 0;
  if (typeof opts === 'string') {
    textId = opts;
  } else if (opts) {
    textId = opts.textId || '';
    batchMode = !!opts.batch;
    batchId = opts.batchId || '';
    total = Number(opts.total) || 0;
  }

  const dlg = document.getElementById('cover-progress-dialog');
  const titleEl = document.getElementById('cover-progress-dialog-title');
  const batchSection = document.getElementById('cover-batch-overall');
  const batchStatusEl = document.getElementById('cover-batch-status');
  const batchBarEl = document.getElementById('cover-batch-bar');
  const batchPercentEl = document.getElementById('cover-batch-percent');
  const currentSectionTitle = document.getElementById('cover-current-section-title');
  const logEl = document.getElementById('cover-progress-log');
  const statusEl = document.getElementById('cover-progress-status');
  const barEl = document.getElementById('cover-progress-bar');
  const percentEl = document.getElementById('cover-progress-percent');
  const stagesEl = document.getElementById('cover-progress-stages');
  const closeBtn = document.getElementById('cover-progress-close');

  if (batchMode) {
    titleEl.textContent = 'Batch: генерация обложек';
    batchSection.classList.remove('hidden');
    currentSectionTitle.classList.remove('hidden');
    batchBarEl.style.width = '0%';
    batchPercentEl.textContent = '0%';
    batchStatusEl.textContent = total > 0 ? `0 / ${total} текстов` : 'Планирование…';
  } else {
    titleEl.textContent = 'Генерация обложки';
    batchSection.classList.add('hidden');
    currentSectionTitle.classList.add('hidden');
  }

  logEl.innerHTML = '';
  barEl.style.width = '0%';
  percentEl.textContent = '0%';
  const defaultStages = defaultCoverStagesList();
  const ui = {
    dlg, logEl, statusEl, barEl, percentEl, stagesEl, closeBtn,
    batchMode, batchId, batchStatusEl, batchBarEl, batchPercentEl,
    lastCurrentTextId: '',
    lastLog: '',
    logLineCount: 0,
    lastStagesKey: '',
    lastPercent: -1,
    lastBatchPercent: -1,
    lastStatus: '',
    lastBatchStatus: '',
  };
  renderCoverProgress(ui, {
    percent: 0,
    stage_label: batchMode ? 'Ожидание…' : 'Подготовка…',
    stages: defaultStages,
    running: true,
  });
  closeBtn.disabled = true;
  closeBtn.onclick = () => dlg.close();
  dlg.showModal();
  return ui;
}

function batchPhaseLabel(phase) {
  switch (phase) {
    case 'prompts': return 'Фаза 1 — промпты (LLM)';
    case 'stopping_llm': return 'Остановка llama.cpp…';
    case 'images': return 'Фаза 2 — картинки (ComfyUI)';
    default: return '';
  }
}

function renderCoverBatchProgress(ui, b) {
  if (!ui?.batchMode || !b) return;
  const pct = Number.isFinite(b.percent) ? b.percent : 0;
  if (pct !== ui.lastBatchPercent) {
    ui.lastBatchPercent = pct;
    ui.batchBarEl.style.width = `${pct}%`;
    ui.batchPercentEl.textContent = `${pct}%`;
  }
  const total = b.total || 0;
  const cur = b.current || 0;
  const phaseLabel = batchPhaseLabel(b.phase);
  let batchStatus = '';
  if (b.done) {
    if (b.error) {
      batchStatus = `Завершено с ошибками: ${b.completed || 0} ok, ${b.skipped || 0} skip, ${b.failed || 0} fail из ${total}`;
    } else {
      batchStatus = `Готово: ${b.completed || 0} ok, ${b.skipped || 0} skip, ${b.failed || 0} fail из ${total}`;
    }
  } else if (b.phase === 'stopping_llm') {
    batchStatus = phaseLabel;
  } else if (b.current_text_id) {
    const title = (b.current_text_title || '').trim();
    batchStatus = `${phaseLabel}: текст ${cur} из ${total} — ${b.current_text_id}${title ? ` — ${title}` : ''}`;
  } else if (phaseLabel) {
    batchStatus = phaseLabel;
  } else {
    batchStatus = total > 0 ? `Текст ${cur} из ${total}` : 'Запуск batch…';
  }
  if (batchStatus && batchStatus !== ui.lastBatchStatus) {
    ui.lastBatchStatus = batchStatus;
    ui.batchStatusEl.textContent = batchStatus;
  }
  if (b.log !== undefined) syncCoverProgressLog(ui, b.log);
  if (b.current_text_id && b.current_text_id !== ui.lastCurrentTextId) {
    ui.lastCurrentTextId = b.current_text_id;
    ui.lastPercent = -1;
    ui.lastStagesKey = '';
    ui.barEl.style.width = '0%';
    ui.percentEl.textContent = '0%';
  }
  const curP = b.current_progress || {};
  if (!curP.stage_label && b.running && b.current_text_id) {
    curP.stage_label = 'Подготовка…';
    curP.running = true;
  }
  renderCoverProgress(ui, curP);
}

function finishCoverBatchProgressDialog(ui, b, isError, errMsg) {
  renderCoverBatchProgress(ui, {
    ...b,
    done: true,
    running: false,
    percent: 100,
    error: isError ? (errMsg || b?.error || 'неизвестная') : '',
  });
  if (isError || b?.error) {
    ui.statusEl.textContent = `Ошибка: ${errMsg || b?.error || 'неизвестная'}`;
    ui.lastStatus = ui.statusEl.textContent;
  } else {
    ui.statusEl.textContent = 'Готово';
    ui.lastStatus = 'Готово';
    ui.barEl.style.width = '100%';
    ui.percentEl.textContent = '100%';
  }
}

function startCoverBatchProgressPoll(batchId, ui, onDone) {
  return setInterval(async () => {
    try {
      const res = await fetch(
        API + `/api/cover-batch-progress?batch_id=${encodeURIComponent(batchId)}`,
      );
      const b = await res.json().catch(() => ({}));
      renderCoverBatchProgress(ui, b);
      if (b.done && typeof onDone === 'function') {
        onDone(b);
      }
    } catch (_) {
      /* ignore poll errors */
    }
  }, 1000);
}

function classifyCoverLogLine(line) {
  const comfyIdx = line.indexOf('COMFYUI_PROMPT|');
  if (comfyIdx >= 0) {
    return { kind: 'comfy-prompt', text: line.slice(comfyIdx + 'COMFYUI_PROMPT|'.length) };
  }
  if (/▶ ComfyUI prompt/i.test(line)) {
    return { kind: 'comfy-header', text: line.replace(/^\[reading-cover[^\]]*\]\s*/, '') };
  }
  if (/prompt ready \(\d+ chars\):/i.test(line)) {
    return { kind: 'comfy-legacy', text: line };
  }
  return { kind: 'normal', text: line };
}

function appendCoverLogLine(container, line) {
  if (line === '') return;
  const row = document.createElement('div');
  row.className = 'log-line';
  const classified = classifyCoverLogLine(line);
  if (classified.kind === 'comfy-header') {
    row.classList.add('log-line--comfy-header');
    row.textContent = classified.text;
  } else if (classified.kind === 'comfy-prompt') {
    row.classList.add('log-line--comfy-prompt');
    const tag = document.createElement('span');
    tag.className = 'log-line-tag';
    tag.textContent = 'ComfyUI prompt';
    const body = document.createElement('span');
    body.className = 'log-line-prompt';
    body.textContent = classified.text;
    row.appendChild(tag);
    row.appendChild(body);
  } else if (classified.kind === 'comfy-legacy') {
    row.classList.add('log-line--comfy-prompt');
    row.textContent = classified.text;
  } else {
    row.textContent = line;
  }
  container.appendChild(row);
}

function syncCoverProgressLog(ui, fullLog) {
  if (!ui?.logEl || fullLog == null) return;
  if (fullLog === ui.lastLog) return;

  const lines = fullLog.split('\n');
  if (lines.length < ui.logLineCount) {
    ui.logEl.innerHTML = '';
    ui.logLineCount = 0;
  }

  const wrap = ui.logEl.parentElement;
  const wasAtBottom = wrap
    && (wrap.scrollHeight - wrap.scrollTop - wrap.clientHeight < 48);

  for (let i = ui.logLineCount; i < lines.length; i++) {
    appendCoverLogLine(ui.logEl, lines[i]);
  }
  ui.logLineCount = lines.length;
  ui.lastLog = fullLog;

  if (wasAtBottom && wrap) {
    wrap.scrollTop = wrap.scrollHeight;
  }
}

function renderCoverProgress(ui, p) {
  if (!ui || !p) return;
  if (p.log !== undefined) syncCoverProgressLog(ui, p.log);

  const pct = Number.isFinite(p.percent) ? p.percent : 0;
  if (pct !== ui.lastPercent) {
    ui.lastPercent = pct;
    ui.barEl.style.width = `${pct}%`;
    ui.percentEl.textContent = `${pct}%`;
  }

  let statusText = '';
  if (p.stage_label) {
    if (p.error) statusText = `Ошибка: ${p.error}`;
    else if (p.done) statusText = 'Готово';
    else if (p.running) statusText = p.stage_label;
  }
  if (statusText && statusText !== ui.lastStatus) {
    ui.lastStatus = statusText;
    ui.statusEl.textContent = statusText;
  }

  if (Array.isArray(p.stages) && p.stages.length) {
    const key = JSON.stringify(p.stages);
    if (key !== ui.lastStagesKey) {
      ui.lastStagesKey = key;
      ui.stagesEl.innerHTML = p.stages.map(st => `
        <li class="cover-stage cover-stage--${escapeHtml(st.status || 'pending')}">
          <span class="cover-stage-icon" aria-hidden="true"></span>
          <span class="cover-stage-label">${escapeHtml(st.label || st.id || '')}</span>
        </li>
      `).join('');
    }
  }
}

async function fetchCoverProgress(courseCode, textId) {
  try {
    const res = await fetch(
      `${API}/api/cover-progress?course_code=${encodeURIComponent(courseCode)}&text_id=${encodeURIComponent(textId)}`,
    );
    return await res.json().catch(() => ({}));
  } catch (_) {
    return {};
  }
}

function defaultCoverStagesList() {
  return [
    { id: 'prepare', label: 'Подготовка', status: 'pending' },
    { id: 'llm', label: 'Промпт (LLM)', status: 'pending' },
    { id: 'comfyui', label: 'Картинка (ComfyUI)', status: 'pending' },
    { id: 'resize', label: 'WebP thumb + hero', status: 'pending' },
    { id: 'save', label: 'Сохранение', status: 'pending' },
  ];
}

function markAllCoverStagesDone(stages) {
  const base = Array.isArray(stages) && stages.length ? stages : defaultCoverStagesList();
  return base.map(st => ({
    ...st,
    status: st.status === 'error' ? 'error' : 'done',
  }));
}

async function finishCoverProgressDialog(ui, log, isError, errMsg, progress, courseCode, textId) {
  let payload = progress && typeof progress === 'object' ? { ...progress } : {};
  if (!isError && courseCode && textId) {
    const live = await fetchCoverProgress(courseCode, textId);
    if (live && typeof live === 'object') {
      payload = { ...live, ...payload };
    }
  }
  if (log) payload.log = log;
  if (isError) {
    payload.error = errMsg || payload.error || 'неизвестная';
    payload.done = true;
    payload.running = false;
    ui.lastStatus = '';
  } else {
    payload.done = true;
    payload.running = false;
    payload.percent = 100;
    payload.stage_label = 'Готово';
    payload.stages = markAllCoverStagesDone(payload.stages);
  }
  ui.lastStagesKey = '';
  renderCoverProgress(ui, payload);
  if (isError) {
    ui.statusEl.textContent = `Ошибка: ${errMsg || 'неизвестная'}`;
    ui.lastStatus = ui.statusEl.textContent;
  }
}

function startCoverProgressPoll(courseCode, textId, ui) {
  return setInterval(async () => {
    try {
      const res = await fetch(
        API + `/api/cover-progress?course_code=${encodeURIComponent(courseCode)}&text_id=${encodeURIComponent(textId)}`,
      );
      const p = await res.json().catch(() => ({}));
      renderCoverProgress(ui, p);
    } catch (_) {
      /* ignore poll errors */
    }
  }, 1000);
}

function pickCoverFields(doc) {
  return {
    cover_thumb_rel_path: doc?.cover_thumb_rel_path,
    cover_hero_rel_path: doc?.cover_hero_rel_path,
    cover_image_prompt: doc?.cover_image_prompt,
    cover_status: 'ready',
  };
}

function coverImageUrl(kind, textId, courseCode, relPath) {
  if (!relPath) return '';
  const file = relPath.split('/').pop();
  if (kind === 'course' || kind === 'pub') {
    return `/api/course-images/${encodeURIComponent(courseCode)}/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`;
  }
  return `/api/images/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`;
}

function openCoverModal(kind, textId, courseCode, prefetched) {
  const dlg = document.getElementById('cover-dialog');
  const titleEl = document.getElementById('cover-dialog-title');
  const bodyEl = document.getElementById('cover-dialog-body');
  titleEl.textContent = textId;
  dlg.showModal();

  const render = (item, doc) => {
    const thumb = item.cover_thumb_rel_path || '';
    const hero = item.cover_hero_rel_path || '';
    const thumbUrl = coverImageUrl(kind === 'draft' ? 'draft' : 'course', textId, courseCode, thumb);
    const heroUrl = coverImageUrl(kind === 'draft' ? 'draft' : 'course', textId, courseCode, hero);
    const ts = Date.now();
    const promptVal = item.cover_image_prompt || '';
    const ruText = readingTextRu(doc);
    bodyEl.innerHTML = `
      <div class="cover-modal-status">Статус: ${escapeHtml(item.cover_status || 'none')}</div>
      <div class="cover-modal-images">
        ${thumbUrl ? `<figure><figcaption>Thumb</figcaption><img src="${thumbUrl}?t=${ts}" alt="thumb" /></figure>` : ''}
        ${heroUrl ? `<figure><figcaption>Hero</figcaption><img class="hero" src="${heroUrl}?t=${ts}" alt="hero" /></figure>` : ''}
        ${!thumbUrl && !heroUrl && promptVal ? '<p class="meta cover-modal-no-image">Картинки ещё нет — промпт сохранён, можно сгенерировать через ComfyUI.</p>' : ''}
      </div>
      ${ruText ? `<div class="cover-modal-ru">${escapeHtml(ruText)}</div>` : ''}
      <label class="cover-modal-prompt-label" for="cover-modal-prompt">Промпт для картинки</label>
      <textarea id="cover-modal-prompt" class="cover-modal-prompt-input" rows="4">${escapeHtml(promptVal)}</textarea>
      <div class="cover-modal-actions">
        <button type="button" class="btn primary" id="cover-modal-comfy">Сгенерировать (ComfyUI)</button>
        <button type="button" class="btn" id="cover-modal-regen">Перегенерировать (LLM + ComfyUI)</button>
        ${(item.cover_status === 'ready') ? '<button type="button" class="btn danger" id="cover-modal-del">Удалить обложку</button>' : ''}
      </div>
    `;
    document.getElementById('cover-modal-comfy').onclick = async () => {
      const prompt = document.getElementById('cover-modal-prompt').value;
      const data = await generateCoverFromPrompt(kind === 'draft' ? 'draft' : 'pub', textId, courseCode, prompt);
      if (!data) return;
      const loaded = await loadCoverModalData(kind, textId, courseCode);
      render(loaded.item, loaded.doc);
    };
    document.getElementById('cover-modal-regen').onclick = async () => {
      dlg.close();
      await generateCover(kind === 'draft' ? 'draft' : 'pub', textId, courseCode, true);
    };
    const delBtn = document.getElementById('cover-modal-del');
    if (delBtn) {
      delBtn.onclick = async () => {
        dlg.close();
        await deleteCover(kind === 'draft' ? 'draft' : 'pub', textId, courseCode);
      };
    }
  };

  (async () => {
    try {
      if (prefetched && (prefetched.cover_thumb_rel_path || prefetched.cover_image_prompt)) {
        let doc = null;
        try {
          const loaded = await loadCoverModalData(kind, textId, courseCode);
          doc = loaded.doc;
        } catch (_) { /* optional */ }
        render(prefetched, doc);
        return;
      }
      const loaded = await loadCoverModalData(kind, textId, courseCode);
      render(loaded.item, loaded.doc);
    } catch (err) {
      bodyEl.innerHTML = `<p class="meta error">${escapeHtml(err.message)}</p>`;
    }
  })();
}

function tableWrapEl(kind) {
  const k = kind === 'draft' || kind === 'drafts' ? 'drafts' : 'published';
  return document.getElementById(k === 'drafts' ? 'drafts-table' : 'published-table');
}

function captureTableScroll(el) {
  return el ? el.scrollTop : 0;
}

function restoreTableScroll(el, scrollTop) {
  if (!el) return;
  const apply = () => { el.scrollTop = scrollTop; };
  requestAnimationFrame(() => {
    apply();
    requestAnimationFrame(apply);
  });
}

function captureListScrollState(kind) {
  const wrap = tableWrapEl(kind);
  return {
    tableScroll: captureTableScroll(wrap),
    windowY: window.scrollY || 0,
  };
}

async function reloadListPreservingScroll(kind, scrollState) {
  const state = scrollState || captureListScrollState(kind);
  if (kind === 'draft' || kind === 'drafts') {
    await loadDrafts({ preserveScroll: true, scrollTop: state.tableScroll });
  } else {
    await loadPublished({ preserveScroll: true, scrollTop: state.tableScroll });
  }
  if (state.windowY) window.scrollTo(0, state.windowY);
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

async function loadPublished(options = {}) {
  const { preserveScroll = false, scrollTop = 0 } = options;
  const wrap = document.getElementById('published-table');
  try {
    const params = new URLSearchParams();
    const course = document.getElementById('pub-course').value || courses[0]?.code;
    const level = document.getElementById('pub-level').value;
    const cover = document.getElementById('pub-cover')?.value || '';
    const git = document.getElementById('pub-git')?.value || '';
    const search = document.getElementById('pub-search')?.value.trim() || '';
    if (course) {
      params.set('course_code', course);
      localStorage.setItem('reading-cms-pub-course', course);
    }
    const sort = document.getElementById('pub-sort')?.value || 'level';
    localStorage.setItem('reading-cms-pub-sort', sort);
    if (level) params.set('level', level);
    if (cover) params.set('cover', cover);
    if (git) params.set('git', git);
    if (search) params.set('search', search);
    if (!preserveScroll) wrap.innerHTML = '<p class="meta">Загрузка…</p>';
    const data = await api('/api/published?' + params.toString());
    renderPublishedTable(data.texts || [], data.total);
    if (preserveScroll) restoreTableScroll(wrap, scrollTop);
  } catch (err) {
    wrap.innerHTML = `<p class="meta error">Ошибка: ${escapeHtml(err.message)}</p>`;
  }
}

async function loadDrafts(options = {}) {
  const { preserveScroll = false, scrollTop = 0 } = options;
  const wrap = document.getElementById('drafts-table');
  try {
    const params = new URLSearchParams();
    const course = document.getElementById('filter-course').value;
    const level = document.getElementById('filter-level').value;
    const status = document.getElementById('filter-status').value;
    const audio = document.getElementById('filter-audio').value;
    const cover = document.getElementById('filter-cover')?.value || '';
    const search = document.getElementById('filter-search').value.trim();
    if (course) params.set('course_code', course);
    if (level) params.set('level', level);
    if (status) params.set('status', status);
    if (audio) params.set('audio', audio);
    if (cover) params.set('cover', cover);
    if (search) params.set('search', search);
    const qs = params.toString();
    const data = await api('/api/drafts' + (qs ? '?' + qs : ''));
    renderDraftsTable(data.drafts || []);
    if (preserveScroll) restoreTableScroll(wrap, scrollTop);
  } catch (err) {
    wrap.innerHTML = `<p class="meta error">Ошибка: ${escapeHtml(err.message)}</p>`;
  }
}

async function syncPublishedToCMS(force = false) {
  const course = document.getElementById('pub-course').value || courses[0]?.code;
  const level = document.getElementById('pub-level').value;
  const cover = document.getElementById('pub-cover')?.value || '';
  const git = document.getElementById('pub-git')?.value || '';
  const search = document.getElementById('pub-search')?.value.trim() || '';
  if (!force && !confirm('Импортировать тексты из course в черновики CMS? Уже импортированные будут пропущены.')) return;
  toast('Импорт из course...');
  const data = await api('/api/published/sync', {
    method: 'POST',
    body: JSON.stringify({ course_code: course, level, cover, git, search, force }),
  });
  toast(`CMS: +${data.imported} новых, ${data.updated} обновлено, ${data.skipped} пропущено`);
  await loadPublished();
  await loadDrafts();
}

function speakerLabel(id) {
  const s = String(id || '').trim();
  if (!s) return '';
  if (s === 'narrator') return 'Рассказчик';
  const m = s.match(/^speaker_([a-z])$/i);
  if (m) return `Говорящий ${m[1].toUpperCase()}`;
  return s.replace(/_/g, ' ');
}

function audioUrl(kind, textId, courseCode, relPath) {
  if (!relPath) return '';
  const file = relPath.split('/').pop();
  if (kind === 'pub' || kind === 'course') {
    return `/api/course-audio/${encodeURIComponent(courseCode)}/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`;
  }
  return `/api/audio/${encodeURIComponent(textId)}/${encodeURIComponent(file)}`;
}

function renderReadableDocument(doc, meta, courseCode, kind) {
  const passage = doc?.reading_passage || {};
  const segs = passage.segments || [];
  const questions = passage.comprehension_questions || [];
  const vocab = passage.vocabulary || passage.vocabulary_items || [];
  const titleRu = doc?.title_translations?.ru || doc?.title_translations?.RU || '';

  let html = '';

  if (doc.cover_hero_rel_path && meta.cover_status === 'ready') {
    const heroUrl = coverImageUrl(kind === 'draft' ? 'draft' : 'course', doc.id, courseCode, doc.cover_hero_rel_path);
    html += `<img class="reader-hero" src="${heroUrl}?t=${Date.now()}" alt="" />`;
  }

  if (titleRu) {
    html += `<p class="reader-ru" style="margin-top:0">${escapeHtml(titleRu)}</p>`;
  }

  if (segs.length) {
    html += `<section class="reader-section"><h4>Текст</h4>`;
    html += segs.map((seg, i) => {
      const sp = speakerLabel(seg.speaker_id);
      const audio = seg.audio_rel_path
        ? `<audio class="reader-audio" controls preload="none" src="${audioUrl(kind, doc.id, courseCode, seg.audio_rel_path)}"></audio>`
        : '';
      return `<article class="reader-segment">
        ${sp ? `<div class="reader-speaker">${escapeHtml(sp)}</div>` : ''}
        <div class="reader-text">${escapeHtml(seg.text || '')}</div>
        ${seg.text_translation_ru ? `<div class="reader-ru">${escapeHtml(seg.text_translation_ru)}</div>` : ''}
        ${audio}
      </article>`;
    }).join('');
    html += `</section>`;
  }

  if (questions.length) {
    html += `<section class="reader-section"><h4>Вопросы на понимание</h4>`;
    html += questions.map((q, i) => {
      const type = q.type === 'true_false' ? 'Верно / неверно' : (q.type || '');
      return `<div class="reader-question">
        <p class="q-prompt"><strong>${i + 1}.</strong> ${escapeHtml(q.prompt || q.question || '')}</p>
        ${type ? `<div class="q-meta">${escapeHtml(type)}</div>` : ''}
        ${q.explanation ? `<div class="q-meta" style="margin-top:6px">${escapeHtml(q.explanation)}</div>` : ''}
      </div>`;
    }).join('');
    html += `</section>`;
  }

  if (Array.isArray(vocab) && vocab.length) {
    html += `<section class="reader-section"><h4>Словарь</h4><ul>`;
    html += vocab.map(v => {
      if (typeof v === 'string') return `<li>${escapeHtml(v)}</li>`;
      const word = v.word || v.term || v.surface || '';
      const gloss = v.translation_ru || v.gloss_ru || v.ru || '';
      return `<li><strong>${escapeHtml(word)}</strong>${gloss ? ` — ${escapeHtml(gloss)}` : ''}</li>`;
    }).join('');
    html += `</ul></section>`;
  }

  if (!html) {
    html = '<p class="meta">Нет сегментов для отображения.</p>';
  }
  return html;
}

let readerContext = null;

async function openTextReader(kind, textId, courseCode) {
  const dlg = document.getElementById('text-reader-dialog');
  const body = document.getElementById('text-reader-body');
  const titleEl = document.getElementById('text-reader-title');
  const metaEl = document.getElementById('text-reader-meta');
  const editBtn = document.getElementById('text-reader-edit');
  const coverBtn = document.getElementById('text-reader-cover');

  body.innerHTML = '<p class="meta">Загрузка…</p>';
  editBtn.classList.add('hidden');
  coverBtn.classList.add('hidden');
  dlg.showModal();

  try {
    let meta, doc;
    if (kind === 'draft') {
      const data = await api('/api/drafts/' + encodeURIComponent(textId));
      meta = data.draft;
      doc = data.document;
      courseCode = meta.course_code || courseCode;
      editBtn.classList.remove('hidden');
    } else {
      const cc = courseCode || document.getElementById('pub-course').value;
      const data = await api(`/api/published/detail?course_code=${encodeURIComponent(cc)}&text_id=${encodeURIComponent(textId)}`);
      meta = data.text;
      doc = data.document;
      courseCode = meta.course_code || cc;
    }

    readerContext = { kind, textId, courseCode, meta };
    titleEl.textContent = doc.title || meta.title || textId;
    metaEl.textContent = [
      textId,
      meta.level || doc.level,
      meta.target_language || doc.target_language,
      `audio ${meta.audio_status || '—'}`,
      `cover ${meta.cover_status || 'none'}`,
    ].filter(Boolean).join(' · ');

    if (meta.cover_status === 'ready') {
      coverBtn.classList.remove('hidden');
    }

    body.innerHTML = renderReadableDocument(doc, meta, courseCode, kind);
  } catch (err) {
    body.innerHTML = `<p class="meta error">${escapeHtml(err.message)}</p>`;
  }
}

function renderPreview(doc) {
  const segs = doc?.reading_passage?.segments || [];
  return segs.map(seg => {
    const audio = seg.audio_rel_path
      ? `<audio controls src="/api/audio/${doc.id}/${seg.audio_rel_path.split('/').pop()}"></audio>`
      : '';
    return `<div class="segment">
      <div class="speaker">${escapeHtml(seg.speaker_id || '')}</div>
      <div>${escapeHtml(seg.text || '')}</div>
      ${seg.text_translation_ru ? `<div class="ru">${escapeHtml(seg.text_translation_ru)}</div>` : ''}
      ${audio}
    </div>`;
  }).join('');
}

function renderCoverPanel(doc, draft) {
  const panel = document.getElementById('detail-cover');
  const thumb = doc?.cover_thumb_rel_path || '';
  const hero = doc?.cover_hero_rel_path || '';
  if (!thumb && !hero) {
    panel.classList.add('hidden');
    panel.innerHTML = '';
    return;
  }
  panel.classList.remove('hidden');
  const thumbUrl = thumb ? `/api/images/${doc.id}/${thumb.split('/').pop()}` : '';
  const heroUrl = hero ? `/api/images/${doc.id}/${hero.split('/').pop()}` : '';
  panel.innerHTML = `
    <div class="cover-status">cover: ${draft.cover_status || 'none'}</div>
    ${thumbUrl ? `<img class="cover-thumb" src="${thumbUrl}" alt="thumb" />` : ''}
    ${heroUrl ? `<img class="cover-hero" src="${heroUrl}" alt="hero" />` : ''}
    ${draft.cover_image_prompt ? `<pre class="cover-prompt">${escapeHtml(draft.cover_image_prompt)}</pre>` : ''}
  `;
}

async function openDraft(id) {
  currentDraftId = id;
  const data = await api('/api/drafts/' + encodeURIComponent(id));
  const d = data.draft;
  const doc = data.document;
  document.getElementById('detail-title').textContent = d.title;
  document.getElementById('detail-meta').textContent =
    `${d.text_id} · ${d.course_code} · ${d.level} · ${d.status} · audio ${d.audio_status} (${d.segments_with_audio}/${d.segments_total}) · cover ${d.cover_status || 'none'}`;
  renderCoverPanel(doc, d);
  document.getElementById('detail-preview').innerHTML = renderPreview(doc);
  document.getElementById('detail-edit').value = JSON.stringify(doc, null, 2);
  document.getElementById('detail-log').textContent = d.last_job_log || '';
  document.getElementById('detail-dialog').showModal();
}

async function draftAction(action) {
  if (!currentDraftId) return;
  const data = await api(`/api/drafts/${encodeURIComponent(currentDraftId)}/${action}`, { method: 'POST' });
  toast(`OK: ${action}`);
  document.getElementById('detail-log').textContent = data.draft?.last_job_log || '';
  await openDraft(currentDraftId);
  await loadDrafts();
}

document.getElementById('generate-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  toast('Generating... (may take a while)');
  try {
    const body = {
      course_code: document.getElementById('gen-course').value,
      level: document.getElementById('gen-level').value,
      format: document.getElementById('gen-format').value,
      count: Number(document.getElementById('gen-count').value) || 1,
      title: document.getElementById('gen-title').value.trim(),
      with_audio: document.getElementById('gen-audio').checked,
    };
    const data = await api('/api/drafts/generate', { method: 'POST', body: JSON.stringify(body) });
    toast(`Generated ${data.total} draft(s)`);
    document.getElementById('generate-dialog').close();
    switchTab('drafts');
    await loadDrafts();
    await loadPublished();
  } catch (err) {
    toast(err.message);
  }
});

document.getElementById('import-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  toast('LLM structuring + TTS... (может занять несколько минут)');
  try {
    const body = {
      course_code: document.getElementById('import-course').value,
      level: document.getElementById('import-level').value,
      format: document.getElementById('import-format').value,
      title: document.getElementById('import-title').value.trim(),
      text: document.getElementById('import-text').value,
      with_audio: document.getElementById('import-audio').checked,
      auto_publish: document.getElementById('import-publish').checked,
      sync_bundle: document.getElementById('import-sync-bundle').checked,
    };
    const data = await api('/api/drafts/import-text', { method: 'POST', body: JSON.stringify(body) });
    const st = data.draft?.status || 'ok';
    toast(`Готово: ${st}`);
    document.getElementById('import-text').value = '';
    document.getElementById('import-dialog').close();
    switchTab('drafts');
    await loadDrafts();
    await loadPublished();
  } catch (err) {
    toast(err.message);
  }
});

async function copyReadingPrompt(kind) {
  try {
    const body = {
      course_code: document.getElementById('import-course').value,
      level: document.getElementById('import-level').value,
      format: document.getElementById('import-format').value,
      title: document.getElementById('import-title').value.trim(),
      kind,
    };
    if (kind === 'transform') {
      body.source_text = document.getElementById('import-text').value;
      if (!body.source_text.trim()) {
        toast('Вставьте текст выше для промпта трансформации');
        return;
      }
    }
    const data = await api('/api/prompts/reading', { method: 'POST', body: JSON.stringify(body) });
    await navigator.clipboard.writeText(data.prompt || '');
    const course = courses.find(c => c.code === body.course_code);
    toast(`Промпт (${course?.title || body.course_code}, ${kind}) скопирован`);
  } catch (err) {
    toast(err.message);
  }
}

document.getElementById('copy-prompt-generate').addEventListener('click', () => copyReadingPrompt('generate'));
document.getElementById('copy-prompt-transform').addEventListener('click', () => copyReadingPrompt('transform'));

document.getElementById('import-json-file').addEventListener('change', async (e) => {
  const file = e.target.files?.[0];
  if (!file) return;
  try {
    document.getElementById('import-json').value = await file.text();
    toast(`Loaded ${file.name}`);
  } catch (err) {
    toast(err.message);
  }
});

document.getElementById('import-json-btn').addEventListener('click', async () => {
  const raw = document.getElementById('import-json').value.trim();
  if (!raw) return;
  toast('JSON batch import + TTS...');
  try {
    const data = await api('/api/drafts/import-json-batch', {
      method: 'POST',
      body: JSON.stringify({
        course_code: document.getElementById('import-course').value,
        level: document.getElementById('import-level').value,
        format: document.getElementById('import-format').value,
        title: document.getElementById('import-title').value.trim(),
        documents_text: raw,
        with_audio: document.getElementById('json-audio').checked,
        auto_publish: document.getElementById('json-publish').checked,
        sync_bundle: document.getElementById('json-sync-bundle').checked,
      }),
    });
    toast(`JSON batch: ${data.succeeded || 0}/${data.total || 0} ok, ${data.failed || 0} failed`);
    if (!data.failed) {
      document.getElementById('import-json').value = '';
      document.getElementById('import-json-file').value = '';
      document.getElementById('import-dialog').close();
    }
    switchTab('drafts');
    await loadDrafts();
    await loadPublished();
  } catch (err) {
    toast(err.message);
  }
});

document.getElementById('refresh-drafts').addEventListener('click', loadDrafts);
document.getElementById('refresh-published').addEventListener('click', loadPublished);
document.getElementById('pub-sync-cms').addEventListener('click', () => syncPublishedToCMS(false));
['pub-course','pub-level','pub-cover','pub-git','pub-sort'].forEach(id => {
  document.getElementById(id).addEventListener('change', loadPublished);
});
document.getElementById('pub-search').addEventListener('input', () => {
  clearTimeout(window._pubSearchT);
  window._pubSearchT = setTimeout(loadPublished, 300);
});
['filter-course','filter-level','filter-status','filter-audio','filter-cover'].forEach(id => {
  document.getElementById(id).addEventListener('change', loadDrafts);
});
document.getElementById('filter-search').addEventListener('input', () => {
  clearTimeout(window._searchT);
  window._searchT = setTimeout(loadDrafts, 300);
});

document.getElementById('btn-save-edit').addEventListener('click', async () => {
  if (!currentDraftId) return;
  let documentJson;
  try {
    documentJson = JSON.parse(document.getElementById('detail-edit').value);
  } catch {
    toast('Невалидный JSON');
    return;
  }
  await api('/api/drafts/' + encodeURIComponent(currentDraftId), {
    method: 'PUT',
    body: JSON.stringify({ document: documentJson }),
  });
  toast('Сохранено');
  await openDraft(currentDraftId);
  await loadDrafts();
});

document.getElementById('btn-approve').addEventListener('click', () => draftAction('approve'));
document.getElementById('btn-reject').addEventListener('click', () => draftAction('reject'));
document.getElementById('btn-audio').addEventListener('click', () => draftAction('audio'));
document.getElementById('btn-cover').addEventListener('click', async () => {
  if (!currentDraftId) return;
  const scrollState = captureListScrollState('draft');
  toast('Generating cover (LLM + ComfyUI)...');
  try {
    const force = document.getElementById('cover-force').checked;
    const data = await api(`/api/drafts/${encodeURIComponent(currentDraftId)}/cover`, {
      method: 'POST',
      body: JSON.stringify({ force }),
    });
    toast(`Cover: ${data.draft?.cover_status || 'ok'}`);
    document.getElementById('detail-log').textContent = data.draft?.last_job_log || '';
    await openDraft(currentDraftId);
    await reloadListPreservingScroll('draft', scrollState);
  } catch (err) {
    toast(err.message);
  }
});
document.getElementById('cover-batch-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const courseCode = document.getElementById('cover-batch-course').value;
  try {
    const data = await api('/api/covers/batch', {
      method: 'POST',
      body: JSON.stringify({
        course_code: courseCode,
        level: document.getElementById('cover-batch-level').value,
        force: document.getElementById('cover-batch-force').checked,
      }),
    });
    document.getElementById('batch-dialog').close();
    const batchScrollState = captureListScrollState('pub');
    const ui = openCoverProgressDialog({
      batch: true,
      batchId: data.batch_id,
      total: data.total,
    });
    let poll = null;
    let finished = false;
    const finish = async (b) => {
      if (finished) return;
      finished = true;
      if (poll) clearInterval(poll);
      poll = null;
      const failed = Number(b?.failed) || 0;
      finishCoverBatchProgressDialog(ui, b, failed > 0 || !!b?.error, b?.error);
      ui.closeBtn.disabled = false;
      toast(failed > 0
        ? `Batch: ${b.completed || 0} ok, ${b.failed} fail`
        : `Batch done: ${b.completed || 0} ok, ${b.skipped || 0} skip`);
      switchTab('texts');
      await reloadListPreservingScroll('pub', batchScrollState);
    };
    poll = startCoverBatchProgressPoll(data.batch_id, ui, finish);
  } catch (err) {
    toast(err.message);
  }
});
document.getElementById('btn-publish').addEventListener('click', async () => {
  if (!currentDraftId) return;
  const sync = document.getElementById('sync-bundle').checked;
  await api(`/api/drafts/${encodeURIComponent(currentDraftId)}/publish`, {
    method: 'POST',
    body: JSON.stringify({ sync_bundle: sync }),
  });
  toast('Published');
  await loadDrafts();
  await loadPublished();
});
document.getElementById('btn-delete-draft').addEventListener('click', async () => {
  if (!currentDraftId || !confirm('Delete draft?')) return;
  const scrollState = captureListScrollState('draft');
  await api('/api/drafts/' + encodeURIComponent(currentDraftId), { method: 'DELETE' });
  document.getElementById('detail-dialog').close();
  currentDraftId = null;
  await reloadListPreservingScroll('draft', scrollState);
});
document.getElementById('text-reader-close').addEventListener('click', () => {
  document.getElementById('text-reader-dialog').close();
});
document.getElementById('text-reader-edit').addEventListener('click', () => {
  document.getElementById('text-reader-dialog').close();
  if (readerContext?.textId) openDraft(readerContext.textId);
});
document.getElementById('text-reader-cover').addEventListener('click', () => {
  if (!readerContext) return;
  const { kind, textId, courseCode, meta } = readerContext;
  document.getElementById('text-reader-dialog').close();
  if (meta?.cover_status === 'ready') {
    openCoverModal(kind === 'draft' ? 'draft' : 'pub', textId, courseCode);
  }
});

document.getElementById('detail-close').addEventListener('click', () => document.getElementById('detail-dialog').close());

function switchTab(name) {
  document.querySelectorAll('.tab').forEach(btn => {
    const on = btn.dataset.tab === name;
    btn.classList.toggle('active', on);
    btn.setAttribute('aria-selected', on ? 'true' : 'false');
  });
  document.getElementById('view-texts').classList.toggle('active', name === 'texts');
  document.getElementById('view-texts').classList.toggle('hidden', name !== 'texts');
  document.getElementById('view-drafts').classList.toggle('active', name === 'drafts');
  document.getElementById('view-drafts').classList.toggle('hidden', name !== 'drafts');
  localStorage.setItem('reading-cms-tab', name);
}

document.querySelectorAll('.tab').forEach(btn => {
  btn.addEventListener('click', () => switchTab(btn.dataset.tab));
});

function openDialog(id) {
  const dlg = document.getElementById(id);
  if (dlg) dlg.showModal();
}

document.getElementById('open-generate-modal').addEventListener('click', () => openDialog('generate-dialog'));
document.getElementById('open-import-modal').addEventListener('click', () => openDialog('import-dialog'));
document.getElementById('open-batch-modal').addEventListener('click', () => openDialog('batch-dialog'));

document.querySelectorAll('[data-close]').forEach(btn => {
  btn.addEventListener('click', () => {
    const dlg = document.getElementById(btn.dataset.close);
    if (dlg) dlg.close();
  });
});

document.querySelectorAll('.import-tab').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.import-tab').forEach(t => t.classList.toggle('active', t === btn));
    const isJson = btn.dataset.import === 'json';
    document.getElementById('import-panel-llm').classList.toggle('hidden', isJson);
    document.getElementById('import-panel-json').classList.toggle('hidden', !isJson);
  });
});

(async function init() {
  const data = await api('/api/courses');
  courses = data.courses || [];
  fillCourseSelects();
  const savedSort = localStorage.getItem('reading-cms-pub-sort');
  const sortEl = document.getElementById('pub-sort');
  if (savedSort && sortEl && [...sortEl.options].some(o => o.value === savedSort)) {
    sortEl.value = savedSort;
  }
  switchTab(localStorage.getItem('reading-cms-tab') || 'texts');
  await Promise.all([loadPublished(), loadDrafts()]);
})();
