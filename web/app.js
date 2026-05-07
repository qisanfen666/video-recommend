function formatDurationNs(ns) {
  if (ns == null || Number.isNaN(ns)) return "—";
  const ms = Number(ns) / 1e6;
  if (ms < 1) return `${(Number(ns) / 1e3).toFixed(0)} µs`;
  return `${ms.toFixed(2)} ms`;
}

function showErr(el, msg) {
  if (!el) return;
  if (msg) {
    el.textContent = msg;
    el.hidden = false;
  } else {
    el.hidden = true;
    el.textContent = "";
  }
}

async function fetchJSON(url, options) {
  const res = await fetch(url, options);
  const text = await res.text();
  let data;
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    throw new Error(`非 JSON 响应 (${res.status})`);
  }
  if (!res.ok) {
    const err = data.error || data.message || res.statusText;
    throw new Error(typeof err === "string" ? err : `HTTP ${res.status}`);
  }
  return data;
}

async function loadMetrics() {
  const tbody = document.querySelector("#stats-table tbody");
  const recentEl = document.getElementById("recent-list");
  tbody.innerHTML = "";
  recentEl.innerHTML = "";

  const [statsData, recentData] = await Promise.all([
    fetchJSON("/debug/stats"),
    fetchJSON("/debug/recent"),
  ]);

  const interfaces = statsData.interfaces || {};
  const paths = Object.keys(interfaces).sort();
  for (const path of paths) {
    const s = interfaces[path];
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><code>${escapeHtml(path)}</code></td>
      <td>${s.Count ?? "—"}</td>
      <td>${formatDurationNs(s.AvgTime)}</td>
      <td>${formatDurationNs(s.MaxTime)}</td>
      <td>${s.SuccessRate != null ? s.SuccessRate.toFixed(1) + "%" : "—"}</td>
    `;
    tbody.appendChild(tr);
  }

  const recent = recentData.recent_requests || [];
  for (const r of recent) {
    const li = document.createElement("li");
    const codeClass = r.StatusCode >= 400 ? "bad" : "ok";
    li.innerHTML = `<span class="${codeClass}">${r.StatusCode}</span> ${escapeHtml(r.Method)} ${escapeHtml(r.Path)} · ${formatDurationNs(r.Duration)}`;
    recentEl.appendChild(li);
  }
}

function escapeHtml(s) {
  const d = document.createElement("div");
  d.textContent = s;
  return d.innerHTML;
}

function formatBytes(n) {
  if (n == null || Number.isNaN(n)) return "—";
  const x = Number(n);
  if (x < 1024) return `${x} B`;
  if (x < 1024 * 1024) return `${(x / 1024).toFixed(1)} KiB`;
  return `${(x / (1024 * 1024)).toFixed(2)} MiB`;
}

document.getElementById("form-simulate").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const count = Number(fd.get("count"));
  const heat_decay = Number(fd.get("heat_decay"));
  const seed = Number(fd.get("seed")) || 0;
  const persist = fd.get("persist") === "on";
  const meta = document.getElementById("simulate-meta");
  const errEl = document.getElementById("simulate-err");
  const cols = document.getElementById("simulate-cols");
  const tbody = document.querySelector("#simulate-hot-table tbody");
  const prefsEl = document.getElementById("simulate-prefs");
  const jsonEl = document.getElementById("simulate-json");
  const hintEl = document.getElementById("simulate-sample-hint");
  const btn = document.getElementById("btn-simulate");
  showErr(errEl, "");
  cols.hidden = true;
  hintEl.hidden = true;
  jsonEl.hidden = true;
  tbody.innerHTML = "";
  prefsEl.innerHTML = "";
  btn.disabled = true;
  meta.textContent = "正在模拟…";

  try {
    const data = await fetchJSON("/api/admin/simulate-behaviors", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        count,
        heat_decay: Number.isFinite(heat_decay) ? heat_decay : 1,
        seed,
        persist,
      }),
    });
    meta.textContent = `已应用 ${data.applied} 条 · 耗时 ${data.duration_ms} ms · 时间窗 [${data.time_window?.min_ts}, ${data.time_window?.max_ts}] · 衰减系数 ${data.heat_decay_factor}`;
    const before = data.hot_top5_before || [];
    const after = data.hot_top5_after || [];
    const rows = Math.max(before.length, after.length);
    for (let i = 0; i < rows; i++) {
      const tr = document.createElement("tr");
      const b = before[i];
      const a = after[i];
      const btxt = b ? `#${b.video_id} ${escapeHtml(b.title)} <small>heat ${Number(b.heat).toFixed(2)} · ${escapeHtml(b.category)}</small>` : "—";
      const atxt = a ? `#${a.video_id} ${escapeHtml(a.title)} <small>heat ${Number(a.heat).toFixed(2)} · ${escapeHtml(a.category)}</small>` : "—";
      tr.innerHTML = `<td>${i + 1}</td><td>${btxt}</td><td>${atxt}</td>`;
      tbody.appendChild(tr);
    }
    for (const p of data.user_preference_shifts || []) {
      const li = document.createElement("li");
      const bef = (p.preferences_before || []).join(", ") || "∅";
      const aft = (p.preferences_after || []).join(", ") || "∅";
      li.innerHTML = `<strong>#${p.user_id}</strong> ${escapeHtml(p.name)}<br/><small>本批 +${p.new_behaviors_in_batch} 条 · 累计行为 ${p.behaviors}</small><br/>偏好 <span class="warn">${escapeHtml(bef)}</span> → <span class="ok">${escapeHtml(aft)}</span>`;
      prefsEl.appendChild(li);
    }
    cols.hidden = false;
    if (data.sample_new_behaviors?.length) {
      hintEl.hidden = false;
      jsonEl.textContent = JSON.stringify(data.sample_new_behaviors, null, 2);
      jsonEl.hidden = false;
    }
    loadMetrics().catch(() => {});
  } catch (err) {
    meta.textContent = "";
    showErr(errEl, err.message);
  } finally {
    btn.disabled = false;
  }
});

document.getElementById("form-datagen").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const video_count = Number(fd.get("video_count"));
  const user_count = Number(fd.get("user_count"));
  const behavior_count = Number(fd.get("behavior_count"));
  const statusEl = document.getElementById("datagen-status");
  const errEl = document.getElementById("datagen-err");
  const tableWrap = document.getElementById("datagen-table-wrap");
  const tbody = document.querySelector("#datagen-table tbody");
  const diskEl = document.getElementById("datagen-disk");
  const reloadEl = document.getElementById("datagen-reload");
  const jsonEl = document.getElementById("datagen-json");
  const btn = document.getElementById("btn-datagen");
  showErr(errEl, "");
  tableWrap.hidden = true;
  diskEl.hidden = true;
  reloadEl.hidden = true;
  jsonEl.hidden = true;
  tbody.innerHTML = "";
  btn.disabled = true;
  statusEl.textContent = "正在生成并加载，请稍候…";

  try {
    const data = await fetchJSON("/api/admin/regenerate-data", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ video_count, user_count, behavior_count }),
    });
    const gen = data.generation;
    const rel = data.reload;
    statusEl.textContent = `生成总耗时 ${gen.total_generation_ms} ms · 完成后落盘合计 ${formatBytes(gen.disk_bytes?.total)}`;
    for (const p of gen.phases || []) {
      const tr = document.createElement("tr");
      const label = p.label ? `<div class="phase-label">${escapeHtml(p.label)}</div>` : "";
      tr.innerHTML = `
        <td><code>${escapeHtml(p.name)}</code>${label}</td>
        <td>${p.count ?? "—"}</td>
        <td>${p.duration_ms ?? "—"} ms</td>
        <td>${p.items_per_sec != null ? p.items_per_sec.toFixed(0) + " /s" : "—"}</td>`;
      tbody.appendChild(tr);
    }
    tableWrap.hidden = false;
    const d = gen.disk_bytes || {};
    diskEl.innerHTML = `落盘：videos ${formatBytes(d.videos)} · users ${formatBytes(d.users)} · behaviors ${formatBytes(d.behaviors)}`;
    diskEl.hidden = false;
    reloadEl.textContent = `重建服务索引：Load ${rel.load_ms} ms · BuildIndex ${rel.index_ms} ms · Hot堆 ${rel.init_ms} ms · 合计 ${rel.total_ms} ms`;
    reloadEl.hidden = false;
    jsonEl.textContent = JSON.stringify(data, null, 2);
    jsonEl.hidden = false;
    loadMetrics().catch(() => {});
  } catch (err) {
    statusEl.textContent = "";
    showErr(errEl, err.message);
  } finally {
    btn.disabled = false;
  }
});

document.getElementById("form-recommend").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const user_id = Number(fd.get("user_id"));
  const page_size = Number(fd.get("page_size")) || 20;
  const meta = document.getElementById("recommend-meta");
  const list = document.getElementById("recommend-list");
  const errEl = document.getElementById("recommend-err");
  showErr(errEl, "");
  list.innerHTML = "";

  try {
    const data = await fetchJSON("/api/recommend", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id, page_size }),
    });
    meta.textContent = `用户 ${data.user_id} · 返回 ${data.count} 条 · 计算耗时 ${data.time_ms} ms`;
    for (const v of data.videos || []) {
      const li = document.createElement("li");
      li.innerHTML = `<span class="score">${Number(v.score).toFixed(4)}</span><strong>#${v.id}</strong> ${escapeHtml(v.title)} <small style="color:var(--muted)">(${escapeHtml(v.reason)})</small>`;
      list.appendChild(li);
    }
    loadMetrics().catch(() => {});
  } catch (err) {
    showErr(errEl, err.message);
    meta.textContent = "";
  }
});

document.getElementById("form-similar").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const user_id = Number(fd.get("user_id"));
  const top_k = Number(fd.get("top_k")) || 10;
  const meta = document.getElementById("similar-meta");
  const list = document.getElementById("similar-list");
  const errEl = document.getElementById("similar-err");
  showErr(errEl, "");
  list.innerHTML = "";

  try {
    const data = await fetchJSON("/api/similar-users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id, top_k }),
    });
    meta.textContent = `用户 ${data.user_id} · ${data.count} 个相似用户 · ${data.time_ms} ms`;
    for (const u of data.users || []) {
      const li = document.createElement("li");
      li.innerHTML = `<span class="score">J=${u.similarity.toFixed(4)}</span><strong>#${u.id}</strong> ${escapeHtml(u.name)} <small style="color:var(--muted)">共同 ${u.common_videos}</small>`;
      list.appendChild(li);
    }
    loadMetrics().catch(() => {});
  } catch (err) {
    showErr(errEl, err.message);
    meta.textContent = "";
  }
});

document.getElementById("btn-hot").addEventListener("click", async () => {
  const list = document.getElementById("hot-list");
  const errEl = document.getElementById("hot-err");
  showErr(errEl, "");
  list.innerHTML = "";
  try {
    const data = await fetchJSON("/api/hot");
    for (const item of data.list || []) {
      const li = document.createElement("li");
      li.innerHTML = `${escapeHtml(item.Title)} <span class="heat">#${item.VideoID} · heat ${Number(item.Heat).toFixed(2)}</span>`;
      list.appendChild(li);
    }
    loadMetrics().catch(() => {});
  } catch (err) {
    showErr(errEl, err.message);
  }
});

document.getElementById("btn-refresh-metrics").addEventListener("click", () => {
  loadMetrics().catch((err) => alert(err.message));
});

document.getElementById("form-video").addEventListener("submit", async (e) => {
  e.preventDefault();
  const id = new FormData(e.target).get("id");
  await doLookup(`/api/videos/${encodeURIComponent(id)}`);
});

document.getElementById("form-user").addEventListener("submit", async (e) => {
  e.preventDefault();
  const id = new FormData(e.target).get("id");
  await doLookup(`/api/users/${encodeURIComponent(id)}`);
});

document.getElementById("form-history").addEventListener("submit", async (e) => {
  e.preventDefault();
  const id = new FormData(e.target).get("id");
  await doLookup(`/api/users/${encodeURIComponent(id)}/history`);
});

async function doLookup(url) {
  const out = document.getElementById("lookup-out");
  const errEl = document.getElementById("lookup-err");
  showErr(errEl, "");
  try {
    const data = await fetchJSON(url);
    out.textContent = JSON.stringify(data, null, 2);
    loadMetrics().catch(() => {});
  } catch (err) {
    out.textContent = "{}";
    showErr(errEl, err.message);
  }
}

loadMetrics().catch((err) => {
  document.getElementById("recent-list").innerHTML = `<li>无法加载监控：${escapeHtml(err.message)}</li>`;
});
