import { ScanProjects, CheckStatus, OpenInVSCode, GetBranches, GetCurrentBranch, SwitchBranch, GetModifiedFiles, GetCommitHistory, SuperSync, SelectDirectory, GetConfig, GetSubfolders, GetRecentProjects, SaveRecentProject, GetQuickLinks, AddQuickLink, UseQuickLink, RemoveQuickLink, GetGitHubURL, OpenURL, CheckoutCommit, GetContributionStats, OpenTerminal } from './wailsjs/go/main/App.js';

const $ = id => document.getElementById(id);
const esc = s => s ? s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;') : '';
const trunc = (s,n) => s && s.length>n ? s.slice(0,n)+'…' : (s||'');

const state = { rootDir:'', projects:[], filtered:[], activeView:'dashboard', activeCategory:'All', searchQuery:'', selectedProject:null, activeBranch:'', syncing:false, statusCache:new Map(), topPeriod:'week' };

// --- Toast ---
function toast(msg, type='info', dur=4000) {
  const icons={success:'✓',error:'✕',info:'i',warning:'!'};
  const t=document.createElement('div');
  t.className=`toast toast-${type}`;
  t.innerHTML=`<span class="toast-icon">${icons[type]}</span><span class="toast-msg">${msg}</span>`;
  $('toast-container').appendChild(t);
  setTimeout(()=>{t.style.animation='toastOut .3s ease forwards';t.addEventListener('animationend',()=>t.remove());},dur);
}
function setStatus(txt){$('status-text').textContent=txt;}
function showMain(name){['welcome-state','loading-state','projects-grid','empty-state','dashboard-view'].forEach(id=>$(id).classList.add('hidden'));$(name).classList.remove('hidden');}
function setActiveNav(id){document.querySelectorAll('.nav-item').forEach(b=>b.classList.remove('active'));if($(id))$(id).classList.add('active');}

// --- World Clocks ---
const CLOCKS=[{label:'Spain',tz:'Europe/Madrid'},{label:'UK',tz:'Europe/London'},{label:'US East',tz:'America/New_York'},{label:'Australia',tz:'Australia/Sydney'}];
function updateClocks(){
  const now=new Date();
  CLOCKS.forEach(({label,tz})=>{
    const te=$(`clock-${label}`), de=$(`clock-date-${label}`);
    if(!te)return;
    te.textContent=now.toLocaleTimeString('en-GB',{timeZone:tz,hour:'2-digit',minute:'2-digit',second:'2-digit'});
    de.textContent=now.toLocaleDateString('en-GB',{timeZone:tz,weekday:'short',day:'numeric',month:'short'});
  });
}

// --- Contribution Grid ---
function contribLevel(n){if(n===0)return 0;if(n<=2)return 1;if(n<=6)return 2;if(n<=14)return 3;return 4;}

let contribActiveFolder = 'All';

function buildContribFolderTabs(folders) {
  const container = $('contrib-folder-tabs');
  if (!container) return;
  const allFolders = ['All', ...folders];
  container.innerHTML = allFolders.map(f =>
    `<button class="period-tab${f === contribActiveFolder ? ' active' : ''}" data-folder="${esc(f)}">${esc(f)}</button>`
  ).join('');
  container.querySelectorAll('.period-tab').forEach(btn => {
    btn.addEventListener('click', () => {
      contribActiveFolder = btn.dataset.folder;
      container.querySelectorAll('.period-tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      loadContribGrid(contribActiveFolder);
    });
  });
}

async function loadContribGrid(folder) {
  const grid = $('contrib-grid');
  const badge = $('contrib-total');
  grid.innerHTML = '';
  badge.textContent = 'Loading...';
  try {
    const stats = await GetContributionStats(folder || 'All');
    badge.textContent = `${stats.total.toLocaleString()} interactions`;
    stats.days.forEach(day => {
      const cell = document.createElement('span');
      cell.className = 'contrib-cell';
      cell.dataset.level = contribLevel(day.count);
      cell.title = `${day.date}: ${day.count} commit${day.count !== 1 ? 's' : ''}`;
      grid.appendChild(cell);
    });
  } catch { badge.textContent = '—'; }
}

// --- Top Repos (removed) ---

// --- Recent Repos ---
async function renderRecentRepos(){
  const list=$('recent-repos-list');
  try {
    const recents=await GetRecentProjects();
    if(!recents||!recents.length){list.innerHTML='<p class="dash-empty">No recently accessed repositories.</p>';return;}
    list.innerHTML=recents.map(p=>`<div class="recent-repo-item" data-path="${esc(p.path)}"><div class="recent-repo-name">${esc(p.name)}</div><div class="recent-repo-path">${esc(p.path)}</div><span class="recent-repo-cat">${esc(p.category)}</span></div>`).join('');
    list.querySelectorAll('.recent-repo-item').forEach(el=>{
      el.addEventListener('click',()=>{const proj=state.projects.find(p=>p.path===el.dataset.path)||{name:el.querySelector('.recent-repo-name').textContent,path:el.dataset.path,category:el.querySelector('.recent-repo-cat').textContent};openDetail(proj);});
    });
  } catch { list.innerHTML='<p class="dash-empty">Could not load.</p>'; }
}

// --- Quick Links Cards ---
async function renderQLCards(){
  const container=$('quicklinks-cards');
  container.innerHTML='';
  try {
    const links=await GetQuickLinks();
    if(!links||!links.length){container.innerHTML='<p class="dash-empty">No quick links yet.</p>';return;}
    links.forEach(lnk=>{
      let domain='';
      try{domain=new URL(lnk.url).hostname;}catch{}
      const card=document.createElement('div');
      card.className='ql-card';
      card.innerHTML=`
        <div class="ql-card-name">${esc(lnk.name)}</div>
        <div class="ql-card-domain">${esc(domain)}</div>
        ${lnk.useCount>0?`<div class="ql-card-uses">${lnk.useCount} visit${lnk.useCount!==1?'s':''}</div>`:''}
        <button class="ql-card-remove" data-url="${esc(lnk.url)}" title="Remove">✕</button>`;
      card.addEventListener('click',async e=>{
        if(e.target.classList.contains('ql-card-remove'))return;
        await UseQuickLink(lnk.url);
        OpenURL(lnk.url);
        setTimeout(renderQLCards,200);
      });
      card.querySelector('.ql-card-remove').addEventListener('click',async e=>{
        e.stopPropagation();
        await RemoveQuickLink(lnk.url);
        renderQLCards();
      });
      container.appendChild(card);
    });
  } catch {}
}

// --- Dashboard ---
async function showDashboard(){
  state.activeView='dashboard';
  setActiveNav('nav-dashboard');
  showMain('dashboard-view');
  updateClocks();
  renderRecentRepos();
  renderQLCards();
  loadContribGrid(contribActiveFolder);
}

// --- Period tabs (removed, now contrib folder tabs) ---

// --- Scan ---
async function loadProjects(rootDir){
  showMain('loading-state');
  setStatus('Scanning...');
  try {
    const projects=await ScanProjects(rootDir||state.rootDir);
    state.projects=projects||[];
    updateSidebarBadges();
    applyFilter();
    setStatus(`${state.projects.length} repositories`);
    fetchStatusBg();
  } catch(e){toast(`Scan failed: ${e}`,'error');showMain('welcome-state');}
}

function updateSidebarBadges(){
  // Update All badge
  const allBadge=$('badge-All');
  if(allBadge)allBadge.textContent=state.projects.length;
  // Update per-folder badges
  state.projects.forEach(p=>{
    const badge=$(`badge-${CSS.escape(p.category)}`);
    if(badge){const cur=parseInt(badge.textContent)||0;badge.textContent=cur+1;}
  });
}

function applyFilter(){
  const q=state.searchQuery.toLowerCase();
  state.filtered=state.projects.filter(p=>{
    const mc=state.activeCategory==='All'||p.category===state.activeCategory;
    const mq=!q||p.name.toLowerCase().includes(q)||p.path.toLowerCase().includes(q);
    return mc&&mq;
  });
  if(!state.filtered.length&&state.projects.length){showMain('empty-state');return;}
  if(!state.filtered.length){showMain('welcome-state');return;}
  showMain('projects-grid');
  renderCards();
}

function buildStatusHtml(s){
  if(!s)return{cls:'status-loading',label:'Checking...'};
  if(s.hasError)return{cls:'status-loading',label:'No remote'};
  if(s.behind>0)return{cls:'status-behind',label:`${s.behind} behind`};
  if(s.ahead>0)return{cls:'status-ahead',label:`${s.ahead} ahead`};
  if(!s.clean)return{cls:'status-dirty',label:'Uncommitted'};
  return{cls:'status-clean',label:'Up to date'};
}

function renderCards(){
  const grid=$('projects-grid');
  grid.innerHTML='';
  state.filtered.forEach(p=>grid.appendChild(makeCard(p)));
}

function makeCard(p){
  const cached=state.statusCache.get(p.path);
  const{cls,label}=buildStatusHtml(cached);
  const card=document.createElement('div');
  card.className='project-card';
  card.innerHTML=`
    <div class="card-top"><div class="card-name">${esc(p.name)}</div><span class="card-category-badge">${esc(p.category)}</span></div>
    <div class="card-path" title="${esc(p.path)}">${esc(p.path)}</div>
    <div class="card-footer">
      <div class="card-status ${cls}" id="cs-${CSS.escape(p.path)}"><span class="status-dot ${cached?'':'pulse'}"></span>${label}</div>
      <button class="btn-vscode" data-path="${esc(p.path)}"><svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M23.15 2.587L18.21.21a1.494 1.494 0 0 0-1.705.29l-9.46 8.63-4.12-3.128a.999.999 0 0 0-1.276.057L.327 7.261A1 1 0 0 0 .326 8.74L3.899 12 .326 15.26a1 1 0 0 0 .001 1.479L1.65 17.94a.999.999 0 0 0 1.276.057l4.12-3.128 9.46 8.63a1.492 1.492 0 0 0 1.704.29l4.942-2.377A1.5 1.5 0 0 0 24 20.06V3.939a1.5 1.5 0 0 0-.85-1.352zm-5.146 14.861L10.826 12l7.178-5.448v10.896z"/></svg>Code</button>
    </div>`;
  card.addEventListener('click',e=>{if(!e.target.closest('.btn-vscode'))openDetail(p);});
  card.querySelector('.btn-vscode').addEventListener('click',async e=>{e.stopPropagation();await OpenInVSCode(p.path);toast(`Opening ${p.name}`,'info',2000);});
  return card;
}

function updateCardStatus(path,s){
  state.statusCache.set(path,s);
  const el=$(`cs-${CSS.escape(path)}`);
  if(!el)return;
  const{cls,label}=buildStatusHtml(s);
  el.className=`card-status ${cls}`;
  el.innerHTML=`<span class="status-dot"></span>${label}`;
}

async function fetchStatusBg(){
  const LIMIT=5,queue=[...state.projects];
  async function next(){if(!queue.length)return;const p=queue.shift();try{updateCardStatus(p.path,await CheckStatus(p.path));}catch{}await next();}
  await Promise.all(Array.from({length:LIMIT},next));
  setStatus(`${state.projects.length} repositories · refreshed`);
}

// --- Sidebar repos ---
async function buildReposSidebar(rootDir){
  const header=$('nav-repos-header'),list=$('nav-repos-list');
  list.innerHTML='';
  if(!rootDir){header.style.display='none';return;}
  header.style.display='';
  const folders=await GetSubfolders(rootDir);

  function makeNavBtn(label,category,icon,badgeId){
    const btn=document.createElement('button');
    btn.className='nav-item'+(state.activeCategory===category&&state.activeView==='repos'?' active':'');
    btn.id=`nav-cat-${CSS.escape(category)}`;
    btn.innerHTML=`${icon} ${esc(label)} <span class="nav-badge" id="${badgeId}">0</span>`;
    btn.addEventListener('click',()=>{state.activeCategory=category;state.activeView='repos';setActiveNav(`nav-cat-${CSS.escape(category)}`);applyFilter();});
    return btn;
  }

  const folderSVG='<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>';
  const homeSVG='<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>';

  list.appendChild(makeNavBtn('All','All',homeSVG,'badge-All'));
  folders.forEach(f=>list.appendChild(makeNavBtn(f,f,folderSVG,`badge-${CSS.escape(f)}`)));

  // Sync folder tabs on contribution grid
  buildContribFolderTabs(folders);
}

// --- Add Quick Link modal ---
function showAddLinkModal(){$('ql-name').value='';$('ql-url').value='';$('quicklink-modal-overlay').classList.remove('hidden');$('quicklink-modal').classList.remove('hidden');$('ql-name').focus();}
function hideAddLinkModal(){$('quicklink-modal-overlay').classList.add('hidden');$('quicklink-modal').classList.add('hidden');}

// --- Detail Panel ---
async function openDetail(project){
  state.selectedProject=project;
  $('detail-repo-name').textContent=project.name;
  $('detail-repo-path').textContent=project.path;
  $('sync-output').classList.add('hidden');
  $('commit-message').value='';
  $('detail-overlay').classList.remove('hidden');
  $('detail-panel').classList.remove('hidden');
  try{await SaveRecentProject(project.path);}catch{}
  await Promise.all([loadBranches(project),loadModFiles(project),loadGraph(project)]);
}
function closeDetail(){$('detail-overlay').classList.add('hidden');$('detail-panel').classList.add('hidden');state.selectedProject=null;}

async function loadBranches(p){
  $('branch-select').innerHTML='<option>Loading...</option>';
  try{
    const[branches,cur]=await Promise.all([GetBranches(p.path),GetCurrentBranch(p.path)]);
    state.activeBranch=cur;
    $('branch-select').innerHTML=(branches||[]).map(b=>`<option value="${esc(b)}"${b===cur?' selected':''}>${esc(b)}</option>`).join('');
    updateBranchBadge();
  }catch{$('branch-select').innerHTML='<option>Error</option>';}
}
function updateBranchBadge(){
  const s=state.selectedProject?state.statusCache.get(state.selectedProject.path):null;
  const{cls,label}=buildStatusHtml(s);
  const badge=$('branch-status-badge');
  badge.className=`branch-status-badge card-status ${cls}`;
  badge.textContent=label;
}

async function loadModFiles(p){
  const list=$('modified-files-list');
  list.innerHTML='<p class="no-changes">Loading...</p>';
  try{
    const files=await GetModifiedFiles(p.path);
    if(!files||!files.length){list.innerHTML='<div class="no-changes">Working tree clean</div>';return;}
    list.innerHTML=files.map((f,i)=>{
      const cls=f.status.includes('M')?'status-M':f.status.includes('A')?'status-A':f.status.includes('D')?'status-D':'status-UU';
      return `<label class="file-item" for="fi-${i}"><input type="checkbox" id="fi-${i}" data-path="${esc(f.path)}" checked/><span class="file-status-code ${cls}">${esc(f.status)}</span><span class="file-path">${esc(f.path)}</span></label>`;
    }).join('');
  }catch{list.innerHTML='<p class="no-changes">Error loading files</p>';}
}

async function loadGraph(p){
  const el=$('git-graph');
  el.innerHTML='<p style="padding:10px;font-size:11px;color:var(--text-muted)">Loading...</p>';
  try{const commits=await GetCommitHistory(p.path,state.activeBranch);renderGraph(commits);}
  catch{el.innerHTML='<p style="padding:10px;font-size:11px;color:var(--text-muted)">Could not load</p>';}
}

function renderGraph(commits){
  const el=$('git-graph');
  el.innerHTML='';
  if(!commits||!commits.length){el.innerHTML='<p style="padding:10px;font-size:11px;color:var(--text-muted)">No commits</p>';return;}

  if(typeof GitgraphJS!=='undefined'){
    try{
      const gitgraph=GitgraphJS.createGitgraph(el,{
        template:GitgraphJS.templateExtend(GitgraphJS.TemplateName.Metro,{
          colors:['#7c3aed','#16a34a','#2563eb','#d97706','#dc2626'],
          commit:{
            message:{displayAuthor:false,displayHash:true,font:'10px Inter,sans-serif',color:'#57606a'},
            dot:{size:5},
            spacing:28,
          },
          branch:{
            lineWidth:1.5,
            spacing:20,
            label:{display:false}, // hide branch name labels to prevent overflow
          },
        }),
      });
      const branchMap=new Map();
      const defName=state.activeBranch||'main';
      const defBranch=gitgraph.branch(defName);
      branchMap.set(defName,defBranch);

      commits.slice(0,80).reverse().forEach(c=>{
        const refs=(c.refs||'').split(',').map(r=>r.trim().replace(/^HEAD -> /,'').replace(/^origin\//,'').trim()).filter(r=>r&&!r.startsWith('tag:'));
        let branch=defBranch;
        if(refs.length){const bn=refs[0];if(!branchMap.has(bn))branchMap.set(bn,gitgraph.branch(bn));branch=branchMap.get(bn);}
        branch.commit({
          subject:trunc(c.message,48),
          hash:c.hash.slice(0,7),
          onClick:()=>handleCommitClick(c),
        });
      });
      return;
    }catch{}
  }
  // Fallback: clickable list
  el.innerHTML=commits.slice(0,60).map(c=>`
    <div class="commit-row" data-hash="${esc(c.hash)}" style="display:flex;gap:8px;align-items:baseline;padding:5px 10px;border-bottom:1px solid var(--border);cursor:pointer" title="Click to checkout">
      <span style="font-family:JetBrains Mono,monospace;font-size:10px;color:var(--accent);flex-shrink:0">${c.hash.slice(0,7)}</span>
      <span style="font-size:11px;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${esc(c.message)}</span>
      <span style="font-size:10px;color:var(--text-muted);white-space:nowrap">${esc(c.date)}</span>
    </div>`).join('');
  el.querySelectorAll('.commit-row').forEach(row=>{
    row.addEventListener('mouseenter',()=>row.style.background='var(--bg-elevated)');
    row.addEventListener('mouseleave',()=>row.style.background='');
    row.addEventListener('click',()=>handleCommitClick({hash:row.dataset.hash}));
  });
}

async function handleCommitClick(c){
  if(!state.selectedProject)return;
  const hash=c.hash.slice(0,40);
  if(!confirm(`Checkout commit ${hash.slice(0,7)}?\nThis will put the repo in detached HEAD state.`))return;
  const r=await CheckoutCommit(state.selectedProject.path,hash);
  if(r==='ok'){toast(`Checked out ${hash.slice(0,7)}`,'success',3000);await loadBranches(state.selectedProject);}
  else toast(r,'error');
}

// --- Super Sync ---
async function runSync(){
  if(state.syncing||!state.selectedProject)return;
  const files=Array.from($('modified-files-list').querySelectorAll('input:checked')).map(cb=>cb.dataset.path);
  const msg=$('commit-message').value.trim();
  if(!files.length){toast('Select at least one file','warning');return;}
  if(!msg){toast('Enter a commit message','warning');$('commit-message').focus();return;}
  state.syncing=true;
  $('sync-btn-label').textContent='Syncing...';
  $('btn-sync').disabled=true;
  try{
    const r=await SuperSync(state.selectedProject.path,files,msg);
    $('sync-output').textContent=r.output||r.error||'';
    $('sync-output').classList.remove('hidden');
    if(r.success){
      toast('Pushed successfully!','success');
      $('commit-message').value='';
      await Promise.all([loadModFiles(state.selectedProject),loadGraph(state.selectedProject)]);
      try{const s=await CheckStatus(state.selectedProject.path);updateCardStatus(state.selectedProject.path,s);updateBranchBadge();}catch{}
    }else toast(`Failed: ${r.error}`,'error',8000);
  }catch(e){toast(`Error: ${e}`,'error');}
  finally{state.syncing=false;$('sync-btn-label').textContent='Sync — Add, Commit, Push';$('btn-sync').disabled=false;}
}

// --- Init ---
async function init(){
  setInterval(updateClocks,1000);

  $('nav-dashboard').addEventListener('click',showDashboard);

  async function pickDir(){
    const dir=await SelectDirectory();
    if(!dir)return;
    state.rootDir=dir;
    $('root-dir-display').textContent=dir;
    await buildReposSidebar(dir);
    await loadProjects(dir);
  }
  $('btn-pick-dir').addEventListener('click',pickDir);
  $('btn-pick-dir-welcome').addEventListener('click',pickDir);
  $('btn-refresh').addEventListener('click',async()=>{
    if(!state.rootDir)return;
    state.statusCache.clear();
    await buildReposSidebar(state.rootDir);
    await loadProjects(state.rootDir);
  });

  $('search-input').addEventListener('input',e=>{
    state.searchQuery=e.target.value;
    $('search-clear').classList.toggle('hidden',!state.searchQuery);
    state.activeView='repos';applyFilter();
  });
  $('search-clear').addEventListener('click',()=>{$('search-input').value='';state.searchQuery='';$('search-clear').classList.add('hidden');applyFilter();});

  $('detail-close').addEventListener('click',closeDetail);
  $('detail-overlay').addEventListener('click',closeDetail);
  $('detail-terminal').addEventListener('click',async()=>{
    if(!state.selectedProject)return;
    const r=await OpenTerminal(state.selectedProject.path);
    if(r==='ok')toast('Terminal opened','info',2000);
    else toast(r,'error');
  });
  $('detail-vscode').addEventListener('click',async()=>{if(state.selectedProject){await OpenInVSCode(state.selectedProject.path);toast('Opening VS Code...','info',2000);}});
  $('detail-github').addEventListener('click',async()=>{
    if(!state.selectedProject)return;
    const url=await GetGitHubURL(state.selectedProject.path);
    if(url)OpenURL(url);else toast('No GitHub remote found','warning');
  });
  $('branch-select').addEventListener('change',async()=>{
    if(!state.selectedProject)return;
    const b=$('branch-select').value;
    const r=await SwitchBranch(state.selectedProject.path,b);
    if(r==='ok'){state.activeBranch=b;toast(`Switched to ${b}`,'success',2000);await Promise.all([loadModFiles(state.selectedProject),loadGraph(state.selectedProject)]);}
    else{toast(r,'error');$('branch-select').value=state.activeBranch;}
  });
  $('btn-select-all').addEventListener('click',()=>{
    const boxes=$('modified-files-list').querySelectorAll('input[type="checkbox"]');
    const all=Array.from(boxes).every(b=>b.checked);
    boxes.forEach(b=>b.checked=!all);
    $('btn-select-all').textContent=all?'Select All':'Deselect All';
  });
  $('btn-sync').addEventListener('click',runSync);
  $('commit-message').addEventListener('keydown',e=>{if(e.key==='Enter'&&(e.metaKey||e.ctrlKey))runSync();});

  const addQL=()=>{$('ql-name').value='';$('ql-url').value='';$('quicklink-modal-overlay').classList.remove('hidden');$('quicklink-modal').classList.remove('hidden');$('ql-name').focus();};
  $('btn-add-quicklink').addEventListener('click',addQL);
  $('btn-add-quicklink2').addEventListener('click',addQL);
  $('ql-cancel').addEventListener('click',hideAddLinkModal);
  $('quicklink-modal-overlay').addEventListener('click',hideAddLinkModal);
  $('ql-save').addEventListener('click',async()=>{
    const name=$('ql-name').value.trim(),url=$('ql-url').value.trim();
    if(!name||!url){toast('Name and URL required','warning');return;}
    await AddQuickLink(name,url);
    hideAddLinkModal();
    renderQLCards();
    toast(`Added "${name}"`,'success',2000);
  });

  try{
    const cfg=await GetConfig();
    if(cfg&&cfg.rootDir){
      state.rootDir=cfg.rootDir;
      $('root-dir-display').textContent=cfg.rootDir;
      await buildReposSidebar(cfg.rootDir);
      await loadProjects(cfg.rootDir);
    }
  }catch{}

  showDashboard();
}

window.addEventListener('DOMContentLoaded',()=>init().catch(console.error));
