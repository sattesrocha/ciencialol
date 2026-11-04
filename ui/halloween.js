// 🎃 Tema de Halloween: troca as cores do app inteiro e adiciona lua, morcegos, névoa, teias,
// cemitério, aranha, abóboras e efeitos. Liga/desliga no botão "🎃 Halloween" do menu.
"use strict";
(function(){
  const MAP = [
    // cores principais (variáveis do :root)
    ["#070d1a","#0a0510"],["#0b1424","#130a1b"],["#0f1b2e","#1b0f25"],["#13223a","#24132f"],
    ["#1f3352","#3a2245"],["#2b4569","#4f2d5c"],["#c8aa6e","#ff8a1f"],["#f0e6d2","#ffe9cf"],
    ["#785a28","#9a3d00"],["#0ac8b9","#9bff5a"],["#0397ab","#5ccf2a"],["#dfe7f2","#f2e8f5"],
    ["#8a9bb4","#ae98ba"],["#5b6d88","#78628a"],["#c9d4e4","#e3d5ea"],["#b9c6da","#d8c8e0"],
    ["#0c1729","#150b1e"],["#7ff3e8","#c8ff9e"],["#1d2f4b","#3d2149"],
    // botões dourados viram abóbora
    ["#d9bd84","#ffb347"],["#a8864a","#e5600e"],["#e6cf9c","#ffc98a"],
    ["#f0d9a6","#ffc070"],["#9c7a3c","#c2410c"],["#f0e0b8","#ffd29a"],
    ["#fbf3df","#ffe7b0"],["#e8cf98","#ffa53a"],["#a8823f","#b8340a"],
    // brilhos e fundos translúcidos
    ["rgba(200,170,110,","rgba(255,138,31,"],["rgba(10,200,185,","rgba(155,255,90,"],
    ["rgba(7,13,26,","rgba(14,6,20,"],["rgba(3,6,12,","rgba(8,3,12,"],["rgba(20,40,70,","rgba(45,16,60,"],
    ["rgba(19,34,58,","rgba(38,18,48,"],["rgba(15,27,46,","rgba(30,14,38,"],["rgba(12,23,41,","rgba(22,10,30,"],
  ];
  const FRASES = [["🎃","Roleta das Bruxas iniciada"],["🧙","Megabrain mexendo o caldeirão"],["🕸️","Algo se mexeu nas teias…"],["👻","UUUH, que ciência assombrada!"],["🦇","Que comecem as travessuras"]];
  const swap = css => { let s=css; for(const [a,b] of MAP) s=s.split(a).join(b).split(a.toUpperCase()).join(b); return s; };

  // ---------- desenhos ----------
  const PUMPKIN = (id) => `<svg viewBox="0 0 200 190" class="hw-pk" aria-hidden="true"><defs>
    <radialGradient id="pk${id}" cx="45%" cy="38%" r="70%"><stop offset="0" stop-color="#ffc061"/><stop offset=".5" stop-color="#ff7a1a"/><stop offset="1" stop-color="#a8320a"/></radialGradient>
    <radialGradient id="pkg${id}" cx="50%" cy="50%" r="60%"><stop offset="0" stop-color="#fff7c2"/><stop offset=".45" stop-color="#ffd23f"/><stop offset="1" stop-color="#ff7b00"/></radialGradient></defs>
    <path d="M96 34C93 20 99 8 113 4c-5 9-5 19-2 30z" fill="#4f6b25" stroke="#2d3f12" stroke-width="2"/>
    <path d="M110 12c10-8 24-6 30 2-10-3-20 0-27 6z" fill="#5f8a2a"/>
    <g stroke="rgba(110,25,0,.45)" stroke-width="2">
      <ellipse cx="58" cy="112" rx="50" ry="62" fill="url(#pk${id})"/><ellipse cx="142" cy="112" rx="50" ry="62" fill="url(#pk${id})"/>
      <ellipse cx="80" cy="110" rx="42" ry="68" fill="url(#pk${id})"/><ellipse cx="120" cy="110" rx="42" ry="68" fill="url(#pk${id})"/>
      <ellipse cx="100" cy="108" rx="33" ry="70" fill="url(#pk${id})"/></g>
    <g class="hw-face" fill="url(#pkg${id})">
      <path d="M56 90 79 68l10 30z"/><path d="M144 90 121 68l-10 30z"/><path d="M100 104l-9 15h18z"/>
      <path d="M48 130q52 48 104 0l-11 3-6 11-10-7-11 15-14-11-14 11-11-15-10 7-6-11z"/></g></svg>`;
  const BAT = `<svg viewBox="0 0 64 30" aria-hidden="true"><g class="hw-wing l"><path d="M31 14C25 5 13 3 1 9c6 2 8 6 6 10 6-4 10 0 14 2 2-4 6-4 10-2z"/></g>
    <g class="hw-wing r"><path d="M33 14c6-9 18-11 30-5-6 2-8 6-6 10-6-4-10 0-14 2-2-4-6-4-10-2z"/></g>
    <ellipse cx="32" cy="16" rx="4.5" ry="6.5"/><path d="M28.5 10.5 29 6l2.3 3.4h1.4L35 6l.5 4.5z"/></svg>`;
  function webSVG(){ // teia de canto (canto de cima à direita)
    const cx=100, cy=0, R=[16,32,50,70,92], A=[90,108,126,144,162,180].map(a=>a*Math.PI/180);
    let d=""; A.forEach(a=>{ d+=`M${cx} ${cy}L${(cx+Math.cos(a)*100).toFixed(1)} ${(cy+Math.sin(a)*100).toFixed(1)}`; });
    R.forEach(r=>{ for(let i=0;i<A.length-1;i++){ const a0=A[i], a1=A[i+1], am=(a0+a1)/2, rs=r*0.86;
      d+=`M${(cx+Math.cos(a0)*r).toFixed(1)} ${(cy+Math.sin(a0)*r).toFixed(1)}Q${(cx+Math.cos(am)*rs).toFixed(1)} ${(cy+Math.sin(am)*rs).toFixed(1)} ${(cx+Math.cos(a1)*r).toFixed(1)} ${(cy+Math.sin(a1)*r).toFixed(1)}`; } });
    return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="${d}" fill="none" stroke="rgba(230,215,240,.55)" stroke-width=".8"/></svg>`;
  }
  const GRAVE = `<svg viewBox="0 0 1600 230" preserveAspectRatio="xMidYMax slice" aria-hidden="true">
    <defs><linearGradient id="hwHill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#160a1f"/><stop offset="1" stop-color="#050208"/></linearGradient></defs>
    <path d="M0 170Q200 120 420 150T860 140T1250 150T1600 130V230H0z" fill="#0e0615" opacity=".85"/>
    <path d="M0 190Q260 150 520 178T1060 170T1600 176V230H0z" fill="url(#hwHill)"/>
    <g fill="#08030d">
      <path d="M120 186V60M120 110l-38-30M120 92l34-36M120 140l-30-12M98 80l-12-20M140 64l14-8M82 80l-18 0" stroke="#08030d" stroke-width="7" stroke-linecap="round" fill="none"/>
      <path d="M112 190q8-30 16 0z"/>
      <path d="M300 182v-36a22 22 0 0 1 44 0v36z"/><path d="M390 186v-26a16 16 0 0 1 32 0v26z"/>
      <path d="M478 178v-40h-14v-12h14v-14h12v14h14v12h-14v40z"/>
      <path d="M1180 180v-34a20 20 0 0 1 40 0v34z"/><path d="M1262 184v-24h-12v-10h12v-12h10v12h12v10h-12v24z"/>
      <path d="M1320 182v-30a18 18 0 0 1 36 0v30z"/>
      <g stroke="#08030d" stroke-width="4"><path d="M1400 182V134M1424 180V130M1448 178V128M1472 176V126M1496 175V125M1520 174V124M1544 173V123M1390 150H1560"/></g>
      <path d="M1396 134l4-10 4 10zM1420 130l4-10 4 10zM1444 128l4-10 4 10zM1468 126l4-10 4 10zM1492 125l4-10 4 10zM1516 124l4-10 4 10zM1540 123l4-10 4 10z"/>
    </g>
    <g class="hw-mini">
      <g transform="translate(560 166) scale(.16)">${PUMPKIN("g1").replace(/<svg[^>]*>|<\/svg>/g,"")}</g>
      <g transform="translate(1100 158) scale(.2)">${PUMPKIN("g2").replace(/<svg[^>]*>|<\/svg>/g,"")}</g>
      <g transform="translate(236 176) scale(.12)">${PUMPKIN("g3").replace(/<svg[^>]*>|<\/svg>/g,"")}</g>
    </g></svg>`;
  const SPIDER = `<div class="hw-spider" title="Boo!"><i class="hw-thread"></i><svg viewBox="0 0 40 40" aria-hidden="true">
    <g stroke="#3a1d4a" stroke-width="2.4" fill="none" stroke-linecap="round"><path d="M14 18 4 10M14 21 2 20M14 24 4 32M15 27 8 38M26 18l10-8M26 21l12-1M26 24l10 8M25 27l7 11"/></g>
    <circle cx="20" cy="16" r="6" fill="#2a1238" stroke="#7a4a96" stroke-width="1"/><circle cx="20" cy="26" r="9" fill="#2a1238" stroke="#7a4a96" stroke-width="1"/><path d="M17 24h6l-3 5z" fill="#ff3b3b" opacity=".85"/>
    <circle cx="18" cy="15" r="1.4" fill="#ffdd55"/><circle cx="22" cy="15" r="1.4" fill="#ffdd55"/></svg></div>`;

  // ---------- liga / desliga ----------
  const HW = { on:false, FRASES, orig:new Map(), els:[] };
  function styles(){ return [...document.querySelectorAll("style")].filter(s=>!s.dataset.hw); }
  function add(el, parent){ (parent||document.body).appendChild(el); HW.els.push(el); return el; }
  function mk(html){ const d=document.createElement("div"); d.innerHTML=html.trim(); return d.firstChild; }

  function decorate(){
    // fundo: lua, estrelas, névoa, cemitério e vaga-lumes
    let stars=""; for(let i=0;i<70;i++) stars+=`<i style="left:${(Math.random()*100).toFixed(1)}%;top:${(Math.random()*55).toFixed(1)}%;animation-delay:${(Math.random()*4).toFixed(2)}s;transform:scale(${(.4+Math.random()*.9).toFixed(2)})"></i>`;
    let flies=""; for(let i=0;i<22;i++){ const g=Math.random()<.45; flies+=`<i class="${g?"g":""}" style="left:${(Math.random()*100).toFixed(1)}%;animation-delay:-${(Math.random()*16).toFixed(1)}s;animation-duration:${(10+Math.random()*10).toFixed(1)}s"></i>`; }
    add(mk(`<div id="hwbg" aria-hidden="true"><div class="hw-stars">${stars}</div><div class="hw-moon"><i></i><i></i><i></i></div>
      <div class="hw-fog a"></div><div class="hw-fog b"></div><div class="hw-grave">${GRAVE}</div><div class="hw-flies">${flies}</div></div>`));
    // morcegos voando por cima
    let bats=""; for(let i=0;i<4;i++) bats+=`<div class="hw-bat b${i}">${BAT}</div>`;
    add(mk(`<div id="hwfly" aria-hidden="true">${bats}</div>`));
    // aranha no menu
    const side=document.querySelector(".side"); if(side) add(mk(SPIDER), side);
    // logo do menu vira abóbora
    const fl=document.querySelector(".logo .flask"); if(fl){ HW.orig.set("flask", fl.innerHTML); fl.innerHTML=PUMPKIN("s"); }
    // início: abóbora gigante, gotas no título e selo
    const hl=document.querySelector(".hero-logo"); if(hl) hl.insertAdjacentElement("afterend", add(mk(`<div class="hw-hero">${PUMPKIN("h")}<div class="hw-bubbles"><i></i><i></i><i></i><i></i><i></i></div></div>`), hl.parentNode));
    const ht=document.querySelector(".hero-title"); if(ht){ let dr=""; [12,27,41,58,71,86].forEach((x,k)=>dr+=`<i style="left:${x}%;animation-delay:${(k*0.7).toFixed(1)}s"></i>`); add(mk(`<div class="hw-drips" aria-hidden="true">${dr}</div>`), ht); }
    const ap=document.querySelector(".alpha-pill"); if(ap) ap.insertAdjacentElement("afterend", add(mk(`<span class="hw-pill">🎃 Edição de Halloween</span>`), ap.parentNode));
    const cdt=document.querySelector(".cd-tag"); if(cdt){ HW.orig.set("cdtag", cdt.textContent); cdt.textContent="ROLETA DAS BRUXAS"; }
    document.documentElement.style.setProperty("--hwweb", `url("data:image/svg+xml,${encodeURIComponent(webSVG())}")`);
  }
  function undecorate(){
    HW.els.forEach(e=>e.remove()); HW.els=[];
    const fl=document.querySelector(".logo .flask"); if(fl && HW.orig.has("flask")) fl.innerHTML=HW.orig.get("flask");
    const cdt=document.querySelector(".cd-tag"); if(cdt && HW.orig.has("cdtag")) cdt.textContent=HW.orig.get("cdtag");
  }
  HW.set = function(on){
    on=!!on; if(on===HW.on) return; HW.on=on;
    styles().forEach(s=>{ if(!HW.orig.has(s)) HW.orig.set(s, s.textContent); s.textContent = on ? swap(HW.orig.get(s)) : HW.orig.get(s); });
    document.documentElement.classList.toggle("hw", on);
    on ? decorate() : undecorate();
    const t=document.getElementById("hwToggle"); if(t) t.classList.toggle("on", on);
  };

  // ---------- efeitos ----------
  function burst(x, y, items, opts={}){
    if(!HW.on) return;
    for(let i=0;i<items.length;i++){
      const e=document.createElement("div"); e.className="hw-pop "+(opts.cls||""); e.innerHTML=items[i];
      const ang = opts.up ? (-Math.PI/2 + (Math.random()-.5)*1.6) : Math.random()*Math.PI*2, dist=(opts.dist||140)*(.6+Math.random()*.6);
      e.style.left=x+"px"; e.style.top=y+"px";
      e.style.setProperty("--dx", (Math.cos(ang)*dist).toFixed(0)+"px"); e.style.setProperty("--dy", (Math.sin(ang)*dist - (opts.lift||0)).toFixed(0)+"px");
      e.style.setProperty("--rot", ((Math.random()-.5)*120).toFixed(0)+"deg"); e.style.animationDelay=(Math.random()*.12).toFixed(2)+"s";
      document.body.appendChild(e); setTimeout(()=>e.remove(), 1900);
    }
  }
  HW.batBurst = el => { const r=el.getBoundingClientRect(); burst(r.left+r.width/2, r.top+r.height/2, Array(7).fill(BAT), {cls:"bat", up:true, dist:260, lift:120}); };
  document.addEventListener("click", e=>{
    if(!HW.on) return;
    const d=e.target.closest("#drawBtn"); if(d && !d.disabled){ HW.batBurst(d); return; }
    const r=e.target.closest("[data-res]"); if(r && !r.classList.contains("on") && !r.closest(".locked")){
      const b=r.getBoundingClientRect();
      if(r.dataset.res==="w") burst(b.left+b.width/2, b.top+b.height/2, ["🍬","🍭","🎃","🍬","🍫","🍭","🍬","✨"], {dist:110});
      else burst(b.left+b.width/2, b.top, ["👻","👻","💀"], {cls:"ghost", up:true, dist:90, lift:60});
    }
    const sp=e.target.closest(".hw-spider"); if(sp){ sp.classList.remove("boo"); void sp.offsetWidth; sp.classList.add("boo"); }
  }, true);

  // botão no menu
  function toggleBtn(){
    const vb=document.querySelector(".volbar"); if(!vb || document.getElementById("hwToggle")) return;
    const b=mk(`<button class="hwtoggle" id="hwToggle" title="Liga ou desliga o tema de Halloween"><span>🎃</span>Halloween<i class="sw"></i></button>`);
    vb.parentNode.insertBefore(b, vb);
    b.classList.toggle("on", HW.on);
    b.onclick=()=>{ HW.set(!HW.on); try{ S.opts.halloween=HW.on; save(); }catch(e){} try{ localStorage.setItem("hwpref", HW.on?"1":"0"); }catch(e){} };
  }
  // ---------- sons extras (se o arquivo existir no app) ----------
  HW.has = { musica:false, roleta:false, lobo:false };
  ["musica","roleta","lobo"].forEach(k=>{ try{ fetch("snd/hw_"+k+".mp3", {method:"HEAD"}).then(r=>{ HW.has[k]=r.ok; if(k==="musica") HW.music.sync(); }).catch(()=>{}); }catch(e){} });

  // ---------- música de fundo (em loop, baixinha; só pausa durante a roleta da prenda, o volume fica com cada um) ----------
  const ducks = new Set(); let voicing = 0;
  const mv = () => { try{ return S.opts.hwMusic ?? 30; }catch(e){ return 30; } };
  const soundOn = () => { try{ return S.opts.sound!==false; }catch(e){ return true; } };
  const master = () => { try{ return Math.pow((S.opts.vol??70)/100, 1.8); }catch(e){ return .5; } };
  HW.music = {
    a:null, fadeT:null,
    vol(){ return HW.minimized ? 0 : Math.min(1, master()*mv()/100) * (voicing ? .3 : 1); },
    want(){ return HW.on && HW.has.musica && soundOn() && mv()>0 && !ducks.size && (typeof NSND==="undefined" || NSND.done || !window.sndReady); },
    ramp(to, ms, done, a=this.a){ if(!a) return; clearInterval(a._ft); const from=+a.volume||0, t0=Date.now();
      a._ft=setInterval(()=>{ const k=Math.min(1,(Date.now()-t0)/ms); try{ a.volume=from+(to-from)*k; }catch(e){} if(k>=1){ clearInterval(a._ft); done&&done(); } }, 60); },
    sync(){
      if(this.want()){
        if(!this.a){
          if(typeof NSND!=="undefined" && NSND.ok && typeof nativeSound==="function") this.a=nativeSound("hw_musica", 0, true);
          else { const a=new Audio("snd/hw_musica.mp3"); a.loop=true; a.volume=0; a.play().catch(()=>{}); this.a=a; }
          this.ramp(this.vol(), 1500);
        } else this.ramp(this.vol(), 250);
      } else if(this.a){ const a=this.a; this.a=null; this.ramp(0, 500, ()=>{ try{ a.pause(); }catch(e){} }, a); }
      renderMus();
    },
  };
  HW.duck = (on, why="prenda") => { on ? ducks.add(why) : ducks.delete(why); HW.music.sync(); };
  function playSnd(name, vol, loop){
    if(typeof NSND!=="undefined" && NSND.ok && typeof nativeSound==="function") return nativeSound(name, vol, loop);
    const a=new Audio("snd/"+name+".mp3"); a.volume=vol; a.loop=!!loop; a.play().catch(()=>{}); return a;
  }
  // música da Roleta das Bruxas (contagem do sorteio)
  HW.cdMusic = () => {
    if(!HW.on || !HW.has.roleta || !soundOn()) return null;
    HW.duck(true, "cd"); const a=playSnd("hw_roleta", Math.min(1, master()), false);
    return { stop(){ HW.music.ramp(0, 900, ()=>{ try{ a.pause(); }catch(e){} HW.duck(false, "cd"); }, a); } };
  };

  // ---------- vozes (no volume dos efeitos; a música abaixa enquanto falam) ----------
  const VOZ_MS = { hw_welcome:8900, hw_v1:3200, hw_v2:2700, hw_v3:5600, hw_v4:3800, hw_v5:5700 };
  HW.voice = (name, scare) => {
    if(!HW.on || !soundOn() || !master()) return false;
    voicing++; HW.music.sync(); playSnd(name, Math.min(1, master()), false);
    if(scare) scareFx(VOZ_MS[name]||4000);
    setTimeout(()=>{ voicing=Math.max(0,voicing-1); HW.music.sync(); }, (VOZ_MS[name]||4000)+400);
    return true;
  };
  function scareFx(ms){
    const e=mk(`<div class="hw-scare" aria-hidden="true"><i class="hw-ghost"><svg viewBox="0 0 120 160" width="120"><defs><radialGradient id="hwgh" cx="50%" cy="35%" r="70%"><stop offset="0" stop-color="#ffffff" stop-opacity=".95"/><stop offset=".7" stop-color="#e6e2ff" stop-opacity=".7"/><stop offset="1" stop-color="#cfc8ff" stop-opacity="0"/></radialGradient></defs><path d="M60 6C28 6 12 32 12 62v80c8-10 14-10 20 0 6-10 14-10 20 0 6-10 14-10 20 0 6-10 14-10 20 0 6-10 12-10 16 0V62C108 32 92 6 60 6z" fill="url(#hwgh)"/><ellipse cx="44" cy="60" rx="9" ry="14" fill="#0a0410"/><ellipse cx="76" cy="60" rx="9" ry="14" fill="#0a0410"/><ellipse cx="60" cy="98" rx="11" ry="17" fill="#0a0410"/></svg></i></div>`); e.style.setProperty("--dur", ms+"ms");
    document.body.appendChild(e); setTimeout(()=>e.remove(), ms+300);
  }
  const ocupado = () => !!document.querySelector(".pfx, .cd.on, #overlay.on, .uic") || ducks.size>0 || (typeof busy!=="undefined" && busy);
  // sustinhos: uma voz aleatória de vez em quando (a cada 8 a 20 min; a primeira depois de 4 a 10 min)
  const SUSTOS = ["hw_v1","hw_v2","hw_v3","hw_v4","hw_v5"]; let lastSusto="";
  function agendaSusto(min, max){ setTimeout(()=>{
      if(!HW.on || !soundOn()) return agendaSusto(4, 10);
      if(ocupado()) return agendaSusto(1, 3);
      let v; do{ v=SUSTOS[Math.floor(Math.random()*SUSTOS.length)]; }while(v===lastSusto && SUSTOS.length>1); lastSusto=v;
      HW.voice(v, true); agendaSusto(8, 20);
    }, (min + Math.random()*(max-min))*60000); }
  agendaSusto(4, 10);
  // janela minimizada: a música fica muda (continua de onde estava ao voltar);
  // "welcome back" ao abrir de novo depois de minimizado (pelo menos 15 s; no máximo 1 vez a cada 30 min)
  let minEm = 0, welcomeEm = 0, nativo = false;
  function winState(st){
    if(st==="min"){ if(!minEm) minEm=Date.now(); HW.minimized=true; HW.music.sync(); const m=HW.music.a; if(m){ clearInterval(m._ft); try{ m.volume=0; }catch(e){} } return; }
    const fora = minEm ? Date.now()-minEm : 0; minEm=0; HW.minimized=false; HW.music.sync();
    if(fora < 15000 || Date.now()-welcomeEm < 30*60000 || ocupado()) return;
    setTimeout(()=>{ if(!HW.minimized && HW.voice("hw_welcome")) welcomeEm = Date.now(); }, 900);
  }

  window.__winState = st => { nativo = true; winState(st); };
  // fora do programa (navegador): usa a visibilidade da página
  document.addEventListener("visibilitychange", ()=>{ if(!nativo) winState(document.hidden ? "min" : "normal"); });

  // morcegos atravessando a roleta da prenda
  HW.pfxBats = () => { let h=""; for(let i=0;i<7;i++) h+=`<div class="hw-pbat" style="top:${(8+Math.random()*70).toFixed(0)}%;animation-delay:${(4.2+i*.35+Math.random()*.3).toFixed(2)}s;width:${(26+Math.random()*30).toFixed(0)}px">${BAT}</div>`; return `<div class="hw-pfxbats">${h}</div>`; };

  // controle da música no menu
  function renderMus(){
    const box=document.getElementById("hwMusic"); if(!box) return;
    box.style.display = HW.on && HW.has.musica ? "" : "none";
    const v=mv(), r=document.getElementById("hwMusVol"), t=document.getElementById("hwMusPct"), b=document.getElementById("hwMusMute");
    if(r && document.activeElement!==r){ r.value=v; } if(r) r.style.setProperty("--p", v+"%");
    if(t) t.textContent = v ? v+"%" : "mudo"; if(b) b.classList.toggle("off", !v);
  }
  function musicBtn(){
    const tg=document.getElementById("hwToggle"); if(!tg || document.getElementById("hwMusic")) return;
    const box=mk(`<div class="hwmusic" id="hwMusic" title="Volume da música de Halloween"><button id="hwMusMute" aria-label="Ligar ou desligar a música">🎵</button><input type="range" id="hwMusVol" min="0" max="100" aria-label="Volume da música"><span id="hwMusPct"></span></div>`);
    tg.insertAdjacentElement("afterend", box);
    let last=30;
    document.getElementById("hwMusVol").oninput=e=>{ try{ S.opts.hwMusic=+e.target.value; save(); }catch(_){} HW.music.sync(); };
    document.getElementById("hwMusMute").onclick=()=>{ try{ const v=mv(); if(v){ last=v; S.opts.hwMusic=0; } else S.opts.hwMusic=last||30; save(); }catch(_){} HW.music.sync(); };
    renderMus();
  }
  const baseSet = HW.set;
  HW.set = function(on){ baseSet(on); renderMus(); HW.music.sync(); };

  HW.init = function(){ toggleBtn(); musicBtn(); let want=true; try{ want = S.opts.halloween!==false; }catch(e){} HW.set(want); renderMus(); HW.music.sync(); };
  window.HW = HW;
  // aplica o tema já na abertura, antes de mostrar a janela (nada de piscar o visual normal)
  try{ HW.set(window.__HWON!==false); }catch(e){ console.error(e); }
  document.documentElement.classList.remove("hwpre");
})();
