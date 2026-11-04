// Salas criptografadas: o código da sala (ex.: RDC-42) + a chave do grupo (só existe dentro do app)
// viram a chave da sala. O servidor público só vê nomes de canal e mensagens embaralhados.
// ChaCha20 + HMAC-SHA256, tudo síncrono e sem dependências.
"use strict";
(function(){
  const te=new TextEncoder(), td=new TextDecoder();
  const K=new Uint32Array([0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2]);
  function sha256(msg){
    const l=msg.length, n=((l+9+63)>>6)<<6, b=new Uint8Array(n); b.set(msg); b[l]=0x80;
    const dv=new DataView(b.buffer); dv.setUint32(n-4, l*8>>>0); dv.setUint32(n-8, Math.floor(l/0x20000000));
    const H=new Uint32Array([0x6a09e667,0xbb67ae85,0x3c6ef372,0xa54ff53a,0x510e527f,0x9b05688c,0x1f83d9ab,0x5be0cd19]), W=new Uint32Array(64);
    const r=(x,k)=>(x>>>k)|(x<<(32-k));
    for(let o=0;o<n;o+=64){
      for(let i=0;i<16;i++) W[i]=dv.getUint32(o+i*4);
      for(let i=16;i<64;i++){ const s0=r(W[i-15],7)^r(W[i-15],18)^(W[i-15]>>>3), s1=r(W[i-2],17)^r(W[i-2],19)^(W[i-2]>>>10); W[i]=(W[i-16]+s0+W[i-7]+s1)|0; }
      let [a,bb,c,d,e,f,g,h]=H;
      for(let i=0;i<64;i++){ const t1=(h+(r(e,6)^r(e,11)^r(e,25))+((e&f)^(~e&g))+K[i]+W[i])|0, t2=((r(a,2)^r(a,13)^r(a,22))+((a&bb)^(a&c)^(bb&c)))|0;
        h=g; g=f; f=e; e=(d+t1)|0; d=c; c=bb; bb=a; a=(t1+t2)|0; }
      H[0]+=a; H[1]+=bb; H[2]+=c; H[3]+=d; H[4]+=e; H[5]+=f; H[6]+=g; H[7]+=h;
    }
    const out=new Uint8Array(32), ov=new DataView(out.buffer); for(let i=0;i<8;i++) ov.setUint32(i*4,H[i]); return out;
  }
  const cat=(...a)=>{ const o=new Uint8Array(a.reduce((s,x)=>s+x.length,0)); let p=0; a.forEach(x=>{ o.set(x,p); p+=x.length; }); return o; };
  const u8=s=>typeof s==="string"?te.encode(s):s;
  function hmac(key, msg){ key=u8(key); if(key.length>64) key=sha256(key); const k=new Uint8Array(64); k.set(key);
    const ip=k.map(x=>x^0x36), op=k.map(x=>x^0x5c); return sha256(cat(op, sha256(cat(ip, u8(msg))))); }
  function chacha(key, nonce, data){
    const k=new DataView(key.buffer,key.byteOffset,32), nv=new DataView(nonce.buffer,nonce.byteOffset,12), out=new Uint8Array(data.length);
    const s=new Uint32Array(16), x=new Uint32Array(16);
    s[0]=0x61707865; s[1]=0x3320646e; s[2]=0x79622d32; s[3]=0x6b206574;
    for(let i=0;i<8;i++) s[4+i]=k.getUint32(i*4,true);
    for(let i=0;i<3;i++) s[13+i]=nv.getUint32(i*4,true);
    const R=(a,n)=>(a<<n)|(a>>>(32-n));
    const q=(a,b,c,d)=>{ x[a]+=x[b]; x[d]=R(x[d]^x[a],16); x[c]+=x[d]; x[b]=R(x[b]^x[c],12); x[a]+=x[b]; x[d]=R(x[d]^x[a],8); x[c]+=x[d]; x[b]=R(x[b]^x[c],7); };
    for(let off=0, ctr=1; off<data.length; off+=64, ctr++){
      s[12]=ctr; x.set(s);
      for(let i=0;i<10;i++){ q(0,4,8,12); q(1,5,9,13); q(2,6,10,14); q(3,7,11,15); q(0,5,10,15); q(1,6,11,12); q(2,7,8,13); q(3,4,9,14); }
      const ks=new Uint8Array(64), kv=new DataView(ks.buffer); for(let i=0;i<16;i++) kv.setUint32(i*4,(x[i]+s[i])>>>0,true);
      for(let i=0;i<64 && off+i<data.length;i++) out[off+i]=data[off+i]^ks[i];
    }
    return out;
  }
  const hex=b=>[...b].map(x=>x.toString(16).padStart(2,"0")).join("");
  const b64=b=>{ let s=""; for(let i=0;i<b.length;i+=0x8000) s+=String.fromCharCode.apply(null,b.subarray(i,i+0x8000)); return btoa(s); };
  const unb64=s=>Uint8Array.from(atob(s),c=>c.charCodeAt(0));
  // uma "fechadura" por sala: o código + a senha da sala (+ a chave do grupo, que só existe no app).
  // O canal "/mk" (envelope pra senha mestre e aviso de sala ocupada) usa só o código.
  function RoomCrypt(code, pw){
    const secret = (typeof window.__RS==="string" && window.__RS) || "ciencialol-teste";
    const C = String(code).toUpperCase();
    const mkBase = b => ({ base:b, ek:hmac(b,"enc"), mk:hmac(b,"mac") });
    const kc = mkBase(hmac(secret, "sala|"+C));
    const kr = pw ? mkBase(hmac(secret, "sala|"+C+"|senha|"+pw)) : kc;
    const names = new Map();
    const keyOf = t => /\/mk$/.test(t) ? kc : kr;
    const topic = t => { const k=keyOf(t), p = "cl2/"+hex(hmac(k.base, "canal|"+t)).slice(0,32); names.set(p, t); return p; };
    return {
      topic,
      chk: hex(hmac(kr.base, "confere")).slice(0,24),
      enc(t, payload){ const k=keyOf(t), n=new Uint8Array(12); crypto.getRandomValues(n); const ct=chacha(k.ek, n, te.encode(payload));
        const tag=hmac(k.mk, cat(te.encode(t+"|"), n, ct)).slice(0,16); return "e1."+b64(cat(n, ct, tag)); },
      dec(phys, payload){ const t=names.get(phys); if(!t || typeof payload!=="string" || payload.slice(0,3)!=="e1.") return null;
        const k=keyOf(t); let b; try{ b=unb64(payload.slice(3)); }catch(e){ return null; } if(b.length<28) return null;
        const n=b.slice(0,12), ct=b.slice(12,b.length-16), tag=b.slice(b.length-16), exp=hmac(k.mk, cat(te.encode(t+"|"), n, ct)).slice(0,16);
        let d=0; for(let i=0;i<16;i++) d|=tag[i]^exp[i]; if(d) return null;
        try{ return { topic:t, payload: td.decode(chacha(k.ek, n, ct)) }; }catch(e){ return null; } }
    };
  }
  window.RoomCrypt = RoomCrypt; window.__rcTest = { sha256, hmac, chacha, hex };
})();
