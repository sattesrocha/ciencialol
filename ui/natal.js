// 🎄 Tema de Natal: troca as cores do app inteiro e adiciona neve caindo, aurora, lua com o trenó do
// Papai Noel, vila nevada, pisca-pisca, árvore iluminada, neve acumulada nos cards e efeitos.
// Liga/desliga no botão "🎄 Natal" do menu (desliga o Halloween, um tema de cada vez).
"use strict";
(function(){
  const MAP = [
    // cores principais (variáveis do :root): noite de inverno verde-azulada
    ["#070d1a","#04100f"],["#0b1424","#081a18"],["#0f1b2e","#0c2422"],["#13223a","#112d2a"],
    ["#1f3352","#1b4440"],["#2b4569","#285a53"],["#c8aa6e","#ecc25e"],["#f0e6d2","#fff5e6"],
    ["#785a28","#8f5a12"],["#0ac8b9","#ff4f5e"],["#0397ab","#c8202f"],["#dfe7f2","#eef6f2"],
    ["#8a9bb4","#9db7ae"],["#5b6d88","#66867c"],["#c9d4e4","#d6e8e1"],["#b9c6da","#c4dad2"],
    ["#0c1729","#0a1f1d"],["#7ff3e8","#ff9ba5"],["#1d2f4b","#1a3d38"],
    // botões dourados viram presente vermelho (com fita dourada)
    ["#d9bd84","#ff6670"],["#a8864a","#b0101f"],["#e6cf9c","#ff9aa1"],
    ["#f0d9a6","#ff7c84"],["#9c7a3c","#a10d1b"],["#f0e0b8","#ffc0c4"],
    ["#fbf3df","#fff0f1"],["#e8cf98","#ff6c76"],["#a8823f","#a30f1d"],
    // brilhos e fundos translúcidos
    ["rgba(200,170,110,","rgba(236,194,94,"],["rgba(10,200,185,","rgba(255,79,94,"],
    ["rgba(7,13,26,","rgba(4,16,15,"],["rgba(3,6,12,","rgba(2,9,9,"],["rgba(20,40,70,","rgba(16,52,48,"],
    ["rgba(19,34,58,","rgba(17,45,42,"],["rgba(15,27,46,","rgba(12,36,34,"],["rgba(12,23,41,","rgba(10,31,29,"],
  ];
  const FRASES = [["🎅","Ho ho ho! Roleta do Noel iniciada"],["🦌","Renas aquecendo os motores"],["🎁","Embrulhando a sua ciência"],["⭐","Megabrain conferindo quem se comportou"],["❄️","Nevando ciência no lobby"]];
  // esboço 2: claro, branco de neve (fundo gelo, cards brancos, texto escuro, vermelho e verde de Natal)
  const MAP2 = [
    ["#070d1a","#e6f1f7"],["#0b1424","#f3f8fb"],["#0f1b2e","#ffffff"],["#13223a","#eef5f9"],
    ["#1f3352","#d3e2eb"],["#2b4569","#b9cdda"],["#c8aa6e","#c42634"],["#f0e6d2","#1b2a38"],
    ["#785a28","#e4b9be"],["#0ac8b9","#1d9a55"],["#0397ab","#137a40"],["#dfe7f2","#1e2d3b"],
    ["#8a9bb4","#5a6e80"],["#5b6d88","#687c8d"],["#c9d4e4","#2c3e50"],["#b9c6da","#3a4d5f"],
    ["#0c1729","#f6fafc"],["#7ff3e8","#2fbf6e"],["#1d2f4b","#dce8ef"],["#0a0f1c","#e9f2f7"],["#0a1322","#f3f8fb"],
    ["#9fb0c8","#4d6274"],["#e8a33c","#b8680a"],["#4fbf6a","#23863b"],["#4aa8ff","#1f6fc4"],["#ef5b6b","#c8323f"],["#c58bff","#8a3fd1"],["#2fd38a","#1d9a55"],["#e84057","#c42634"],["#ffcf5c","#b8680a"],["#ffb3bb","#c8323f"],["#c9a8ff","#7a3fc0"],["#e6d4ff","#5a2fa0"],["#ffd38a","#a85a08"],["#ffb056","#b8680a"],["#aebbd0","#4d6274"],["#1a1206","#ffffff"],["#3a2d12","#f6e3e5"],
    ["#d9bd84","#ff6670"],["#a8864a","#b0101f"],["#e6cf9c","#ff9aa1"],
    ["#f0d9a6","#ff7c84"],["#9c7a3c","#a10d1b"],["#f0e0b8","#ffc0c4"],
    ["#fbf3df","#fff0f1"],["#e8cf98","#ff6c76"],["#a8823f","#a30f1d"],
    ["rgba(200,170,110,","rgba(196,38,52,"],["rgba(10,200,185,","rgba(29,154,85,"],
    ["rgba(7,13,26,","rgba(214,230,240,"],["rgba(3,6,12,","rgba(205,224,236,"],["rgba(20,40,70,","rgba(218,234,244,"],
    ["rgba(19,34,58,","rgba(255,255,255,"],["rgba(15,27,46,","rgba(246,250,252,"],["rgba(12,23,41,","rgba(238,245,249,"],
    ["rgba(255,255,255,.08)","rgba(20,50,80,.06)"],["rgba(255,255,255,.05)","rgba(20,50,80,.05)"],["rgba(255,255,255,.06)","rgba(20,50,80,.05)"],["rgba(255,255,255,.04)","rgba(20,50,80,.04)"],["rgba(255,255,255,.03)","rgba(20,50,80,.03)"],
  ];
  const swap = css => { let s=css; for(const [a,b] of (XM.v2?MAP2:MAP)) s=s.split(a).join(b).split(a.toUpperCase()).join(b); return s; };
  const rnd = (a,b) => a + Math.random()*(b-a);
  const LUZES = ["#ff4f5e","#ffd23f","#43e07a","#5ab8ff","#ff9a3c","#ff6bd6"];

  // ---------- desenhos ----------
  const STAR = (cls="") => `<svg viewBox="0 0 100 100" class="${cls}" aria-hidden="true"><path d="M50 4l13 30 33 3-25 22 8 33-29-17-29 17 8-33L4 37l33-3z"/></svg>`;
  function TREE(id){
    // árvore com 4 camadas, neve nas pontas, festão dourado, bolas que piscam, estrela e presentes
    const g = `<defs>
      <linearGradient id="tr${id}" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#2fae5d"/><stop offset=".55" stop-color="#16733b"/><stop offset="1" stop-color="#0b4425"/></linearGradient>
      <radialGradient id="st${id}" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#fffbe0"/><stop offset=".55" stop-color="#ffd23f"/><stop offset="1" stop-color="#e08a00"/></radialGradient>
      <linearGradient id="gf${id}" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#b8860b"/><stop offset=".5" stop-color="#ffe08a"/><stop offset="1" stop-color="#b8860b"/></linearGradient></defs>`;
    const camadas = [[100,34,46,78],[100,62,62,112],[100,94,78,148],[100,128,92,184]];
    let t=""; camadas.forEach(([cx,top,w,bot])=>{
      t+=`<path d="M${cx} ${top}L${cx+w} ${bot}Q${cx} ${bot+12} ${cx-w} ${bot}Z" fill="url(#tr${id})" stroke="#0a3a1f" stroke-width="1.5"/>`;
      // neve escorrendo na borda da camada
      t+=`<path d="M${cx-w+6} ${bot-3}Q${cx-w/2} ${bot+9} ${cx} ${bot+8}Q${cx+w/2} ${bot+9} ${cx+w-6} ${bot-3}Q${cx+w/2} ${bot+3} ${cx} ${bot+2}Q${cx-w/2} ${bot+3} ${cx-w+6} ${bot-3}Z" fill="#f4fbff" opacity=".92"/>`;
      t+=`<path d="M${cx-9} ${top+12}L${cx} ${top}L${cx+9} ${top+12}Q${cx} ${top+8} ${cx-9} ${top+12}Z" fill="#f4fbff" opacity=".95"/>`;
    });
    const fest = `<g fill="none" stroke="url(#gf${id})" stroke-width="3" stroke-linecap="round" opacity=".95">
      <path d="M66 86Q102 104 136 82"/><path d="M52 122Q100 142 150 116"/><path d="M38 158Q100 182 164 152"/></g>`;
    const bolas=[[86,70],[118,74],[72,104],[104,112],[132,100],[60,140],[92,150],[124,146],[150,136],[48,172],[80,182],[114,184],[146,176]];
    let b=""; bolas.forEach(([x,y],i)=>{ const c=LUZES[i%LUZES.length];
      b+=`<g class="xm-orn" style="animation-delay:${(i*.37%2.6).toFixed(2)}s"><circle cx="${x}" cy="${y}" r="6.2" fill="${c}"/><circle cx="${x-2}" cy="${y-2}" r="2" fill="#fff" opacity=".75"/></g>`; });
    const presentes = `<g class="xm-gifts">
      <rect x="40" y="196" width="40" height="30" rx="3" fill="#d6283a"/><rect x="56" y="196" width="8" height="30" fill="#ffd23f"/><rect x="40" y="206" width="40" height="7" fill="#ffd23f"/>
      <path d="M60 196c-10-12-20-6-12 0zM60 196c10-12 20-6 12 0z" fill="#ffd23f"/>
      <rect x="126" y="202" width="34" height="24" rx="3" fill="#2b7de9"/><rect x="139" y="202" width="7" height="24" fill="#f4fbff"/><rect x="126" y="210" width="34" height="6" fill="#f4fbff"/>
      <path d="M142 202c-8-10-17-5-10 0zM142 202c8-10 17-5 10 0z" fill="#f4fbff"/>
      <rect x="84" y="208" width="30" height="18" rx="3" fill="#25a55a"/><rect x="96" y="208" width="6" height="18" fill="#ff4f5e"/></g>`;
    return `<svg viewBox="0 0 200 232" class="xm-tree" aria-hidden="true">${g}
      <rect x="90" y="184" width="20" height="22" rx="2" fill="#5b3418"/>${t}${fest}${b}
      <g class="xm-topstar"><circle cx="100" cy="30" r="20" fill="#ffd23f" opacity=".18"/><path d="M100 10l5.6 12.6 13.8 1.3-10.4 9.2 3.1 13.5L100 39.6l-12.1 7 3.1-13.5-10.4-9.2 13.8-1.3z" fill="url(#st${id})" stroke="#b86f00" stroke-width="1"/></g>
      ${presentes}</svg>`;
  }
  const BAUBLE = (id, cor="#e0283a") => `<svg viewBox="0 0 60 72" class="xm-bauble" aria-hidden="true"><defs>
    <radialGradient id="bb${id}" cx="35%" cy="35%" r="70%"><stop offset="0" stop-color="#fff" stop-opacity=".9"/><stop offset=".18" stop-color="${cor}"/><stop offset="1" stop-color="#4a0710"/></radialGradient>
    <linearGradient id="bc${id}" x1="0" x2="1"><stop offset="0" stop-color="#9a6a12"/><stop offset=".5" stop-color="#ffe08a"/><stop offset="1" stop-color="#9a6a12"/></linearGradient></defs>
    <path d="M26 6a4 4 0 1 1 8 0" fill="none" stroke="#c9a14a" stroke-width="2"/>
    <rect x="22" y="8" width="16" height="9" rx="2" fill="url(#bc${id})"/>
    <circle cx="30" cy="42" r="26" fill="url(#bb${id})"/>
    <path d="M7 40q23 10 46 0" fill="none" stroke="#ffe08a" stroke-width="2.4" opacity=".85"/>
    <path d="M9 50q21 8 42 0" fill="none" stroke="#ffe08a" stroke-width="1.2" opacity=".6"/>
    <g fill="#fff" opacity=".85"><circle cx="18" cy="44" r="1.6"/><circle cx="30" cy="47" r="1.6"/><circle cx="42" cy="44" r="1.6"/></g></svg>`;
  // trenó do Papai Noel com 3 renas (silhueta na frente da lua); a rena da frente tem o nariz vermelho
  const RENA = x => `<g transform="translate(${x} 0)"><ellipse cx="16" cy="24" rx="13" ry="6"/><path d="M5 26l-4 12M9 27l2 12M23 27l-3 12M27 25l5 11" stroke-width="2.4" stroke-linecap="round"/>
    <path d="M27 21q6-8 10-10l4 3q-3 3-8 6z"/><path d="M37 12l-2-8M37 12l3-8M35 7l-4-3M40 7l4-3" fill="none" stroke-width="1.6" stroke-linecap="round"/><path d="M3 22l-4-3" stroke-width="2"/></g>`;
  const SLEIGH = `<svg viewBox="0 0 250 50" aria-hidden="true"><g fill="#071412" stroke="#071412">
    ${RENA(118)}${RENA(160)}${RENA(202)}
    <path d="M20 28L120 20M60 26l60-6" fill="none" stroke-width="1.2"/>
    <path d="M2 40q4 6 12 6h70q8 0 10-6" fill="none" stroke-width="3" stroke-linecap="round"/>
    <path d="M8 40V22q0-10 12-10h28q4 0 6 6l12 6h14q8 0 10 8v8z"/>
    <circle cx="34" cy="8" r="8"/><path d="M28 4q6-14 16-4l6 2" stroke-width="3" fill="none"/><circle cx="52" cy="2" r="3"/>
    <path d="M42 14q14-2 18 8l-18 2z"/></g><circle cx="244" cy="12" r="2.6" fill="#ff3344" class="xm-rudolph"/></svg>`;
  // vila nevada com pinheiros, casinhas de janela acesa, igrejinha e boneco de neve
  const PINE = (x, y, s) => `<g transform="translate(${x} ${y}) scale(${s})"><path d="M0-90L26-40H12L34 0H-34l22-40H-26z" fill="#082a1c"/><path d="M0-90l9 18Q0-76-9-72zM-14-40q14 6 28 0l-4-6q-10 4-20 0zM-20 0q20 6 40 0l-5-7q-15 5-30 0z" fill="#e8f3f6" opacity=".85"/><rect x="-4" y="0" width="8" height="10" fill="#1c140c"/></g>`;
  const HOUSE = (x, y, s, k) => `<g transform="translate(${x} ${y}) scale(${s})"><path d="M-30 0V-34L0-58 30-34V0z" fill="#0b1f1d"/><path d="M-36-32L0-62 36-32l-6 3L0-54-30-29z" fill="#eef7fa"/>
    <rect x="-20" y="-28" width="12" height="12" class="xm-win" style="animation-delay:${k}s"/><rect x="8" y="-28" width="12" height="12" class="xm-win" style="animation-delay:${k+1.3}s"/><rect x="-6" y="-16" width="12" height="16" fill="#3a2414"/>
    <rect x="14" y="-62" width="9" height="18" fill="#0b1f1d"/><path d="M12-62h13v-3H12z" fill="#eef7fa"/></g>`;
  const VILA = `<svg viewBox="0 0 1600 250" preserveAspectRatio="xMidYMax slice" aria-hidden="true">
    <defs><linearGradient id="xmHill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#d9e9ee"/><stop offset="1" stop-color="#7f9fae"/></linearGradient>
    <linearGradient id="xmHill2" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#9fbccb"/><stop offset=".25" stop-color="#4d6878"/><stop offset="1" stop-color="#16262e"/></linearGradient></defs>
    <path d="M0 175Q220 128 470 158T930 146T1300 154T1600 136V250H0z" fill="url(#xmHill)" opacity=".32"/>
    ${PINE(90,170,.9)}${PINE(150,176,.6)}${PINE(1480,160,1)}${PINE(1540,168,.7)}${PINE(760,160,.55)}${PINE(1010,158,.5)}
    ${HOUSE(300,182,.9,0)}${HOUSE(390,186,.7,.8)}${HOUSE(1180,180,.95,1.6)}${HOUSE(1290,186,.7,2.4)}
    <g transform="translate(1390 186) scale(.9)"><path d="M-22 0v-40L0-58l22 18V0z" fill="#0b1f1d"/><path d="M-6-58v-34L0-104l6 12v34z" fill="#0b1f1d"/><path d="M-2-112h4v-8h-4zM-5-116h10v-3H-5z" fill="#ffd23f"/><path d="M-26-38L0-62l26 24-5 2L0-56-21-36z" fill="#eef7fa"/><rect x="-6" y="-30" width="12" height="18" rx="6" class="xm-win" style="animation-delay:.5s"/><circle cx="0" cy="-74" r="4" class="xm-win" style="animation-delay:2s"/></g>
    <path d="M0 200Q260 168 540 190T1080 184T1600 190V250H0z" fill="url(#xmHill2)" opacity=".85"/><path d="M0 200Q260 168 540 190T1080 184T1600 190" fill="none" stroke="#eef7fa" stroke-width="3" opacity=".7"/>
    <g transform="translate(560 204)"><circle cx="0" cy="-12" r="14" fill="#f4fbff"/><circle cx="0" cy="-34" r="10" fill="#f4fbff"/><path d="M-9-42h18v-3h-4v-10H-5v10h-4z" fill="#14201f"/><path d="M0-34l10 2-10 2z" fill="#ff8a1f"/><circle cx="-3" cy="-37" r="1.3" fill="#14201f"/><circle cx="3" cy="-37" r="1.3" fill="#14201f"/><path d="M-10-26q10 4 20 0l-2 4q-8 3-16 0z" fill="#d6283a"/><path d="M-14-16l-12-8M14-16l12-10" stroke="#3a2414" stroke-width="2"/></g>
    ${PINE(640,214,.45)}${PINE(1080,212,.5)}</svg>`;

  // esboço 2: paisagem branca de dia (montanhas, floresta nevada, casinhas)
  const SPINE = (x, y, s, c="#1f6b45") => `<g transform="translate(${x} ${y}) scale(${s})"><path d="M0-96L28-42H14L36 0H-36l22-42H-28z" fill="${c}"/><path d="M0-96l12 22Q0-68-12-74zM-28-42q14 10 28 4 14 6 28-4l-6-8q-10 6-22 2-12 4-22-2zM-36 0q18 8 36 2 18 6 36-2l-7-9q-14 7-29 2-15 5-29-2z" fill="#fff"/></g>`;
  const VILA2 = `<svg viewBox="0 0 1600 300" preserveAspectRatio="xMidYMax slice" aria-hidden="true">
    <defs><linearGradient id="xm2m" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#c9dcea"/></linearGradient>
    <linearGradient id="xm2h" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#e3eef5"/></linearGradient></defs>
    <path d="M0 150L120 70 210 120 330 40 460 130 560 80 700 140 820 60 940 130 1060 70 1200 140 1320 50 1440 120 1600 70V300H0z" fill="url(#xm2m)" opacity=".9"/>
    <path d="M330 40l-40 34 22-6 14 12 10-16 18 10zM820 60l-34 30 20-4 12 10 10-14 16 8zM1320 50l-38 32 22-5 14 11 10-15 17 9z" fill="#ffffff"/>
    <g opacity=".55">${SPINE(60,190,.55,"#5f8f7a")}${SPINE(140,196,.45,"#5f8f7a")}${SPINE(980,186,.5,"#5f8f7a")}${SPINE(1040,192,.4,"#5f8f7a")}${SPINE(1500,184,.55,"#5f8f7a")}</g>
    <path d="M0 210Q260 170 520 196T1040 186T1600 180V300H0z" fill="url(#xm2h)"/>
    ${SPINE(100,226,.9)}${SPINE(170,232,.65)}${SPINE(1450,222,1)}${SPINE(1530,230,.7)}${SPINE(1110,226,.6)}
    <g transform="translate(330 236) scale(.95)"><path d="M-32 0V-34L0-58 32-34V0z" fill="#b8323c"/><path d="M-38-32L0-64 38-32l-6 4L0-54-32-28z" fill="#fff"/><rect x="-20" y="-28" width="12" height="12" class="xm-win"/><rect x="8" y="-28" width="12" height="12" class="xm-win" style="animation-delay:1.2s"/><rect x="-6" y="-16" width="12" height="16" fill="#5a2a18"/></g>
    <g transform="translate(1240 232) scale(.8)"><path d="M-32 0V-34L0-58 32-34V0z" fill="#2c6e8a"/><path d="M-38-32L0-64 38-32l-6 4L0-54-32-28z" fill="#fff"/><rect x="-20" y="-28" width="12" height="12" class="xm-win" style="animation-delay:.6s"/><rect x="8" y="-28" width="12" height="12" class="xm-win" style="animation-delay:2s"/><rect x="-6" y="-16" width="12" height="16" fill="#3a2414"/></g>
    <path d="M0 252Q300 226 640 244T1260 238T1600 242V300H0z" fill="#ffffff"/>
    <g transform="translate(468 262)"><ellipse cx="0" cy="4" rx="26" ry="5" fill="#cfe0ea"/><circle cx="0" cy="-14" r="17" fill="#fff" stroke="#d3e2eb"/><circle cx="0" cy="-40" r="12" fill="#fff" stroke="#d3e2eb"/><path d="M-11-50h22v-4h-5v-12H-6v12h-5z" fill="#1d2b38"/><path d="M0-40l12 2-12 3z" fill="#ff8a1f"/><circle cx="-4" cy="-43" r="1.5" fill="#1d2b38"/><circle cx="4" cy="-43" r="1.5" fill="#1d2b38"/><path d="M-12-30q12 5 24 0l-2 5q-10 4-20 0z" fill="#c42634"/><path d="M12-29l4 14-6-1z" fill="#c42634"/><path d="M-16-18l-14-10M16-18l14-12" stroke="#5a3a20" stroke-width="2.4"/></g>
    ${SPINE(1120,272,.55)}</svg>`;
  const flakeSprites = [];
  function mkSprites(){ // floco com sombrinha azulada (aparece no fundo branco)
    flakeSprites.length=0;
    [2,3,4.5,6].forEach(r=>{ const cv=document.createElement("canvas"), sz=Math.ceil(r*2+10); cv.width=cv.height=sz; const c=cv.getContext("2d");
      c.shadowColor="rgba(70,110,150,.55)"; c.shadowBlur=4; c.fillStyle="#fff"; c.beginPath(); c.arc(sz/2,sz/2,r,0,6.283); c.fill();
      flakeSprites.push({cv, sz}); });
  }

  // ---------- liga / desliga ----------
  const XM = { on:false, FRASES, orig:new Map(), els:[] };
  // qual esboço (1: noite / 2: branco nevado). Enquanto escolhe: localStorage "xmvar"
  // estilo do Natal: "branco" (neve, claro; padrão) ou "noite"
  { let v; try{ v=localStorage.getItem("xmvar"); }catch(e){} if(v!=="1" && v!=="2") v = window.__XMVAR==="noite" ? "1" : "2"; XM.v2 = v==="2"; }
  XM.esboco = n => { const was=XM.on; if(XM.music) XM.music.keep=true; // trocar o estilo não mexe na música
    if(was) XM.set(false); XM.orig=new Map(); XM.v2 = n===2;
    try{ localStorage.setItem("xmvar", String(n)); }catch(e){} try{ S.opts.natalEstilo = n===2 ? "branco" : "noite"; save(); }catch(e){}
    if(was) XM.set(true); if(XM.music) XM.music.keep=false; renderEstilo(); };
  function styles(){ return [...document.querySelectorAll("style")].filter(s=>!s.dataset.hw); }
  function add(el, parent){ (parent||document.body).appendChild(el); XM.els.push(el); return el; }
  function mk(html){ const d=document.createElement("div"); d.innerHTML=html.trim(); return d.firstChild; }

  // neve (canvas leve: 3 camadas, vento suave; pausa com a janela minimizada)
  const NEVE = { cv:null, ctx:null, fl:[], raf:0, w:0, h:0, last:0, boost:0 };
  function neveStart(){
    const cv=add(mk(`<canvas id="xmsnow" aria-hidden="true"></canvas>`)); NEVE.cv=cv; NEVE.ctx=cv.getContext("2d");
    const fit=()=>{ const d=Math.min(2, window.devicePixelRatio||1); NEVE.w=innerWidth; NEVE.h=innerHeight; cv.width=NEVE.w*d; cv.height=NEVE.h*d; NEVE.ctx.setTransform(d,0,0,d,0,0); };
    fit(); NEVE.fit=fit; addEventListener("resize", fit);
    if(XM.v2) mkSprites();
    NEVE.fl=[]; for(let i=0;i<(XM.v2?130:140);i++) NEVE.fl.push(flake(true));
    NEVE.last=performance.now(); NEVE.raf=requestAnimationFrame(neveTick);
  }
  function flake(any){ const z=Math.random(); return { x:Math.random()*NEVE.w, y:any?Math.random()*NEVE.h:-10, z, r:.8+z*2.6, v:18+z*48, ph:Math.random()*6.28, sw:8+Math.random()*22, a:.35+z*.55 }; }
  function neveTick(t){
    if(!XM.on) return;
    NEVE.raf=requestAnimationFrame(neveTick);
    const dt=Math.min(.05,(t-NEVE.last)/1000); NEVE.last=t;
    if(XM.minimized || document.hidden) return;
    const c=NEVE.ctx; c.clearRect(0,0,NEVE.w,NEVE.h);
    const vento=Math.sin(t/7000)*14, mult=1+NEVE.boost;
    for(const f of NEVE.fl){
      f.y+=f.v*dt*mult; f.ph+=dt*(.6+f.z); f.x+= (vento*(.4+f.z) + Math.cos(f.ph)*f.sw*.6)*dt;
      if(f.y>NEVE.h+6||f.x<-20||f.x>NEVE.w+20){ Object.assign(f, flake(false)); }
      c.globalAlpha=f.a;
      if(XM.v2){ const sp=flakeSprites[Math.min(3, Math.floor(f.z*4))]; c.drawImage(sp.cv, f.x-sp.sz/2, f.y-sp.sz/2); }
      else { c.beginPath(); c.arc(f.x,f.y,f.r,0,6.283); c.fillStyle="#fff"; c.fill(); }
    }
    c.globalAlpha=1; if(NEVE.boost>0) NEVE.boost=Math.max(0, NEVE.boost-dt*.6);
  }
  function neveStop(){ cancelAnimationFrame(NEVE.raf); if(NEVE.fit) removeEventListener("resize", NEVE.fit); NEVE.cv=null; }
  XM.nevasca = (s=2.5) => { NEVE.boost=s; };

  // pisca-pisca no alto da tela (fio com curvas e lâmpadas coloridas em sequência)
  function piscaPisca(){
    let segs=""; for(let i=0;i<36;i++){
      let bulbs=""; [[22,13],[60,20],[98,13]].forEach(([x,y],k)=>{ const n=i*3+k, c=LUZES[n%LUZES.length];
        bulbs+=`<g class="xm-bulb" style="--c:${c};animation-delay:${((n%6)*.35).toFixed(2)}s"><rect x="${x-2.5}" y="${y-1}" width="5" height="5" rx="1" fill="#2c3a36"/><path d="M${x} ${y+3}c-5 4-5 13 0 16 5-3 5-12 0-16z" fill="${c}"/></g>`; });
      segs+=`<svg viewBox="0 0 120 40" aria-hidden="true"><path d="M0 8Q30 26 60 18T120 8" fill="none" stroke="#18302b" stroke-width="2"/>${bulbs}</svg>`;
    }
    const side=document.querySelector(".side"), left = side ? side.getBoundingClientRect().width : 0;
    const el=add(mk(`<div id="xmlights" aria-hidden="true">${segs}</div>`)); el.style.left=left+"px";
  }

  function decorate(){
    let stars=""; for(let i=0;i<80;i++) stars+=`<i style="left:${rnd(0,100).toFixed(1)}%;top:${rnd(0,55).toFixed(1)}%;animation-delay:${rnd(0,4).toFixed(2)}s;transform:scale(${rnd(.4,1.3).toFixed(2)})"></i>`;
    if(XM.v2) add(mk(`<div id="xmbg" class="v2" aria-hidden="true"><div class="xm-sun"></div><div class="xm-sleigh">${SLEIGH}</div><div class="xm-vila v2">${VILA2}</div><div class="xm-frost"></div></div>`));
    else add(mk(`<div id="xmbg" aria-hidden="true"><div class="xm-aurora a"></div><div class="xm-aurora b"></div><div class="xm-stars">${stars}</div>
      <div class="xm-moon"><i></i><i></i></div><div class="xm-sleigh">${SLEIGH}</div><div class="xm-vila">${VILA}</div></div>`));
    neveStart(); piscaPisca();
    // enfeite pendurado no menu (balança; clique: gira e solta brilho)
    const side=document.querySelector(".side"); if(side){ add(mk(`<div class="xm-hang" title="Ho ho ho!"><i class="xm-thread"></i>${BAUBLE("h","#e0283a")}</div>`), side); add(mk(`<div class="xm-hang b" aria-hidden="true"><i class="xm-thread"></i>${BAUBLE("h2","#2b7de9")}</div>`), side); add(mk(`<div class="xm-sidesnow" aria-hidden="true"></div>`), side); }
    // logo do menu vira bola de Natal
    const fl=document.querySelector(".logo .flask"); if(fl){ XM.orig.set("flask", fl.innerHTML); fl.innerHTML=BAUBLE("s"); }
    // início: árvore iluminada, estrelinhas subindo, pingentes de gelo no título e selo
    const hl=document.querySelector(".hero-logo"); if(hl) hl.insertAdjacentElement("afterend", add(mk(`<div class="xm-hero">${TREE("h")}<div class="xm-sparks"><i></i><i></i><i></i><i></i><i></i><i></i></div></div>`), hl.parentNode));
    const ht=document.querySelector(".hero-title"); if(ht){ let ic=""; [6,14,22,31,40,49,58,67,76,85,93].forEach((x,k)=>ic+=`<i style="left:${x}%;height:${(8+(k*7%13)).toFixed(0)}px;animation-delay:${(k*.45).toFixed(2)}s"></i>`); add(mk(`<div class="xm-icicles" aria-hidden="true">${ic}</div>`), ht); add(mk(`<div class="xm-titlesnow" aria-hidden="true"></div>`), ht); }
    const ap=document.querySelector(".alpha-pill"); if(ap) ap.insertAdjacentElement("afterend", add(mk(`<span class="xm-pill">🎄 Edição de Natal</span>`), ap.parentNode));
    const cdt=document.querySelector(".cd-tag"); if(cdt){ XM.orig.set("cdtag", cdt.textContent); cdt.textContent="ROLETA DO NOEL"; }
    agendaTreno();
  }
  function undecorate(){
    neveStop(); XM.els.forEach(e=>e.remove()); XM.els=[]; clearTimeout(trenoT);
    const fl=document.querySelector(".logo .flask"); if(fl && XM.orig.has("flask")) fl.innerHTML=XM.orig.get("flask");
    const cdt=document.querySelector(".cd-tag"); if(cdt && XM.orig.has("cdtag")) cdt.textContent=XM.orig.get("cdtag");
  }
  // o trenó passa na frente da lua de tempos em tempos
  let trenoT=0;
  function agendaTreno(first=true){ clearTimeout(trenoT); trenoT=setTimeout(()=>{ const s=document.querySelector(".xm-sleigh"); if(XM.on && s){ s.classList.remove("go"); void s.offsetWidth; s.classList.add("go"); } agendaTreno(false); }, first?6000:rnd(35000,70000)); }

  XM.set = function(on){
    on=!!on; if(on===XM.on) return;
    if(on && window.HW && HW.on) HW.set(false); // um tema de cada vez
    XM.on=on;
    styles().forEach(s=>{ if(!XM.orig.has(s)) XM.orig.set(s, s.textContent); s.textContent = on ? swap(XM.orig.get(s)) : XM.orig.get(s); });
    document.documentElement.classList.toggle("xm", on);
    document.documentElement.classList.toggle("xm2", on && XM.v2);
    on ? decorate() : undecorate();
    const t=document.getElementById("xmToggle"); if(t) t.classList.toggle("on", on);
  };

  // ---------- efeitos ----------
  function burst(x, y, items, opts={}){
    if(!XM.on) return;
    for(let i=0;i<items.length;i++){
      const e=document.createElement("div"); e.className="xm-pop "+(opts.cls||""); e.innerHTML=items[i];
      const ang = opts.up ? (-Math.PI/2 + (Math.random()-.5)*1.7) : opts.down ? (Math.PI/2 + (Math.random()-.5)*1.2) : Math.random()*Math.PI*2, dist=(opts.dist||140)*(.6+Math.random()*.6);
      e.style.left=x+"px"; e.style.top=y+"px";
      e.style.setProperty("--dx", (Math.cos(ang)*dist).toFixed(0)+"px"); e.style.setProperty("--dy", (Math.sin(ang)*dist - (opts.lift||0)).toFixed(0)+"px");
      e.style.setProperty("--rot", ((Math.random()-.5)*160).toFixed(0)+"deg"); e.style.animationDelay=(Math.random()*.12).toFixed(2)+"s";
      document.body.appendChild(e); setTimeout(()=>e.remove(), 1900);
    }
  }
  const FLOCO = `<svg viewBox="0 0 40 40" aria-hidden="true"><g stroke="#f4fbff" stroke-width="2.6" stroke-linecap="round" fill="none"><path d="M20 3v34M5 11l30 18M5 29l30-18"/><path d="M15 6l5 5 5-5M15 34l5-5 5 5M4 17l7-1-2-6M36 23l-7 1 2 6M4 23l7 1-2 6M36 17l-7-1 2-6"/></g></svg>`;
  XM.giftBurst = el => { const r=el.getBoundingClientRect(); burst(r.left+r.width/2, r.top+r.height/2, ["🎁",FLOCO,"⭐",FLOCO,"🎁",FLOCO,"✨","🔔"], {cls:"gift", up:true, dist:240, lift:90}); XM.nevasca(2.5); };
  document.addEventListener("click", e=>{
    if(!XM.on) return;
    const d=e.target.closest("#drawBtn"); if(d && !d.disabled){ XM.giftBurst(d); return; }
    const r=e.target.closest("[data-res]"); if(r && !r.classList.contains("on") && !r.closest(".locked")){
      const b=r.getBoundingClientRect();
      if(r.dataset.res==="w") burst(b.left+b.width/2, b.top+b.height/2, ["🎁","⭐","🍪","🎄","✨","🔔","🎁","⭐"], {dist:120});
      else burst(b.left+b.width/2, b.top+b.height/2, ["🪨","🪨","🪨","⛄"], {cls:"coal", down:true, dist:90});
    }
    const h=e.target.closest(".xm-hang"); if(h){ h.classList.remove("spin"); void h.offsetWidth; h.classList.add("spin"); const b=h.querySelector("svg").getBoundingClientRect(); burst(b.left+b.width/2, b.top+b.height/2, ["✨","⭐","✨",FLOCO,"✨"], {dist:70}); }
  }, true);

  // botão no menu
  function toggleBtn(){
    if(document.getElementById("xmToggle")) return;
    const ref=document.getElementById("hwToggle") || document.querySelector(".volbar"); if(!ref) return;
    const b=mk(`<button class="xmtoggle" id="xmToggle" title="Liga ou desliga o tema de Natal"><span>🎄</span>Natal<i class="sw"></i></button>`);
    ref.parentNode.insertBefore(b, ref);
    b.classList.toggle("on", XM.on);
    b.onclick=()=>{ XM.set(!XM.on); XM.save(); };
  }
  XM.save = () => {
    try{ S.opts.natal=XM.on; if(XM.on) S.opts.halloween=false; save(); }catch(e){}
    try{ localStorage.setItem("xmpref", XM.on?"1":"0"); if(XM.on) localStorage.setItem("hwpref","0"); }catch(e){}
    XM.music.sync();
  };

  // escolha do estilo (fica embaixo no menu; só aparece com o Natal ligado)
  function renderEstilo(){
    const b=document.getElementById("xmEstilo"); if(!b) return;
    b.style.display = XM.on ? "" : "none";
    b.innerHTML = `<span>Estilo</span><button data-e="2" class="${XM.v2?"on":""}">❄️ Branco</button><button data-e="1" class="${XM.v2?"":"on"}">🌙 Noite</button>`;
  }
  function estiloBtn(){
    const cr=document.querySelector(".credits"); if(!cr || document.getElementById("xmEstilo")) return;
    const b=mk(`<div class="xmestilo" id="xmEstilo" title="Escolha o visual do tema de Natal"></div>`);
    cr.insertAdjacentElement("afterend", b);
    b.onclick=e=>{ const x=e.target.closest("[data-e]"); if(!x) return; const n=+x.dataset.e; if((n===2)!==XM.v2) XM.esboco(n); };
    renderEstilo();
  }

  // ---------- sons (entram quando os arquivos existirem no app: snd/xm_*.mp3) ----------
  XM.has = { musica:false, roleta:false, prenda:false };
  [["musica","xm_musica1"],["roleta","xm_roleta"],["prenda","xm_prenda"]].forEach(([k,f])=>{ try{ fetch("snd/"+f+".mp3", {method:"HEAD"}).then(r=>{ XM.has[k]=r.ok; if(k==="musica") XM.music.sync(); }).catch(()=>{}); }catch(e){} });
  const ducks = new Set();
  const mv = () => { try{ return S.opts.xmMusic ?? 40; }catch(e){ return 40; } };
  const soundOn = () => { try{ return S.opts.sound!==false; }catch(e){ return true; } };
  const master = () => { try{ return Math.pow((S.opts.vol??70)/100, 1.8); }catch(e){ return .5; } };
  // trilha de fundo: 3 músicas em sequência (sempre começa pela 1), dá pra pular e voltar
  const FAIXAS = [{n:"xm_musica1", d:279.3},{n:"xm_musica2", d:175.5},{n:"xm_musica3", d:225.9}];
  const nativo = () => typeof NSND!=="undefined" && NSND.ok && typeof nativeSound==="function";
  XM.music = {
    a:null, i:0, nextT:0, keep:false,
    // pausada (prenda, contagem, vídeo do rato): fica muda e continua de onde parou, sem recomeçar
    vol(){ return (XM.minimized || ducks.size) ? 0 : Math.min(1, master()*mv()/100); },
    want(){ return (XM.on || this.keep) && XM.has.musica && soundOn() && mv()>0 && (typeof NSND==="undefined" || NSND.done || !window.sndReady); },
    ramp(to, ms, done, a=this.a){ if(!a) return; clearInterval(a._ft); const from=+a.volume||0, t0=Date.now();
      a._ft=setInterval(()=>{ const k=Math.min(1,(Date.now()-t0)/ms); try{ a.volume=from+(to-from)*k; }catch(e){} if(k>=1){ clearInterval(a._ft); done&&done(); } }, 60); },
    start(fadeMs=1500){
      const f=FAIXAS[this.i]; clearTimeout(this.nextT);
      let a;
      if(nativo()){ a=nativeSound(f.n, 0, false); this.nextT=setTimeout(()=>this.go(1, true), f.d*1000+300); }
      else { a=new Audio("snd/"+f.n+".mp3"); a.volume=0; a.onended=()=>{ if(this.a===a) this.go(1, true); }; a.play().catch(()=>{}); }
      this.a=a; this.ramp(this.vol(), fadeMs); renderMus();
    },
    stopCur(ms=500){ clearTimeout(this.nextT); const a=this.a; this.a=null; if(a) this.ramp(0, ms, ()=>{ try{ a.pause(); }catch(e){} }, a); },
    go(dir, auto=false){ // pula (+1) ou volta (-1) de música
      this.i=(this.i+dir+FAIXAS.length)%FAIXAS.length;
      if(!this.want()){ renderMus(); return; }
      this.stopCur(auto?200:350); this.start(auto?800:600);
    },
    sync(){
      if(this.want()){
        if(!this.a) this.start(1500);
        else this.ramp(this.vol(), ducks.size ? 400 : 900);
      } else if(this.a) this.stopCur(500);
      renderMus();
    },
  };
  XM.duck = (on, why="prenda") => { on ? ducks.add(why) : ducks.delete(why); XM.music.sync(); };
  XM.cdMusic = () => {
    if(!XM.on || !XM.has.roleta || !soundOn()) return null;
    XM.duck(true, "cd");
    let a; if(typeof NSND!=="undefined" && NSND.ok && typeof nativeSound==="function") a=nativeSound("xm_roleta", Math.min(1, master()), false);
    else { a=new Audio("snd/xm_roleta.mp3"); a.volume=Math.min(1, master()); a.play().catch(()=>{}); }
    return { stop(){ XM.music.ramp(0, 900, ()=>{ try{ a.pause(); }catch(e){} XM.duck(false, "cd"); }, a); } };
  };
  // janela minimizada: neve e música param
  const baseWin = window.__winState;
  window.__winState = st => { XM.minimized = st==="min"; XM.music.sync(); if(baseWin) baseWin(st); };
  document.addEventListener("visibilitychange", ()=>{ XM.minimized=document.hidden; XM.music.sync(); });

  // ---------- rato de Natal (vídeo com pisca-pisca; aparece ao abrir, com opção de não mostrar mais) ----------
  XM.rato = (manual=false) => {
    if(!XM.on || document.getElementById("xmRato")) return;
    let bulbs=""; const N=46;
    for(let i=0;i<N;i++){ // lâmpadas em volta do quadro (perímetro)
      const t=i/N, c=LUZES[i%LUZES.length]; let x,y;
      if(t<.25){ x=t/.25*100; y=0; } else if(t<.5){ x=100; y=(t-.25)/.25*100; } else if(t<.75){ x=100-(t-.5)/.25*100; y=100; } else { x=0; y=100-(t-.75)/.25*100; }
      bulbs+=`<i style="left:${x.toFixed(1)}%;top:${y.toFixed(1)}%;--c:${c};animation-delay:${((i%3)*.4).toFixed(1)}s"></i>`; }
    const el=mk(`<div class="xm-rato" id="xmRato" role="dialog" aria-label="Feliz Natal"><div class="xm-rato-card"><div class="xm-rato-lights" aria-hidden="true">${bulbs}</div>
      <div class="xm-rato-title">FELIZ NATAL</div><div class="xm-rato-sub">🎄 Ho ho ho! Boas festas e boas partidas 🎁</div>
      <div class="xm-rato-vid"><video src="vid/rato_natal.mp4" playsinline preload="auto"></video><button class="xm-rato-play" hidden>▶ Ver de novo</button><button class="xm-rato-som" hidden>🔊 Tocar com som</button></div>
      <div class="xm-rato-foot">${manual?"":`<label><input type="checkbox" id="xmRatoNo"> Não mostrar mais</label>`}<button class="btn gold" id="xmRatoOk">Fechar</button></div></div></div>`);
    document.body.appendChild(el);
    const v=el.querySelector("video"), rep=el.querySelector(".xm-rato-play"), som=el.querySelector(".xm-rato-som");
    const vol=()=>Math.min(1, master()*1.1);
    const tocar=()=>{ rep.hidden=true; v.currentTime=0; v.muted=!soundOn(); v.volume=vol(); XM.duck(true,"rato");
      v.play().catch(()=>{ v.muted=true; v.play().catch(()=>{}); if(soundOn()) som.hidden=false; }); };
    som.onclick=()=>{ som.hidden=true; v.muted=false; v.volume=vol(); };
    v.onended=()=>fechar(); // acabou o vídeo: fecha sozinho
    rep.onclick=tocar;
    const fechar=()=>{ try{ v.pause(); }catch(e){} XM.duck(false,"rato"); const no=el.querySelector("#xmRatoNo"); if(no && no.checked){ try{ S.opts.xmRatoOff=true; save(); }catch(e){} }
      el.classList.add("out"); setTimeout(()=>el.remove(), 350); document.removeEventListener("keydown", esc); };
    const esc=e=>{ if(e.key==="Escape") fechar(); };
    el.querySelector("#xmRatoOk").onclick=fechar; el.onclick=e=>{ if(e.target===el) fechar(); }; document.addEventListener("keydown", esc);
    requestAnimationFrame(()=>el.classList.add("on")); setTimeout(tocar, 450);
  };
  // ao abrir o app (uma vez por abertura), se ninguém pediu pra não mostrar mais
  function ratoAoAbrir(tent=0){
    setTimeout(()=>{
      if(!XM.on) return; try{ if(S.opts.xmRatoOff) return; }catch(e){}
      const ocupado = document.querySelector(".pfx, .cd.on, #overlay.on, .uic") || document.hidden;
      if(ocupado){ if(tent<20) ratoAoAbrir(tent+1); return; }
      XM.rato();
    }, tent ? 4000 : 3500);
  }
  // clicar na árvore do início mostra o rato de novo
  document.addEventListener("click", e=>{ if(XM.on && e.target.closest(".xm-hero")) XM.rato(true); });

  // trenó atravessando a roleta da prenda
  XM.pfxFly = () => `<div class="xm-pfxfly"><div class="xm-psleigh">${SLEIGH}</div></div>`;

  // controle da música no menu (só aparece quando a música existir)
  function renderMus(){
    const box=document.getElementById("xmMusic"), pl=document.getElementById("xmPlayer"); if(!box) return;
    const show = XM.on && XM.has.musica;
    box.style.display = show ? "" : "none"; if(pl) pl.style.display = show ? "" : "none";
    const v=mv(), r=document.getElementById("xmMusVol"), t=document.getElementById("xmMusPct"), b=document.getElementById("xmMusMute");
    if(r && document.activeElement!==r){ r.value=v; } if(r) r.style.setProperty("--p", v+"%");
    if(t) t.textContent = v ? v+"%" : "mudo"; if(b) b.classList.toggle("off", !v);
    const fx=document.getElementById("xmFaixa"); if(fx) fx.textContent = `Música ${XM.music.i+1} de ${FAIXAS.length}`;
  }
  function musicBtn(){
    const vb=document.querySelector(".volbar"); if(!vb || document.getElementById("xmMusic")) return;
    // de cima pra baixo: trocar música, volume da música, volume geral
    const pl=mk(`<div class="xmplayer" id="xmPlayer" title="Música de fundo do Natal"><button id="xmPrev" aria-label="Música anterior">⏮</button><span id="xmFaixa"></span><button id="xmNext" aria-label="Próxima música">⏭</button></div>`);
    const box=mk(`<div class="xmmusic" id="xmMusic" title="Volume da música de Natal"><button id="xmMusMute" aria-label="Ligar ou desligar a música">🎵</button><input type="range" id="xmMusVol" min="0" max="100" aria-label="Volume da música"><span id="xmMusPct"></span></div>`);
    vb.parentNode.insertBefore(pl, vb); vb.parentNode.insertBefore(box, vb);
    let last=40;
    document.getElementById("xmMusVol").oninput=e=>{ try{ S.opts.xmMusic=+e.target.value; save(); }catch(_){} XM.music.sync(); };
    document.getElementById("xmMusMute").onclick=()=>{ try{ const v=mv(); if(v){ last=v; S.opts.xmMusic=0; } else S.opts.xmMusic=last||40; save(); }catch(_){} XM.music.sync(); };
    document.getElementById("xmPrev").onclick=()=>XM.music.go(-1);
    document.getElementById("xmNext").onclick=()=>XM.music.go(1);
    renderMus();
  }
  const baseSet = XM.set;
  XM.set = function(on){ baseSet(on); renderMus(); renderEstilo(); XM.music.sync(); };

  XM.init = function(){
    toggleBtn(); musicBtn(); estiloBtn();
    try{ const e=S.opts.natalEstilo; if((e==="branco"||e==="noite") && (e==="branco")!==XM.v2){ XM.set(false); XM.orig=new Map(); XM.v2 = e==="branco"; try{ localStorage.setItem("xmvar", XM.v2?"2":"1"); }catch(_){} } }catch(_){}
    let want=true; try{ want = S.opts.natal!==false; }catch(e){}
    XM.set(want); renderMus(); XM.music.sync();
    ratoAoAbrir();
  };
  window.XM = XM;
  // aplica já na abertura, antes de mostrar a janela
  try{ if(window.__XMON) XM.set(true); }catch(e){ console.error(e); }
  document.documentElement.classList.remove("xmpre");
})();
