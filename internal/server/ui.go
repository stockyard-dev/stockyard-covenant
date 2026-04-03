package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Covenant</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--blue:#4a7ec4;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.hdr{padding:.6rem 1.2rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--serif);font-size:1rem}.hdr h1 span{color:var(--rl)}
.main{max-width:900px;margin:0 auto;padding:1rem 1.2rem}
.btn{font-family:var(--mono);font-size:.68rem;padding:.3rem .6rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s;white-space:nowrap}
.btn-p{border-color:var(--rust);color:var(--rl)}.btn-p:hover{background:var(--rust);color:var(--cream)}
.btn-d{border-color:var(--bg3);color:var(--cm)}.btn-d:hover{border-color:var(--red);color:var(--red)}
.btn-s{border-color:var(--green);color:var(--green)}.btn-s:hover{background:var(--green);color:var(--bg)}

.overview{display:flex;gap:1.5rem;margin-bottom:1.2rem;font-size:.7rem;color:var(--leather);flex-wrap:wrap;align-items:flex-end}
.overview .stat b{display:block;font-size:1.2rem;color:var(--cream)}
.compliance-big{font-size:2rem!important;font-weight:700}
.compliance-big.good{color:var(--green)!important}.compliance-big.warn{color:var(--gold)!important}.compliance-big.bad{color:var(--red)!important}

.tabs{display:flex;gap:0;margin-bottom:1rem;border-bottom:1px solid var(--bg3)}
.tab{padding:.4rem 1rem;cursor:pointer;font-size:.75rem;color:var(--cm);border-bottom:2px solid transparent;transition:.15s}
.tab:hover{color:var(--cream)}.tab.active{color:var(--rl);border-bottom-color:var(--rl)}

.policy-card{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;margin-bottom:.5rem;cursor:pointer;transition:.1s}
.policy-card:hover{background:var(--bg3)}
.policy-top{display:flex;align-items:center;gap:.5rem}
.policy-title{font-size:.82rem;font-weight:600;flex:1}
.policy-cat{font-size:.6rem;padding:.1rem .3rem;background:var(--bg3);color:var(--ll);border-radius:2px}
.policy-status{font-size:.6rem;padding:.1rem .3rem;border-radius:2px;text-transform:uppercase}
.ps-active{background:rgba(74,158,92,.15);color:var(--green)}.ps-draft{background:rgba(212,168,67,.15);color:var(--gold)}.ps-retired{background:rgba(122,112,96,.15);color:var(--cm)}
.policy-bar{height:4px;background:var(--bg3);border-radius:2px;margin-top:.4rem;overflow:hidden}
.policy-fill{height:100%;transition:width .3s}
.pf-good{background:var(--green)}.pf-warn{background:var(--gold)}.pf-bad{background:var(--red)}
.policy-meta{font-size:.65rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.7rem}

.member-row{display:flex;align-items:center;gap:.6rem;padding:.4rem .5rem;border-bottom:1px solid var(--bg3);font-size:.75rem}
.member-name{font-weight:600;flex:1}.member-dept{font-size:.65rem;color:var(--leather)}

.modal-bg{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.65);display:flex;align-items:center;justify-content:center;z-index:100}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:95%;max-width:700px;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--serif);font-size:.95rem;margin-bottom:1rem}
label.fl{display:block;font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.25rem;margin-top:.7rem}
input[type=text],input[type=date],textarea,select{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .6rem;font-family:var(--mono);font-size:.8rem;width:100%;outline:none}
input:focus,textarea:focus,select:focus{border-color:var(--rust)}
textarea{resize:vertical;min-height:100px}
.form-row{display:flex;gap:.5rem}.form-row>*{flex:1}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-family:var(--serif)}
.ack-row{padding:.3rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem;display:flex;justify-content:space-between}
.pending-row{padding:.3rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem;display:flex;justify-content:space-between;align-items:center}
.ev-row{padding:.4rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="hdr">
<h1><span>Covenant</span></h1>
<div style="display:flex;gap:.3rem">
<button class="btn btn-p" onclick="showNewPolicy()">+ Policy</button>
<button class="btn btn-p" onclick="showNewMember()">+ Member</button>
</div>
</div>
<div class="main"><div id="upgrade-banner" style="display:none;background:#241e18;border:1px solid #8b3d1a;border-left:3px solid #c45d2c;padding:.6rem 1rem;font-size:.78rem;color:#bfb5a3;margin-bottom:.8rem"><strong style="color:#f0e6d3">Free tier</strong> — 10 items max. <a href="https://stockyard.dev/covenant/" target="_blank" style="color:#e8753a">Upgrade to Pro →</a></div>
<div class="overview" id="overview"></div>
<div class="tabs">
<div class="tab active" data-tab="policies" onclick="switchTab('policies')">Policies</div>
<div class="tab" data-tab="members" onclick="switchTab('members')">Members</div>
</div>
<div id="pane-policies"></div>
<div id="pane-members" style="display:none"></div>
</div>
<div id="modal"></div>

<script>
let policies=[],members=[];

async function api(url,opts){const r=await fetch(url,opts);return r.json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}
function timeAgo(d){if(!d)return'';const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago'}

async function init(){
  const[pd,md,sd]=await Promise.all([api('/api/policies'),api('/api/members'),api('/api/stats')]);
  policies=pd.policies||[];members=md.members||[];
  const compCls=sd.overall_compliance>=90?'good':sd.overall_compliance>=70?'warn':'bad';
  document.getElementById('overview').innerHTML=
    '<div class="stat"><b class="compliance-big '+compCls+'">'+sd.overall_compliance.toFixed(0)+'%</b>Overall Compliance</div>'+
    '<div class="stat"><b>'+sd.active+'</b>Active Policies</div>'+
    '<div class="stat"><b>'+sd.members+'</b>Team Members</div>'+
    '<div class="stat"><b>'+sd.acknowledgments+'</b>Acknowledgments</div>'+
    '<div class="stat"><b>'+sd.evidence+'</b>Evidence Items</div>';
  renderPolicies();renderMembers();
}

function renderPolicies(){
  const el=document.getElementById('pane-policies');
  if(!policies.length){el.innerHTML='<div class="empty">No policies yet. Create one to start tracking compliance.</div>';return}
  el.innerHTML=policies.map(p=>{
    const pct=p.compliance_pct.toFixed(0);
    const fillCls=p.compliance_pct>=90?'pf-good':p.compliance_pct>=70?'pf-warn':'pf-bad';
    return '<div class="policy-card" onclick="showPolicy(\''+p.id+'\')">'+
      '<div class="policy-top">'+
        '<span class="policy-status ps-'+p.status+'">'+p.status+'</span>'+
        (p.category?'<span class="policy-cat">'+esc(p.category)+'</span>':'')+
        '<span class="policy-title">'+esc(p.title)+'</span>'+
        '<span style="font-size:.8rem;font-weight:600;color:'+(p.compliance_pct>=90?'var(--green)':p.compliance_pct>=70?'var(--gold)':'var(--red)')+'">'+pct+'%</span>'+
      '</div>'+
      '<div class="policy-bar"><div class="policy-fill '+fillCls+'" style="width:'+pct+'%"></div></div>'+
      '<div class="policy-meta">'+
        '<span>v'+p.version+'</span>'+
        '<span>'+p.ack_count+'/'+p.member_count+' acknowledged</span>'+
        (p.owner?'<span>Owner: '+esc(p.owner)+'</span>':'')+
        '<span>'+p.evidence_count+' evidence</span>'+
        '<span>'+timeAgo(p.updated_at)+'</span>'+
      '</div></div>'
  }).join('')
}

async function showPolicy(id){
  const[p,acks,pending,ev,vers]=await Promise.all([api('/api/policies/'+id),api('/api/policies/'+id+'/acks'),api('/api/policies/'+id+'/pending'),api('/api/policies/'+id+'/evidence'),api('/api/policies/'+id+'/versions')]);
  const ackHTML=(acks.acknowledgments||[]).map(a=>'<div class="ack-row"><span style="color:var(--green)">✓ '+esc(a.member_name)+'</span><span style="color:var(--cm)">v'+a.policy_version+' · '+timeAgo(a.acked_at)+'</span></div>').join('');
  const pendHTML=(pending.pending||[]).map(m=>'<div class="pending-row"><span style="color:var(--gold)">⏳ '+esc(m.name)+'</span><button class="btn btn-s" style="font-size:.55rem;padding:.1rem .3rem" onclick="ack(\''+id+'\',\''+m.id+'\')">Acknowledge</button></div>').join('');
  const evHTML=(ev.evidence||[]).map(e=>'<div class="ev-row"><b>'+esc(e.title)+'</b> <span style="color:var(--cm)">'+esc(e.description)+' · by '+esc(e.collected_by)+' · '+timeAgo(e.collected_at)+'</span></div>').join('');

  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<div style="display:flex;justify-content:space-between;align-items:flex-start">'+
      '<h2>'+esc(p.title)+' <span class="policy-status ps-'+p.status+'" style="vertical-align:middle">'+p.status+'</span></h2>'+
      '<div style="display:flex;gap:.3rem"><button class="btn btn-p" onclick="editPolicy(\''+id+'\')">Edit</button><button class="btn btn-d" onclick="if(confirm(\'Delete?\'))delPolicy(\''+id+'\')">Del</button></div>'+
    '</div>'+
    '<div style="font-size:.7rem;color:var(--leather);margin:.3rem 0;display:flex;gap:1rem;flex-wrap:wrap">'+
      (p.category?'<span>'+esc(p.category)+'</span>':'')+
      '<span>v'+p.version+'</span>'+
      (p.owner?'<span>Owner: '+esc(p.owner)+'</span>':'')+
      '<span>Compliance: <b style="color:'+(p.compliance_pct>=90?'var(--green)':'var(--gold)')+'">'+p.compliance_pct.toFixed(0)+'%</b></span>'+
    '</div>'+
    '<div style="padding:.7rem;background:var(--bg);border:1px solid var(--bg3);font-size:.78rem;color:var(--cd);white-space:pre-wrap;margin:.5rem 0;max-height:200px;overflow-y:auto">'+esc(p.body)+'</div>'+
    '<div style="font-size:.7rem;color:var(--leather);margin:.8rem 0 .3rem">Acknowledged ('+((acks.acknowledgments||[]).length)+')</div>'+
    (ackHTML||'<div style="color:var(--cm);font-size:.72rem">No acknowledgments yet.</div>')+
    '<div style="font-size:.7rem;color:var(--leather);margin:.8rem 0 .3rem">Pending ('+(( pending.pending||[]).length)+')</div>'+
    (pendHTML||'<div style="color:var(--green);font-size:.72rem">Everyone has acknowledged!</div>')+
    '<div style="font-size:.7rem;color:var(--leather);margin:.8rem 0 .3rem">Evidence ('+((ev.evidence||[]).length)+')</div>'+
    (evHTML||'<div style="color:var(--cm);font-size:.72rem">No evidence collected yet.</div>')+
    '<button class="btn btn-p" style="margin-top:.3rem;font-size:.6rem" onclick="showAddEvidence(\''+id+'\')">+ Evidence</button>'+
    '<div style="font-size:.7rem;color:var(--leather);margin:.8rem 0 .3rem">Version History</div>'+
    (vers.versions||[]).map(v=>'<div style="padding:.2rem 0;border-bottom:1px solid var(--bg3);font-size:.68rem">v'+v.version+' <span style="color:var(--cm)">by '+esc(v.changed_by||'-')+' · '+timeAgo(v.created_at)+'</span></div>').join('')+
  '</div></div>'
}

async function ack(policyID,memberID){
  await api('/api/acknowledge',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({policy_id:policyID,member_id:memberID})});
  showPolicy(policyID);init()
}

async function delPolicy(id){await api('/api/policies/'+id,{method:'DELETE'});closeModal();init()}

function renderMembers(){
  const el=document.getElementById('pane-members');
  if(!members.length){el.innerHTML='<div class="empty">No team members yet.</div>';return}
  el.innerHTML=members.map(m=>
    '<div class="member-row"><span class="member-name">'+esc(m.name)+'</span>'+
    (m.email?'<span style="color:var(--cm);font-size:.68rem">'+esc(m.email)+'</span>':'')+
    (m.department?'<span class="member-dept">'+esc(m.department)+'</span>':'')+
    '<span style="color:var(--green);font-size:.65rem">'+m.ack_count+' acked</span>'+
    (m.pending_ack?'<span style="color:var(--gold);font-size:.65rem">'+m.pending_ack+' pending</span>':'')+
    '<span style="cursor:pointer;font-size:.6rem;color:var(--cm)" onclick="delMember(\''+m.id+'\')">del</span></div>'
  ).join('')
}

function switchTab(tab){
  document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.dataset.tab===tab));
  document.getElementById('pane-policies').style.display=tab==='policies'?'':'none';
  document.getElementById('pane-members').style.display=tab==='members'?'':'none';
}

function showNewPolicy(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>New Policy</h2>'+
    '<label class="fl">Title</label><input type="text" id="np-title" placeholder="Data Retention Policy">'+
    '<label class="fl">Category</label><input type="text" id="np-cat" placeholder="Security">'+
    '<label class="fl">Policy Content</label><textarea id="np-body" rows="6" placeholder="Write the policy in markdown..."></textarea>'+
    '<div class="form-row">'+
      '<div><label class="fl">Owner</label><input type="text" id="np-owner"></div>'+
      '<div><label class="fl">Status</label><select id="np-status"><option value="draft">Draft</option><option value="active">Active</option></select></div>'+
    '</div>'+
    '<label class="fl">Acknowledgment Deadline</label><input type="date" id="np-deadline">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveNewPolicy()">Create</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div>'+
  '</div></div>'
}

async function saveNewPolicy(){
  const body={title:document.getElementById('np-title').value,body:document.getElementById('np-body').value,category:document.getElementById('np-cat').value,status:document.getElementById('np-status').value,owner:document.getElementById('np-owner').value,ack_deadline:document.getElementById('np-deadline').value};
  if(!body.title){alert('Title required');return}
  await api('/api/policies',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});closeModal();init()
}

function editPolicy(id){
  const p=policies.find(x=>x.id===id);if(!p)return;
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Edit Policy (will create v'+(p.version+1)+' if content changes)</h2>'+
    '<label class="fl">Title</label><input type="text" id="ep-title" value="'+esc(p.title)+'">'+
    '<label class="fl">Category</label><input type="text" id="ep-cat" value="'+esc(p.category)+'">'+
    '<label class="fl">Policy Content</label><textarea id="ep-body" rows="6">'+esc(p.body)+'</textarea>'+
    '<div class="form-row">'+
      '<div><label class="fl">Owner</label><input type="text" id="ep-owner" value="'+esc(p.owner)+'"></div>'+
      '<div><label class="fl">Status</label><select id="ep-status"><option value="draft"'+(p.status==='draft'?' selected':'')+'>Draft</option><option value="active"'+(p.status==='active'?' selected':'')+'>Active</option><option value="retired"'+(p.status==='retired'?' selected':'')+'>Retired</option></select></div>'+
    '</div>'+
    '<label class="fl">Your name (for audit)</label><input type="text" id="ep-actor">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveEditPolicy(\''+id+'\')">Save</button><button class="btn btn-d" onclick="showPolicy(\''+id+'\')">Cancel</button></div>'+
  '</div></div>'
}

async function saveEditPolicy(id){
  const body={title:document.getElementById('ep-title').value,body:document.getElementById('ep-body').value,category:document.getElementById('ep-cat').value,status:document.getElementById('ep-status').value,owner:document.getElementById('ep-owner').value,actor:document.getElementById('ep-actor').value};
  await api('/api/policies/'+id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});closeModal();init()
}

function showNewMember(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Add Team Member</h2>'+
    '<label class="fl">Name</label><input type="text" id="nm-name">'+
    '<label class="fl">Email</label><input type="text" id="nm-email">'+
    '<label class="fl">Department</label><input type="text" id="nm-dept">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveNewMember()">Add</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div>'+
  '</div></div>'
}

async function saveNewMember(){
  const body={name:document.getElementById('nm-name').value,email:document.getElementById('nm-email').value,department:document.getElementById('nm-dept').value};
  if(!body.name){alert('Name required');return}
  await api('/api/members',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});closeModal();init()
}

async function delMember(id){if(!confirm('Remove member?'))return;await api('/api/members/'+id,{method:'DELETE'});init()}

function showAddEvidence(policyID){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Add Evidence</h2>'+
    '<label class="fl">Title</label><input type="text" id="ae-title" placeholder="Screenshot of MFA enforcement">'+
    '<label class="fl">Description</label><textarea id="ae-desc" rows="2" placeholder="Evidence details..."></textarea>'+
    '<label class="fl">Collected By</label><input type="text" id="ae-by">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveEvidence(\''+policyID+'\')">Add</button><button class="btn btn-d" onclick="showPolicy(\''+policyID+'\')">Cancel</button></div>'+
  '</div></div>'
}

async function saveEvidence(policyID){
  const body={title:document.getElementById('ae-title').value,description:document.getElementById('ae-desc').value,collected_by:document.getElementById('ae-by').value};
  if(!body.title){alert('Title required');return}
  await api('/api/policies/'+policyID+'/evidence',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  showPolicy(policyID);init()
}

function closeModal(){document.getElementById('modal').innerHTML=''}
init();
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script></body></html>`
