(() => {
  const form = document.getElementById('upload-form');
  const dropzone = document.getElementById('dropzone');
  const input = document.getElementById('files');
  const queue = document.getElementById('queue');
  const submitBtn = document.getElementById('submit-btn');
  const resultsDiv = document.getElementById('results');
  const resultsList = document.getElementById('results-list');

  let filesList = [];

  dropzone.addEventListener('click', () => input.click());
  dropzone.addEventListener('dragover', (e) => { e.preventDefault(); dropzone.classList.add('drag'); });
  dropzone.addEventListener('dragleave', () => dropzone.classList.remove('drag'));
  dropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropzone.classList.remove('drag');
    addFiles(e.dataTransfer.files);
  });
  input.addEventListener('change', (e) => addFiles(e.target.files));

  function addFiles(files) {
    for (const f of files) {
      if (f.size > 5 * 1024 * 1024) { alert(`${f.name} excede 5 MB`); continue; }
      filesList.push(f);
      const li = document.createElement('li');
      li.textContent = `${f.name} (${(f.size / 1024).toFixed(1)} KB)`;
      queue.appendChild(li);
    }
    submitBtn.disabled = filesList.length === 0;
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    submitBtn.disabled = true;
    submitBtn.textContent = 'Enviando...';

    const fd = new FormData();
    fd.append('path', document.getElementById('path').value);
    fd.append('csrf', form.querySelector('[name=csrf]').value);
    for (const f of filesList) fd.append('files', f);

    try {
      const r = await fetch('/admin/upload', { method: 'POST', body: fd });
      if (!r.ok) {
        const text = await r.text();
        throw new Error(text || `HTTP ${r.status}`);
      }
      const json = await r.json();
      form.hidden = true;
      resultsDiv.hidden = false;
      renderResults(json.results);
    } catch (err) {
      alert('Erro: ' + err.message);
      submitBtn.disabled = false;
      submitBtn.textContent = 'Enviar arquivos';
    }
  });

  function renderResults(results) {
    resultsList.innerHTML = '';
    for (const r of results) {
      const card = document.createElement('div');
      card.className = 'result-card';

      const h3 = document.createElement('h3');
      h3.textContent = r.final_filename;
      if (r.renamed) {
        const badge = document.createElement('span');
        badge.className = 'renamed-badge';
        badge.textContent = 'renomeado';
        h3.appendChild(badge);
      }
      card.appendChild(h3);

      card.appendChild(labeledCopy('URL', r.url));
      card.appendChild(labeledCopy('Tag <img>', r.img_tag));

      resultsList.appendChild(card);
    }
  }

  function labeledCopy(label, value) {
    const wrap = document.createElement('div');

    const lab = document.createElement('div');
    lab.innerHTML = `<strong>${label}</strong>`;
    lab.style.marginTop = '0.5rem';
    wrap.appendChild(lab);

    const block = document.createElement('div');
    block.className = 'copy-block';
    block.textContent = value;
    wrap.appendChild(block);

    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'copy-btn';
    btn.textContent = '📋 Copiar';
    btn.addEventListener('click', async () => {
      try {
        await navigator.clipboard.writeText(value);
        btn.textContent = '✓ copiado!';
        setTimeout(() => { btn.textContent = '📋 Copiar'; }, 1500);
      } catch {
        btn.textContent = 'falha';
      }
    });
    wrap.appendChild(btn);

    return wrap;
  }
})();
