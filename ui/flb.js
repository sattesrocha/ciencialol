// =====================================================================
//  FunLeagueBuilds (funleaguebuilds.de) — adaptador + exibição
//  © das guias: FunLeagueBuilds e seus autores. Mostramos só runas,
//  itens, ordem de habilidades e feitiços, com crédito e link pro site.
//  Enquanto a autorização não chega, usamos GUIAS DE EXEMPLO inventadas.
// =====================================================================
"use strict";
(function(){
const FLB_SITE = "https://funleaguebuilds.de";
const LANE2ROLE = { top:"top", jungle:"jungle", mid:"mid", middle:"mid", adc:"adc", bot:"adc", bottom:"adc", support:"sup", sup:"sup" };

// ---------- guias de EXEMPLO (inventadas, no mesmo formato da API) ----------
const it = n => ({name:n}); const sh = n => ({name:n});
const page = (pt,ks,pp,st,sp,shards) => ({ name:"Rune Page 1", primary:{tree:pt, keystone:{name:ks}, perks:pp.map(it)}, secondary:{tree:st, perks:sp.map(it)}, shards:shards.map(sh) });
const set = (start,boots,core,sit) => ({ name:"Core Build", starting:start.map(it), boots:boots.map(it), core:core.map(it), situational:sit.map(it) });
const SO = s => s.split("");
const DEMO = [
  { id:"demo-01", title:"Malphite AP de um botão só", champion:"Malphite", lanes:["top"], patch:"26.20", authorName:"Ciência LOL", approved:true, stats:{score:41},
    visual:{ summonerSpells:[it("Flash"),it("Teleport")], skillOrder:SO("QEQWQRQEQERE EWWRWW".replace(" ","")),
      runePages:[page("Sorcery","Arcane Comet",["Manaflow Band","Transcendence","Scorch"],"Inspiration",["Magical Footwear","Biscuit Delivery"],["StatModsAdaptiveForceIcon","StatModsAdaptiveForceIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Doran's Ring","Health Potion"],["Sorcerer's Shoes"],["Luden's Companion","Stormsurge","Rabadon's Deathcap"],["Void Staff","Zhonya's Hourglass","Shadowflame"])] } },
  { id:"demo-02", title:"Shaco tanque das caixinhas", champion:"Shaco", lanes:["jungle"], patch:"26.20", authorName:"Ciência LOL", approved:false, stats:{score:17},
    visual:{ summonerSpells:[it("Smite"),it("Flash")], skillOrder:SO("WEWQWRWEWEREEQQRQQ"),
      runePages:[page("Resolve","Grasp of the Undying",["Demolish","Bone Plating","Overgrowth"],"Inspiration",["Cosmic Insight","Approach Velocity"],["StatModsAttackSpeedIcon","StatModsHealthPlusIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Gustwalker Hatchling","Health Potion"],["Plated Steelcaps"],["Liandry's Torment","Heartsteel","Jak'Sho, The Protean"],["Thornmail","Spirit Visage","Kaenic Rookern"])] } },
  { id:"demo-03", title:"Galio de crítico (sim, crítico)", champion:"Galio", lanes:["mid"], patch:"26.19", authorName:"Ciência LOL", approved:false, stats:{score:9},
    visual:{ summonerSpells:[it("Flash"),it("Ignite")], skillOrder:SO("QEQWQRQEQERE EWWRWW".replace(" ","")),
      runePages:[page("Precision","Lethal Tempo",["Presence of Mind","Legend: Alacrity","Cut Down"],"Domination",["Sudden Impact","Treasure Hunter"],["StatModsAttackSpeedIcon","StatModsAdaptiveForceIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Doran's Blade","Health Potion"],["Berserker's Greaves"],["Kraken Slayer","Infinity Edge","Phantom Dancer"],["Lord Dominik's Regards","Guardian Angel","Bloodthirster"])] } },
  { id:"demo-04", title:"Ezreal todo de vida", champion:"Ezreal", lanes:["adc"], patch:"26.20", authorName:"Ciência LOL", approved:true, stats:{score:33},
    visual:{ summonerSpells:[it("Flash"),it("Barrier")], skillOrder:SO("QWEQQRQWQWRWWEEREE"),
      runePages:[page("Resolve","Grasp of the Undying",["Shield Bash","Conditioning","Overgrowth"],"Sorcery",["Manaflow Band","Transcendence"],["StatModsCDRScalingIcon","StatModsHealthPlusIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Doran's Shield","Health Potion"],["Ionian Boots of Lucidity"],["Tear of the Goddess","Iceborn Gauntlet","Heartsteel"],["Frozen Heart","Spirit Visage","Warmog's Armor"])] } },
  { id:"demo-05", title:"Yuumi encantadora de AP", champion:"Yuumi", lanes:["support"], patch:"26.18", authorName:"Ciência LOL", approved:false, stats:{score:12},
    visual:{ summonerSpells:[it("Ignite"),it("Exhaust")], skillOrder:SO("QEQWQRQEQERE EWWRWW".replace(" ","")),
      runePages:[page("Sorcery","Arcane Comet",["Axiom Arcanist","Transcendence","Scorch"],"Domination",["Cheap Shot","Ultimate Hunter"],["StatModsAdaptiveForceIcon","StatModsAdaptiveForceIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["World Atlas","Health Potion"],["Sorcerer's Shoes"],["Malignance","Shadowflame","Rabadon's Deathcap"],["Void Staff","Horizon Focus"])] } },
  { id:"demo-06", title:"Darius & Thresh: a dupla do gancho", champion:"Darius", lanes:["adc","support"], patch:"26.20", authorName:"Ciência LOL", approved:true, stats:{score:58},
    guideType:"duo", selectedChampions:["Darius","Thresh"],
    visual:{ champLanes:{Darius:"adc", Thresh:"support"},
      champBuilds:{
        Darius:{ summonerSpells:[it("Flash"),it("Ghost")], skillOrder:SO("QEQWQRQEQERE EWWRWW".replace(" ","")),
          runePages:[page("Precision","Conqueror",["Triumph","Legend: Haste","Last Stand"],"Resolve",["Bone Plating","Unflinching"],["StatModsAttackSpeedIcon","StatModsAdaptiveForceIcon","StatModsHealthScalingIcon"])],
          itemSets:[set(["Doran's Blade","Health Potion"],["Plated Steelcaps"],["Trinity Force","Sterak's Gage","Dead Man's Plate"],["Death's Dance","Force of Nature","Spirit Visage"])] },
        Thresh:{ summonerSpells:[it("Flash"),it("Ignite")], skillOrder:SO("QEWQQRQEQEREEWWRWW"),
          runePages:[page("Resolve","Aftershock",["Font of Life","Bone Plating","Unflinching"],"Inspiration",["Hextech Flashtraption","Cosmic Insight"],["StatModsCDRScalingIcon","StatModsHealthPlusIcon","StatModsHealthScalingIcon"])],
          itemSets:[set(["World Atlas","Health Potion"],["Mercury's Treads"],["Locket of the Iron Solari","Zeke's Convergence","Knight's Vow"],["Redemption","Mikael's Blessing","Thornmail"])] } } } },
  { id:"demo-07", title:"Nasus da velocidade de ataque", champion:"Nasus", lanes:["top","jungle"], patch:"26.19", authorName:"Ciência LOL", approved:false, stats:{score:21},
    visual:{ summonerSpells:[it("Flash"),it("Ghost")], skillOrder:SO("QEWQQRQEQEREEWWRWW"),
      runePages:[page("Precision","Lethal Tempo",["Triumph","Legend: Alacrity","Last Stand"],"Resolve",["Second Wind","Overgrowth"],["StatModsAttackSpeedIcon","StatModsAdaptiveForceIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Doran's Blade","Health Potion"],["Berserker's Greaves"],["Blade of The Ruined King","Guinsoo's Rageblade","Wit's End"],["Terminus","Jak'Sho, The Protean","Sterak's Gage"])] } },
  { id:"demo-08", title:"Lux tanque de Abismo Uivante", champion:"Lux", lanes:["aram"], patch:"26.20", authorName:"Ciência LOL", approved:false, stats:{score:5},
    visual:{ summonerSpells:[it("Flash"),it("Mark")], skillOrder:SO("EQEWERE QEQRQQWWRWW".replace(" ","")),
      runePages:[page("Resolve","Guardian",["Font of Life","Conditioning","Revitalize"],"Sorcery",["Manaflow Band","Transcendence"],["StatModsCDRScalingIcon","StatModsHealthPlusIcon","StatModsHealthScalingIcon"])],
      itemSets:[set(["Health Potion"],["Ionian Boots of Lucidity"],["Sunfire Aegis","Heartsteel","Kaenic Rookern"],["Thornmail","Randuin's Omen"])] } },
];

// ---------- adaptador: guia da API -> ciência do app ----------
// o site às vezes usa o nome ("Lee Sin") e às vezes o id ("LeeSin") como chave
const ckey = s => String(s||"").toLowerCase().replace(/[^a-z0-9]/g,"");
const pickBuild = (v, champ) => { const cb=v.champBuilds; if(!cb) return v;
  if(cb[champ]) return cb[champ]; const k=Object.keys(cb).find(x=>ckey(x)===ckey(champ)); return k?cb[k]:v; };
// fragmentos podem vir como ícone ("StatModsAttackSpeedIcon") ou como texto ("Attack Speed")
const SHARD_TXT = { adaptiveforce:"StatModsAdaptiveForceIcon", attackspeed:"StatModsAttackSpeedIcon", abilityhaste:"StatModsCDRScalingIcon",
  movespeed:"StatModsMovementSpeedIcon", movementspeed:"StatModsMovementSpeedIcon", health:"StatModsHealthPlusIcon", healthplus:"StatModsHealthPlusIcon",
  healthscaling:"StatModsHealthScalingIcon", scalinghealth:"StatModsHealthScalingIcon", tenacityandslowresist:"StatModsTenacityIcon", tenacity:"StatModsTenacityIcon",
  armor:"StatModsArmorIcon", magicresist:"StatModsMagicResIcon", magicresistance:"StatModsMagicResIcon" };
const shardId = s => /^statmods/i.test(s) ? s : (SHARD_TXT[ckey(s)] || s);
const names = arr => (Array.isArray(arr)?arr:[]).map(x=>x&&x.name).filter(Boolean).map(s=>String(s).slice(0,60));
function sanitizeBuild(b){
  b=b||{};
  const page = rp => { rp=rp||{}; const pr=rp.primary||{}, se=rp.secondary||{};
    return { primary:{tree:String(pr.tree||"").slice(0,20), keystone:String((pr.keystone||{}).name||"").slice(0,40), perks:names(pr.perks).slice(0,3)},
             secondary:{tree:String(se.tree||"").slice(0,20), perks:names(se.perks).slice(0,2)}, shards:names(rp.shards).slice(0,3).map(shardId) }; };
  const pages=(Array.isArray(b.runePages)?b.runePages:[]).slice(0,3).map(page).filter(p=>p.primary.keystone);
  // nomes dos conjuntos são texto do autor: trocamos por rótulos neutros
  const sets=(Array.isArray(b.itemSets)?b.itemSets:[]).slice(0,3).map((s,i)=>({ name: i?`Alternativa ${i}`:"Build principal",
    starting:names(s.starting).slice(0,6), boots:names(s.boots).slice(0,4), core:names(s.core).slice(0,8), situational:names(s.situational).slice(0,12) }));
  return {
    spells: names(b.summonerSpells).slice(0,2),
    skills: (Array.isArray(b.skillOrder)?b.skillOrder:[]).slice(0,18).map(x=>String(x).toUpperCase()).filter(x=>"QWER".includes(x)&&x.length===1),
    runes: pages[0] || page({}), runePages: pages,
    items: sets,
  };
}
function fromGuide(g, demo){
  if(!g || typeof g!=="object" || !g.id) return null;
  const v=g.visual||{};
  const sel=(Array.isArray(g.selectedChampions)?g.selectedChampions:[]).filter(Boolean);
  const isDuo = (g.guideType==="duo" || (v.guideType==="duo")) && sel.length===2;
  const lanes=(Array.isArray(g.lanes)?g.lanes:[]).map(x=>String(x).toLowerCase());
  let roles=[...new Set(lanes.map(l=>LANE2ROLE[l]).filter(Boolean))];
  const champ = String(isDuo ? sel[0] : (g.champion||sel[0]||"")).slice(0,30);
  const builds={}; (isDuo?sel:[champ]).forEach(c=>builds[c]=sanitizeBuild(pickBuild(v,c)));
  let duo=false, champ2, c1=champ;
  if(isDuo){
    const cl=v.champLanes||{}; const supC = sel.find(c=>/sup/i.test(cl[c]||"")); const adcC = sel.find(c=>/adc|bot/i.test(cl[c]||""));
    if(supC){ c1 = sel.find(c=>c!==supC); champ2=supC; duo=true; }
    else if(adcC){ c1=adcC; champ2=sel.find(c=>c!==adcC); duo=true; }
    else { c1=sel[0]; champ2=sel[1]; duo = roles.includes("sup") || roles.includes("adc"); }
  }
  const main=builds[c1]||builds[champ]||{};
  const extraModes = lanes.filter(l=>l==="aram"||l==="arena").map(l=>l.toUpperCase());
  return {
    id:"flb-"+String(g.id).replace(/[^\w-]/g,"").slice(0,40), src:"flb", demo:!!demo,
    nome:String(g.title||"Guia sem título").slice(0,120), champ:c1, champ2: duo?champ2:undefined, duo,
    roles: duo ? ["duo"] : roles, modes: extraModes,
    runas: main.runes ? [main.runes.primary.keystone, main.runes.secondary.tree].filter(Boolean).join(" + ") : "—",
    build: main.items && main.items[0] ? [...main.items[0].boots.slice(0,1), ...main.items[0].core] : [],
    ideia: "", posts: [],
    flb:{ url: demo ? FLB_SITE+"/" : FLB_SITE+"/guide.html?id="+encodeURIComponent(g.id), author:String(g.authorName||"").slice(0,40),
          patch:String(g.patch||"").slice(0,10), approved:!!g.approved, score:+((g.stats||{}).score)||0, builds, order: duo?[c1,champ2]:[c1] },
  };
}
window.FLB = { DEMO, fromGuide, sanitizeBuild, SITE:FLB_SITE };
window.FLB_BASE = DEMO.map(g=>fromGuide(g,true)).filter(Boolean);

// =====================================================================
//  Nomes e ícones oficiais (Data Dragon), traduzidos pra pt-BR
// =====================================================================
const DDX = { ready:false, loading:null, ver:"", items:{}, runes:{}, champs:{} };
const norm = s => String(s||"").toLowerCase().replace(/[’'`.:,!?-]/g,"").replace(/\s+/g," ").trim();
const TREE = { precision:[8000,"7201_Precision","Precisão"], domination:[8100,"7200_Domination","Dominação"], sorcery:[8200,"7202_Sorcery","Feitiçaria"],
               inspiration:[8300,"7203_Whimsy","Inspiração"], resolve:[8400,"7204_Resolve","Determinação"] };
const SHARD = { statmodsadaptiveforceicon:"Força Adaptável", statmodsattackspeedicon:"Velocidade de Ataque", statmodscdrscalingicon:"Aceleração de Habilidade",
  statmodsmovementspeedicon:"Velocidade de Movimento", statmodshealthplusicon:"Vida", statmodshealthscalingicon:"Vida (por nível)",
  statmodstenacityicon:"Tenacidade e Resist. a Lentidão", statmodsarmoricon:"Armadura", statmodsmagicresicon:"Resistência Mágica" };
const SPELL = { flash:["SummonerFlash","Flash"], ignite:["SummonerDot","Incendiar"], barrier:["SummonerBarrier","Barreira"], exhaust:["SummonerExhaust","Exaustão"],
  heal:["SummonerHeal","Curar"], teleport:["SummonerTeleport","Teleporte"], smite:["SummonerSmite","Golpear"], ghost:["SummonerHaste","Fantasma"],
  cleanse:["SummonerBoost","Purificar"], clarity:["SummonerMana","Clareza"], mark:["SummonerSnowball","Marcar"], snowball:["SummonerSnowball","Marcar"] };
async function j(url){ const r=await fetch(url); if(!r.ok) throw new Error(r.status); return r.json(); }
function loadDD(){
  if(DDX.loading) return DDX.loading;
  DDX.loading = (async()=>{
    try{
      const vs=await j("https://ddragon.leagueoflegends.com/api/versions.json"); const v=DDX.ver=vs[0];
      const B=`https://ddragon.leagueoflegends.com/cdn/${v}/data/`;
      const [ien,ipt,ren,rpt]=await Promise.all([j(B+"en_US/item.json"),j(B+"pt_BR/item.json"),j(B+"en_US/runesReforged.json"),j(B+"pt_BR/runesReforged.json")]);
      for(const [id,o] of Object.entries(ien.data||{})){ const k=norm(o.name); if(!DDX.items[k] || +id<+DDX.items[k].id) DDX.items[k]={id, pt:(ipt.data[id]||o).name}; }
      const ptRunes={}; rpt.forEach(t=>{ ptRunes[t.id]=t.name; t.slots.forEach(s=>s.runes.forEach(r=>ptRunes[r.id]=r.name)); });
      ren.forEach(t=>{ DDX.runes[norm(t.key)]=DDX.runes[norm(t.name)]={icon:t.icon, pt:ptRunes[t.id]||t.name};
        t.slots.forEach(s=>s.runes.forEach(r=>DDX.runes[norm(r.name)]={icon:r.icon, pt:ptRunes[r.id]||r.name})); });
      DDX.tree=ren; DDX.ptR=ptRunes;
      // por número (o que vem do cliente do LoL e do op.gg)
      DDX.itemById={}; for(const [id,o] of Object.entries(ien.data||{})) DDX.itemById[id]={ pt:(ipt.data[id]||o).name, icon:`https://ddragon.leagueoflegends.com/cdn/${v}/img/item/${id}.png` };
      DDX.runeById={}; ren.forEach(t=>{ DDX.runeById[t.id]={pt:ptRunes[t.id]||t.name, icon:IMG(t.icon)}; t.slots.forEach(s=>s.runes.forEach(r=>DDX.runeById[r.id]={pt:ptRunes[r.id]||r.name, icon:IMG(r.icon)})); });
      // lista de itens compráveis (Summoner's Rift ou ARAM), um por nome
      const seen=new Set(); DDX.itemList=[];
      Object.entries(ien.data||{}).sort((a,b)=>+a[0]-+b[0]).forEach(([id,o])=>{
        const m=o.maps||{}; if(!(m["11"]||m["12"]) || !(o.gold&&o.gold.purchasable) || o.requiredChampion || o.requiredAlly || o.inStore===false || o.hideFromAll) return;
        const k=norm(o.name); if(seen.has(k)) return; seen.add(k);
        const tags=o.tags||[], gold=o.gold.total||0;
        const cat = tags.includes("Boots") ? "boots" : (gold<=500 && !(o.from||[]).length) ? "start" : (o.into||[]).length ? "comp" : "leg";
        DDX.itemList.push({ id, en:o.name, pt:(ipt.data[id]||o).name, gold, cat });
      });
      DDX.ready=true;
    }catch(e){ console.warn("Data Dragon indisponível", e); }
    document.dispatchEvent(new Event("flb-dd"));
  })();
  return DDX.loading;
}
async function champData(name){
  const id=window.cidOf ? window.cidOf(name) : name;
  if(DDX.champs[id]) return DDX.champs[id];
  await loadDD(); if(!DDX.ver) return null;
  try{ const d=await j(`https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/data/pt_BR/champion/${id}.json`); return DDX.champs[id]=d.data[id]; }catch(e){ return null; }
}
const IMG = p => `https://ddragon.leagueoflegends.com/cdn/img/${p}`;
const itemInfo = n => { const o=DDX.items[norm(n)]; return { name: o?o.pt:n, icon: o?`https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/img/item/${o.id}.png`:"" }; };
const runeInfo = n => { const o=DDX.runes[norm(n)]; return { name:o?o.pt:n, icon:o?IMG(o.icon):"" }; };
const treeInfo = n => { const t=TREE[norm(n)]; const o=DDX.runes[norm(n)]; return { name: t?t[2]:(o?o.pt:n), icon: t?IMG(`perk-images/Styles/${t[1]}.png`):"" }; };
const shardInfo = n => ({ name: SHARD[norm(n)]||n, icon: /^statmods/i.test(n)?IMG(`perk-images/StatMods/${n}.png`):"" });
const spellInfo = n => { const s=SPELL[norm(n)]; return { name:s?s[1]:n, icon: s&&DDX.ver?`https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/img/spell/${s[0]}.png`:"" }; };
const tName = n => DDX.ready ? itemInfo(n).name : n;
// ---- Importar pro LoL: traduz a build (nomes) pros números que o cliente usa ----
const SHARD_ID = { statmodsadaptiveforceicon:5008, statmodsattackspeedicon:5005, statmodscdrscalingicon:5007, statmodsmovementspeedicon:5010,
  statmodshealthplusicon:5011, statmodshealthscalingicon:5001, statmodstenacityicon:5013, statmodsarmoricon:5002, statmodsmagicresicon:5003 };
const SPELL_KEY = { flash:4, ignite:14, barrier:21, exhaust:3, heal:7, teleport:12, smite:11, ghost:6, cleanse:1, clarity:13, mark:32, snowball:32 };
async function toLCU(vb, champ, title){
  await loadDD(); const miss=[], missR=[]; const out={ champ, title:String(title||"Ciência LOL").slice(0,75) };
  if(!DDX.ready) throw new Error("não consegui carregar os dados do LoL (Data Dragon)");
  const runeId={}, treeId={};
  (DDX.tree||[]).forEach(t=>{ treeId[norm(t.key)]=treeId[norm(t.name)]=t.id; t.slots.forEach(sl=>sl.runes.forEach(r=>{ runeId[norm(r.name)]=r.id; runeId[norm(r.key)]=r.id; })); });
  const R=vb&&vb.runes;
  if(R && R.primary && R.primary.keystone){
    const tid = n => (TREE[norm(n)]||[])[0] || treeId[norm(n)];
    const rid = n => { const id=runeId[norm(n)]; if(!id) missR.push(n); return id; };
    const prim=tid(R.primary.tree), sub=tid(R.secondary.tree);
    const perks=[rid(R.primary.keystone), ...R.primary.perks.map(rid), ...R.secondary.perks.map(rid)];
    let sh=(R.shards||[]).map(x=>SHARD_ID[norm(shardId(x))]).filter(Boolean);
    if(sh.length!==3) sh=[5008,5008,5011];
    if(!prim) missR.push(R.primary.tree||"árvore principal"); if(!sub) missR.push(R.secondary.tree||"árvore secundária");
    if(perks.length!==6) missR.push("runa faltando na build");
    if(prim && sub && perks.length===6 && perks.every(Boolean)) out.page={ primaryStyleId:prim, subStyleId:sub, selectedPerkIds:[...perks, ...sh] };
  }
  const sp=(vb&&vb.spells||[]).map(n=>SPELL_KEY[norm(n)]).filter(Boolean);
  if(sp.length===2) out.spells=sp;
  const set=vb&&vb.items&&vb.items[0];
  const cd=await champData(champ); out.champKey = cd ? +cd.key : 0;
  if(set){
    const ids = arr => (arr||[]).map(n=>{ const o=DDX.items[norm(n)]; if(!o) miss.push(n); return o&&o.id; }).filter(Boolean);
    out.blocks=[["Inicial",set.starting],["Botas",set.boots],["Build principal",set.core],["Situacional",set.situational]]
      .map(([type,a])=>({type, items:ids(a)})).filter(b=>b.items.length);
  }
  out.missing=miss; out.missingRunes=missR; out.hasRunes=!!(R && R.primary && R.primary.keystone); out.hasSpells=!!(vb&&vb.spells&&vb.spells.length);
  if(!out.champKey) out.noChamp=true; return out;
}
// ---- por número (cliente do LoL / op.gg) ----
const SHARD_BY_ID = { 5001:"StatModsHealthScalingIcon", 5002:"StatModsArmorIcon", 5003:"StatModsMagicResIcon", 5005:"StatModsAttackSpeedIcon",
  5007:"StatModsCDRScalingIcon", 5008:"StatModsAdaptiveForceIcon", 5010:"StatModsMovementSpeedIcon", 5011:"StatModsHealthPlusIcon", 5013:"StatModsTenacityIcon" };
const SPELL_BY_ID = {}; Object.entries(SPELL_KEY).forEach(([n,k])=>{ if(!SPELL_BY_ID[k]) SPELL_BY_ID[k]=n; });
const itemIdInfo = id => { const o=(DDX.itemById||{})[id]; return { name:o?o.pt:"Item "+id, icon:o?o.icon:"" }; };
const runeIdInfo = id => { if(SHARD_BY_ID[id]) return shardInfo(SHARD_BY_ID[id]); const o=(DDX.runeById||{})[id]; return { name:o?o.pt:"Runa "+id, icon:o?o.icon:"" }; };
const spellIdInfo = id => SPELL_BY_ID[id] ? spellInfo(SPELL_BY_ID[id]) : { name:"Feitiço "+id, icon:"" };
window.FLBDD = { toLCU, itemIdInfo, runeIdInfo, spellIdInfo, loadDD, champData, itemInfo, runeInfo, treeInfo, shardInfo, spellInfo, tName, get ready(){return DDX.ready;}, get ver(){return DDX.ver;} };

// =====================================================================
//  Desenho do detalhe (runas, itens, habilidades, feitiços)
// =====================================================================
const e = s => String(s??"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
const ico = (o,cls="") => `<span class="fi ${cls}" title="${e(o.name)}">${o.icon?`<img src="${e(o.icon)}" alt="" onerror="this.remove()">`:""}<b>${e((o.name||"?")[0])}</b></span>`;
function runesHtml(r){
  if(!r || !r.primary || !r.primary.keystone) return `<div class="fnone">Runas ainda não preenchidas.</div>`;
  const p=r.primary, s=r.secondary;
  return `<div class="frunes">
    <div class="ftree"><div class="fth">${ico(treeInfo(p.tree),"t")}<span>${e(treeInfo(p.tree).name)}</span></div>
      <div class="fks">${ico(runeInfo(p.keystone),"k")}<div><small>Runa principal</small><b>${e(runeInfo(p.keystone).name)}</b></div></div>
      ${p.perks.map(x=>`<div class="frow">${ico(runeInfo(x))}<span>${e(runeInfo(x).name)}</span></div>`).join("")}</div>
    <div class="ftree sec"><div class="fth">${ico(treeInfo(s.tree),"t")}<span>${e(treeInfo(s.tree).name)}</span></div>
      ${s.perks.map(x=>`<div class="frow">${ico(runeInfo(x))}<span>${e(runeInfo(x).name)}</span></div>`).join("")}
      ${r.shards.length?`<div class="fshards">${r.shards.map(x=>ico(shardInfo(x),"s")).join("")}</div><div class="fshn">${r.shards.map(x=>e(shardInfo(x).name)).join(" · ")}</div>`:""}</div>
  </div>`;
}
function itemsHtml(sets){
  if(!sets || !sets.length) return `<div class="fnone">Itens ainda não preenchidos.</div>`;
  const grp=(t,arr,arrow)=>arr&&arr.length?`<div class="fgrp"><h6>${t}</h6><div class="fitems">${arr.map((n,i)=>`${i&&arrow?'<i class="farr">›</i>':""}<div class="fit">${ico(itemInfo(n),"i")}<span>${e(itemInfo(n).name)}</span></div>`).join("")}</div></div>`:"";
  return sets.map(s=>`<div class="fset">${sets.length>1?`<div class="fsetn">${e(s.name)}</div>`:""}
    ${grp("Itens iniciais",s.starting)}${grp(s.boots.length>1?"Botas (opções)":"Botas",s.boots)}${grp("Build principal (na ordem)",s.core,true)}${grp("Situacionais",s.situational)}</div>`).join("");
}
function skillsHtml(order, cd){
  if(!order || !order.length) return `<div class="fnone">Ordem de habilidades ainda não preenchida.</div>`;
  const sp = cd && cd.spells ? cd.spells : null;
  const row = (k,i)=>{ const s=sp&&sp[i]; const icon = s&&DDX.ver?`https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/img/spell/${s.image.full}`:"";
    return `<tr><th>${ico({name:s?s.name:k, icon},"sk")}<span class="kk">${k}</span></th>${order.map((x,l)=>`<td class="${x===k?"on k"+k:""}">${x===k?l+1:""}</td>`).join("")}</tr>`; };
  // prioridade: ordem em que cada habilidade (fora o R) chega ao máximo
  const cnt={Q:0,W:0,E:0}, maxAt={}; order.forEach((k,i)=>{ if(cnt[k]!=null){ cnt[k]++; if(cnt[k]===5 && !maxAt[k]) maxAt[k]=i; } });
  const pri=Object.keys(maxAt).sort((a,b)=>maxAt[a]-maxAt[b]);
  return `${pri.length?`<div class="fpri">Prioridade: ${pri.map(k=>`<b class="k${k}">${k}</b>`).join(" › ")}</div>`:""}
    <div class="fskw"><table class="fsk"><tr><th></th>${order.map((x,i)=>`<td class="lv">${i+1}</td>`).join("")}</tr>${["Q","W","E","R"].map(row).join("")}</table></div>`;
}
function spellsHtml(sp){ return sp&&sp.length ? `<div class="fspells">${sp.map(n=>`<div class="fsp">${ico(spellInfo(n),"sp")}<span>${e(spellInfo(n).name)}</span></div>`).join("")}</div>` : ""; }
async function renderBuild(target, c, champ){ return renderBuildObj(target, c.flb.builds[champ], champ); }
async function renderBuildObj(target, b, champ){
  if(!b){ target.innerHTML=""; return; }
  const draw=(cd)=>{ target.innerHTML=`
    <div class="fcols"><section class="fsec"><h5>Runas</h5>${(b.runePages&&b.runePages.length>1)
        ? b.runePages.map((p,i)=>`<div class="fsetn">${i?`Alternativa ${i}`:"Página principal"}</div>${runesHtml(p)}`).join("")
        : runesHtml(b.runes)}</section>
      <section class="fsec"><h5>Feitiços</h5>${spellsHtml(b.spells)||'<div class="fnone">—</div>'}<h5 style="margin-top:18px">Ordem de habilidades</h5>${skillsHtml(b.skills, cd)}</section></div>
    <section class="fsec"><h5>Itens</h5>${itemsHtml(b.items)}</section>`; };
  draw(null);
  await loadDD(); const cd=await champData(champ); if(target.isConnected) draw(cd);
}
window.FLBView = { renderBuild, renderBuildObj };

// reaplica o filtro de segurança numa build já convertida (sala online, correções, ciências próprias)
function reSanitize(v){
  if(!v || typeof v!=="object") return null;
  const L = a => (Array.isArray(a)?a:[]).map(n=>({name:n}));
  const pages=(Array.isArray(v.runePages)&&v.runePages.length?v.runePages:[v.runes]).slice(0,3).map(r=>{ r=r||{}; const p=r.primary||{}, q=r.secondary||{};
    return { primary:{tree:p.tree,keystone:{name:p.keystone},perks:L(p.perks)}, secondary:{tree:q.tree,perks:L(q.perks)}, shards:L(r.shards) }; });
  return sanitizeBuild({ summonerSpells:L(v.spells), skillOrder:v.skills, runePages:pages,
    itemSets:(Array.isArray(v.items)?v.items:[]).map(x=>({ starting:L(x&&x.starting), boots:L(x&&x.boots), core:L(x&&x.core), situational:L(x&&x.situational) })) });
}
window.FLB.reSanitize = reSanitize;

// =====================================================================
//  Editor visual de build (Nova ciência / Editar / Corrigir build)
// =====================================================================
const SHARD_ROWS = [["StatModsAdaptiveForceIcon","StatModsAttackSpeedIcon","StatModsCDRScalingIcon"],
                    ["StatModsAdaptiveForceIcon","StatModsMovementSpeedIcon","StatModsHealthScalingIcon"],
                    ["StatModsHealthPlusIcon","StatModsTenacityIcon","StatModsHealthScalingIcon"]];
const SPELL_LIST = ["Flash","Ignite","Teleport","Smite","Heal","Barrier","Exhaust","Ghost","Cleanse","Clarity","Mark"];
const GROUPS = [["starting","Itens iniciais",6],["boots","Botas",4],["core","Build principal (na ordem)",8],["situational","Situacionais",12]];
const CATS = [["all","Todos"],["leg","Lendários"],["boots","Botas"],["start","Iniciais"],["comp","Componentes"]];

function buildSummary(v){
  if(!v) return null; const r=v.runes||{}; const p=r.primary||{}, s=r.secondary||{};
  const runas = p.keystone ? `${runeInfo(p.keystone).name}${s.tree?" + "+treeInfo(s.tree).name:""}` : "";
  const set=(v.items||[])[0]||{}; const build=[...(set.core||[]),...(set.situational||[])].slice(0,10).map(n=>itemInfo(n).name);
  return { runas, build };
}

function mountEditor(el, vb, getChamp){
  const b = vb ? reSanitize(vb) : null;
  const r = (b&&b.runes)||{primary:{perks:[]},secondary:{perks:[]},shards:[]}; const set=(b&&b.items&&b.items[0])||{};
  const st = { spells:[...((b&&b.spells)||[])], skills:[...((b&&b.skills)||[])], prim:r.primary.tree||"", ks:r.primary.keystone||"",
    pp:[...(r.primary.perks||[])], sec:r.secondary.tree||"", sp:[...(r.secondary.perks||[])], shards:[...(r.shards||[])],
    items:{ starting:[...(set.starting||[])], boots:[...(set.boots||[])], core:[...(set.core||[])], situational:[...(set.situational||[])] },
    grp:"core", q:"", cat:"all", cd:null };
  while(st.skills.length<18) st.skills.push("");
  el.innerHTML = `<div class="bed">
    <div class="bsec"><h5>Runas</h5><div id="beRunes"></div></div>
    <div class="bcols"><div class="bsec"><h5>Feitiços <small>(escolha 2)</small></h5><div id="beSpells"></div></div>
      <div class="bsec"><h5>Ordem de habilidades <small>(clique no nível de cada habilidade)</small></h5><div id="beSkills"></div></div></div>
    <div class="bsec"><h5>Itens</h5><div id="beItems"></div><div id="bePick"></div></div></div>`;
  const $e = s => el.querySelector(s);
  const treeOf = key => (DDX.tree||[]).find(t=>t.key===key);
  const ptR = r => (DDX.ptR&&DDX.ptR[r.id])||r.name;
  const runeBtn = (r,on,data) => `<button type="button" class="bopt ${on?"on":""}" ${data} title="${e(ptR(r))}"><img src="${e(IMG(r.icon))}" alt=""><span>${e(ptR(r))}</span></button>`;

  function drawRunes(){
    if(!DDX.tree){ $e("#beRunes").innerHTML=`<div class="fnone">Carregando runas… (precisa de internet)</div>`; return; }
    const trees = (DDX.tree||[]);
    const tsel = (cur, attr, skip) => `<div class="btrees">${trees.filter(t=>t.key!==skip).map(t=>`<button type="button" ${attr}="${t.key}" class="${cur===t.key?"on":""}"><img src="${e(IMG(t.icon))}" alt=""><span>${e(ptR(t))}</span></button>`).join("")}</div>`;
    const P=treeOf(st.prim), Sx=treeOf(st.sec);
    $e("#beRunes").innerHTML = `<div class="bru">
      <div class="btree"><div class="blab">Árvore principal</div>${tsel(st.prim,"data-prim")}
        ${P?P.slots.map((s,i)=>`<div class="brow ${i?"":"ks"}">${s.runes.map(r=>runeBtn(r, i? st.pp[i-1]===r.name : st.ks===r.name, `data-pr="${i}" data-n="${e(r.name)}"`)).join("")}</div>`).join(""):`<div class="fnone">Escolha a árvore principal.</div>`}</div>
      <div class="btree"><div class="blab">Árvore secundária <small>(2 runas)</small></div>${tsel(st.sec,"data-sec",st.prim)}
        ${Sx?Sx.slots.slice(1).map(s=>`<div class="brow">${s.runes.map(r=>runeBtn(r, st.sp.includes(r.name), `data-sr="1" data-n="${e(r.name)}"`)).join("")}</div>`).join(""):`<div class="fnone">Escolha a árvore secundária.</div>`}
        <div class="blab" style="margin-top:12px">Fragmentos</div>
        ${SHARD_ROWS.map((row,i)=>`<div class="brow sh">${row.map(id=>`<button type="button" class="bopt s ${st.shards[i]===id?"on":""}" data-sh="${i}" data-n="${id}" title="${e(SHARD[id.toLowerCase()])}"><img src="${e(IMG("perk-images/StatMods/"+id+".png"))}" alt=""><span>${e(SHARD[id.toLowerCase()])}</span></button>`).join("")}</div>`).join("")}
      </div></div>`;
    $e("#beRunes").onclick = ev => { const t=ev.target.closest("button"); if(!t) return; const d=t.dataset;
      if(d.prim){ if(st.prim!==d.prim){ st.prim=d.prim; st.ks=""; st.pp=[]; if(st.sec===d.prim){ st.sec=""; st.sp=[]; } } }
      else if(d.sec){ if(st.sec!==d.sec){ st.sec=d.sec; st.sp=[]; } }
      else if(d.pr!=null){ const i=+d.pr; if(i===0) st.ks=d.n; else st.pp[i-1]=d.n; }
      else if(d.sr){ const T=treeOf(st.sec); const slotOf=n=>T.slots.findIndex(s=>s.runes.some(r=>r.name===n));
        if(st.sp.includes(d.n)) st.sp=st.sp.filter(x=>x!==d.n);
        else { st.sp=st.sp.filter(x=>slotOf(x)!==slotOf(d.n)); st.sp.push(d.n); if(st.sp.length>2) st.sp.shift(); } }
      else if(d.sh!=null){ st.shards[+d.sh]=d.n; }
      drawRunes(); };
  }
  function drawSpells(){
    $e("#beSpells").innerHTML = `<div class="bspl">${SPELL_LIST.map(n=>{ const i=spellInfo(n); return `<button type="button" class="bopt sp ${st.spells.includes(n)?"on":""}" data-n="${n}" title="${e(i.name)}">${i.icon?`<img src="${e(i.icon)}" alt="">`:""}<span>${e(i.name)}</span></button>`; }).join("")}</div>`;
    $e("#beSpells").onclick = ev => { const t=ev.target.closest("button"); if(!t) return; const n=t.dataset.n;
      if(st.spells.includes(n)) st.spells=st.spells.filter(x=>x!==n); else { st.spells.push(n); if(st.spells.length>2) st.spells.shift(); } drawSpells(); };
  }
  function drawSkills(){
    const sp=st.cd&&st.cd.spells;
    const row=(k,i)=>{ const s=sp&&sp[i]; const icon=s&&DDX.ver?`https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/img/spell/${s.image.full}`:"";
      return `<tr><th>${ico({name:s?s.name:k, icon},"sk")}<span class="kk">${k}</span></th>${st.skills.map((x,l)=>`<td data-k="${k}" data-l="${l}" class="ed ${x===k?"on k"+k:""}">${x===k?l+1:""}</td>`).join("")}</tr>`; };
    $e("#beSkills").innerHTML = `<div class="fskw"><table class="fsk"><tr><th></th>${st.skills.map((x,i)=>`<td class="lv">${i+1}</td>`).join("")}</tr>${["Q","W","E","R"].map(row).join("")}</table></div>
      <div class="bskb"><button type="button" class="btn sm ghost" data-clr="1">Limpar</button><span>${st.skills.filter(Boolean).length}/18 níveis</span></div>`;
    $e("#beSkills").onclick = ev => { if(ev.target.closest("[data-clr]")){ st.skills=st.skills.map(()=> ""); return drawSkills(); }
      const td=ev.target.closest("td.ed"); if(!td) return; const l=+td.dataset.l; st.skills[l] = st.skills[l]===td.dataset.k ? "" : td.dataset.k; drawSkills(); };
  }
  function drawItems(){
    $e("#beItems").innerHTML = GROUPS.map(([k,t,max])=>`<div class="bgrp ${st.grp===k?"on":""}" data-g="${k}">
      <div class="bgh"><b>${t}</b><span>${st.items[k].length}/${max}</span>${st.grp===k?'<em>↓ escolha abaixo</em>':'<button type="button" class="btn sm" data-g="'+k+'">+ adicionar</button>'}</div>
      <div class="fitems">${st.items[k].map((n,i)=>{ const it=itemInfo(n); return `${i&&k==="core"?'<i class="farr">›</i>':""}<div class="fit bit">${ico(it,"i")}<span>${e(it.name)}</span>
        ${k==="core"&&i?`<button type="button" class="bmv" data-mv="${k}" data-i="${i}" title="Mover pra antes">‹</button>`:""}<button type="button" class="bx" data-rm="${k}" data-i="${i}" title="Remover">×</button></div>`; }).join("")||'<span class="fnone">Nenhum item.</span>'}</div></div>`).join("");
    $e("#beItems").onclick = ev => { const t=ev.target.closest("button,.bgrp"); if(!t) return; const d=t.dataset;
      if(d.rm){ st.items[d.rm].splice(+d.i,1); }
      else if(d.mv){ const a=st.items[d.mv], i=+d.i; [a[i-1],a[i]]=[a[i],a[i-1]]; }
      else if(d.g){ st.grp=d.g; }
      drawItems(); };
  }
  function drawPickGrid(){
    const box=$e("#bePickGrid"); if(!box) return;
    if(!DDX.itemList){ box.innerHTML=`<div class="fnone">Carregando itens… (precisa de internet)</div>`; return; }
    const q=norm(st.q); const list=DDX.itemList.filter(x=>(st.cat==="all"||x.cat===st.cat) && (!q || norm(x.pt).includes(q) || norm(x.en).includes(q)))
      .sort((a,b)=> b.gold-a.gold);
    box.innerHTML = list.map(x=>`<button type="button" class="bpi" data-n="${e(x.en)}" title="${e(x.pt)} · ${x.gold} de ouro"><img src="https://ddragon.leagueoflegends.com/cdn/${DDX.ver}/img/item/${x.id}.png" alt=""><span>${e(x.pt)}</span></button>`).join("") || `<div class="fnone">Nenhum item com esse nome.</div>`;
  }
  function drawPicker(){
    const g=GROUPS.find(x=>x[0]===st.grp);
    $e("#bePick").innerHTML = `<div class="bpick"><div class="bph"><span>Adicionando em <b>${g[1]}</b></span>
        <input id="beQ" placeholder="Buscar item… (ex.: Rabadon, Botas)" value="${e(st.q)}">
        <div class="filters">${CATS.map(([k,t])=>`<button type="button" data-cat="${k}" class="${st.cat===k?"on":""}">${t}</button>`).join("")}</div></div>
      <div class="bpgrid" id="bePickGrid"></div></div>`;
    $e("#beQ").oninput = ev => { st.q=ev.target.value; drawPickGrid(); };
    $e("#bePick").onclick = ev => { const c=ev.target.closest("[data-cat]"); if(c){ st.cat=c.dataset.cat; drawPicker(); return; }
      const t=ev.target.closest(".bpi"); if(!t) return; const arr=st.items[st.grp], max=g[2];
      if(arr.includes(t.dataset.n) && st.grp!=="starting"){ toastE("Esse item já está nesse grupo"); return; }
      if(arr.length>=max){ toastE(`Limite de ${max} itens nesse grupo`); return; }
      arr.push(t.dataset.n); drawItems(); };
    drawPickGrid();
  }
  const toastE = m => (window.toast ? window.toast(m) : console.log(m));
  const redrawItemsAndPicker = () => { drawItems(); drawPicker(); };
  const origDrawItems = drawItems;
  // trocar de grupo também atualiza o título do seletor
  $e("#beItems").addEventListener("click", ev => { if(ev.target.closest("[data-g]") && !ev.target.closest("[data-rm],[data-mv]")) setTimeout(drawPicker); });

  async function loadChamp(){ const n=getChamp&&getChamp(); st.cd = n ? await champData(n) : null; drawSkills(); }
  const all = () => { drawRunes(); drawSpells(); drawSkills(); origDrawItems(); drawPicker(); };
  all();
  loadDD().then(()=>{ if(el.isConnected){ all(); loadChamp(); } });

  return {
    refreshChamp: loadChamp,
    get(){
      const firstGap = st.skills.indexOf(""); const skills = firstGap<0 ? st.skills.slice() : st.skills.slice(0, firstGap);
      const runes = { primary:{tree:st.prim, keystone:st.ks, perks:st.pp.filter(Boolean)}, secondary:{tree:st.sec, perks:st.sp.slice()}, shards:st.shards.filter(Boolean) };
      const itemsAny = GROUPS.some(([k])=>st.items[k].length);
      const v = { spells:st.spells.slice(), skills, runes, runePages: st.ks?[runes]:[],
        items: itemsAny ? [{ name:"Build principal", ...Object.fromEntries(GROUPS.map(([k])=>[k, st.items[k].slice()])) }] : [] };
      if(!st.ks && !itemsAny && !v.spells.length && !skills.length) return null;
      return reSanitize(v);
    },
    problems(){
      const p=[]; if(st.prim && (!st.ks || st.pp.filter(Boolean).length<3)) p.push("faltam runas na árvore principal");
      if(st.sec && st.sp.length<2) p.push("faltam runas na árvore secundária");
      const gap=st.skills.indexOf(""); if(gap>0 && st.skills.slice(gap).some(Boolean)) p.push(`a ordem de habilidades tem um buraco no nível ${gap+1}`);
      return p;
    }
  };
}
window.FLBEdit = { mount: mountEditor, summary: buildSummary };
})();
