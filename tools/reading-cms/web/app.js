const API = '';

let courses = [];
let currentDraftId = null;

async function api(path, opts = {}) {
  const res = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
    ...opts,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function toast(msg) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.classList.remove('hidden');
  setTimeout(() => el.classList.add('hidden'), 4000);
}

function fillCourseSelects() {
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
  }
}

function badgeAudio(st) {
  const cls = st === 'ready' ? 'ready' : st === 'partial' ? 'partial' : 'none';
  return `<span class="badge ${cls}">${st || 'none'}</span>`;
}

function renderDraftsTable(drafts) {
  const wrap = document.getElementById('drafts-table');
  if (!drafts.length) {
    wrap.innerHTML = '<p class="meta">No drafts.</p>';
    return;
  }
  wrap.innerHTML = `<table>
    <thead><tr><th>Title</th><th>Course</th><th>Level</th><th>Status</th><th>Audio</th><th>Seg</th></tr></thead>
    <tbody>${drafts.map(d => `<tr data-id="${d.text_id}">
      <td>${escapeHtml(d.title)}</td>
      <td>${d.course_code}</td>
      <td>${d.level}</td>
      <td>${d.status}</td>
      <td>${badgeAudio(d.audio_status)} ${d.segments_with_audio}/${d.segments_total}</td>
      <td>${d.segments_total}</td>
    </tr>`).join('')}</tbody></table>`;
  wrap.querySelectorAll('tbody tr').forEach(row => {
    row.addEventListener('click', () => openDraft(row.dataset.id));
  });
}

function renderPublishedTable(texts) {
  const wrap = document.getElementById('published-table');
  if (!texts.length) {
    wrap.innerHTML = '<p class="meta">No published texts in course catalog.</p>';
    return;
  }
  wrap.innerHTML = `<table>
    <thead><tr><th>Title</th><th>Level</th><th>Audio</th><th>Segments</th><th></th></tr></thead>
    <tbody>${texts.map(t => `<tr>
      <td>${escapeHtml(t.title)}</td>
      <td>${t.level}</td>
      <td>${t.audio_ready ? '✓' : '—'}</td>
      <td>${t.segments_count}</td>
      <td><button class="btn danger pub-del" data-id="${t.text_id}">Delete</button></td>
    </tr>`).join('')}</tbody></table>`;
  wrap.querySelectorAll('.pub-del').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      if (!confirm('Удалить из course/bundle?')) return;
      const course = document.getElementById('pub-course').value;
      await api(`/api/published?course_code=${encodeURIComponent(course)}&text_id=${encodeURIComponent(btn.dataset.id)}`, { method: 'DELETE' });
      toast('Deleted from course');
      loadPublished();
    });
  });
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

async function loadDrafts() {
  const params = new URLSearchParams();
  const course = document.getElementById('filter-course').value;
  const level = document.getElementById('filter-level').value;
  const status = document.getElementById('filter-status').value;
  const audio = document.getElementById('filter-audio').value;
  const search = document.getElementById('filter-search').value.trim();
  if (course) params.set('course_code', course);
  if (level) params.set('level', level);
  if (status) params.set('status', status);
  if (audio) params.set('audio', audio);
  if (search) params.set('search', search);
  const qs = params.toString();
  const data = await api('/api/drafts' + (qs ? '?' + qs : ''));
  renderDraftsTable(data.drafts || []);
}

async function loadPublished() {
  const params = new URLSearchParams();
  const course = document.getElementById('pub-course').value || courses[0]?.code;
  const level = document.getElementById('pub-level').value;
  if (course) params.set('course_code', course);
  if (level) params.set('level', level);
  const data = await api('/api/published?' + params.toString());
  renderPublishedTable(data.texts || []);
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
    await loadDrafts();
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

document.getElementById('import-json-btn').addEventListener('click', async () => {
  const raw = document.getElementById('import-json').value.trim();
  if (!raw) return;
  let documentJson;
  try { documentJson = JSON.parse(raw); } catch { toast('Invalid JSON'); return; }
  toast('JSON import + TTS...');
  try {
    const data = await api('/api/drafts/import-json', {
      method: 'POST',
      body: JSON.stringify({
        course_code: document.getElementById('import-course').value,
        level: document.getElementById('import-level').value,
        format: document.getElementById('import-format').value,
        title: document.getElementById('import-title').value.trim(),
        document: documentJson,
        with_audio: document.getElementById('json-audio').checked,
        auto_publish: document.getElementById('json-publish').checked,
        sync_bundle: document.getElementById('json-sync-bundle').checked,
      }),
    });
    toast(`JSON: ${data.draft?.status || 'ok'}`);
    document.getElementById('import-json').value = '';
    await loadDrafts();
    await loadPublished();
  } catch (err) {
    toast(err.message);
  }
});

document.getElementById('refresh-drafts').addEventListener('click', loadDrafts);
document.getElementById('refresh-published').addEventListener('click', loadPublished);
['filter-course','filter-level','filter-status','filter-audio'].forEach(id => {
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
    await loadDrafts();
  } catch (err) {
    toast(err.message);
  }
});
document.getElementById('cover-batch-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  toast('Batch cover generation started...');
  try {
    const data = await api('/api/covers/batch', {
      method: 'POST',
      body: JSON.stringify({
        course_code: document.getElementById('cover-batch-course').value,
        level: document.getElementById('cover-batch-level').value,
        force: document.getElementById('cover-batch-force').checked,
      }),
    });
    toast(`Batch done: ${data.generated ?? 0} with covers`);
    await loadPublished();
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
  await api('/api/drafts/' + encodeURIComponent(currentDraftId), { method: 'DELETE' });
  document.getElementById('detail-dialog').close();
  await loadDrafts();
});
document.getElementById('detail-close').addEventListener('click', () => document.getElementById('detail-dialog').close());

(async function init() {
  const data = await api('/api/courses');
  courses = data.courses || [];
  fillCourseSelects();
  await loadDrafts();
  await loadPublished();
})();
