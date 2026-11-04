// Cliente MQTT 3.1.1 mínimo sobre WebSocket (QoS 0, retain, will, keepalive, reconexão).
// Sem dependências externas.
"use strict";
(function(){
  const te = new TextEncoder(), td = new TextDecoder();

  function encLen(n){ const out=[]; do{ let b=n%128; n=Math.floor(n/128); if(n>0) b|=128; out.push(b); }while(n>0); return out; }
  function str(s){ const b=te.encode(s); return [b.length>>8, b.length&255, ...b]; }
  function packet(header, body){ return new Uint8Array([header, ...encLen(body.length), ...body]); }

  class MiniMQTT {
    constructor(urls, opts){
      this.urls = urls; this.urlIdx = 0; this.opts = opts||{};
      this.clientId = (opts.clientId || "cl") + "-" + Math.random().toString(36).slice(2,10);
      this.subs = new Set(); this.handlers = []; this.buf = new Uint8Array(0);
      this.connected = false; this.closedByUser = false; this.pid = 1; this.retry = 0;
      this.onstatus = () => {};
    }
    connect(){
      this.closedByUser = false;
      const url = this.urls[this.urlIdx % this.urls.length];
      let ws;
      try { ws = new WebSocket(url, "mqtt"); } catch(e){ this._fail(); return; }
      ws.binaryType = "arraybuffer"; this.ws = ws; this.buf = new Uint8Array(0);
      this.onstatus("connecting", url);
      const timeout = setTimeout(()=>{ if(!this.connected) try{ ws.close(); }catch(e){} }, 8000);
      ws.onopen = () => this._sendConnect();
      ws.onmessage = ev => this._onData(new Uint8Array(ev.data));
      ws.onclose = () => { clearTimeout(timeout); this._closed(); };
      ws.onerror = () => {};
    }
    _sendConnect(){
      const o = this.opts; let flags = 0x02; // clean session
      const payload = [...str(this.clientId)];
      if(o.will){ flags |= 0x04 | (o.will.retain?0x20:0); const c=o.crypt; payload.push(...str(c?c.topic(o.will.topic):o.will.topic), ...str(c?c.enc(o.will.topic,o.will.payload):o.will.payload)); }
      const body = [0,4,77,81,84,84, 4, flags, 0, 30, ...payload];
      this._raw(packet(0x10, body));
    }
    _raw(bytes){ try{ if(this.ws && this.ws.readyState===1) this.ws.send(bytes); }catch(e){} }
    _onData(chunk){
      const b = new Uint8Array(this.buf.length + chunk.length); b.set(this.buf); b.set(chunk, this.buf.length); this.buf = b;
      while(true){
        if(this.buf.length < 2) return;
        let mult=1, len=0, i=1, byte;
        do{ if(i>=this.buf.length) return; byte=this.buf[i++]; len += (byte&127)*mult; mult*=128; }while(byte&128);
        if(this.buf.length < i+len) return;
        const type = this.buf[0], body = this.buf.slice(i, i+len); this.buf = this.buf.slice(i+len);
        this._packet(type, body);
      }
    }
    _packet(type, body){
      const t = type>>4;
      if(t===2){ // CONNACK
        if(body[1]!==0){ this.onstatus("refused"); try{this.ws.close();}catch(e){} return; }
        this.connected = true; this.retry = 0; this.onstatus("connected", this.urls[this.urlIdx % this.urls.length]);
        for(const s of this.subs) this._sub(s);
        clearInterval(this.ping); this.ping = setInterval(()=>this._raw(new Uint8Array([0xC0,0])), 20000);
        this.opts.onconnect && this.opts.onconnect();
      } else if(t===3){ // PUBLISH
        const qos=(type>>1)&3, retained=!!(type&1);
        const tl=(body[0]<<8)|body[1]; const topic=td.decode(body.slice(2,2+tl));
        let p=2+tl; if(qos>0) p+=2;
        let payload = td.decode(body.slice(p)), lt = topic;
        if(this.opts.crypt){ const m=this.opts.crypt.dec(topic, payload); if(!m) return; lt=m.topic; payload=m.payload; }
        for(const h of this.handlers) try{ h(lt, payload, retained); }catch(e){ console.error(e); }
      }
    }
    _ph(topic){ return this.opts.crypt ? this.opts.crypt.topic(topic) : topic; }
    _sub(topic){ const id=this.pid++ & 0xffff || 1; this._raw(packet(0x82, [id>>8, id&255, ...str(this._ph(topic)), 0])); }
    subscribe(topic){ this.subs.add(topic); if(this.connected) this._sub(topic); }
    unsubscribe(topic){ this.subs.delete(topic); if(this.connected){ const id=this.pid++ & 0xffff || 1; this._raw(packet(0xA2,[id>>8,id&255,...str(this._ph(topic))])); } }
    publish(topic, payload, retain){ if(!this.connected) return false; const c=this.opts.crypt;
      this._raw(packet(0x30 | (retain?1:0), [...str(c?c.topic(topic):topic), ...te.encode(c?c.enc(topic,payload):payload)])); return true; }
    on(fn){ this.handlers.push(fn); }
    _closed(){
      const was = this.connected; this.connected=false; clearInterval(this.ping);
      if(this.closedByUser){ this.onstatus("closed"); return; }
      if(!was) this.urlIdx++; // servidor não respondeu: tenta o próximo da lista
      this._fail();
    }
    _fail(){
      this.onstatus("reconnecting");
      const wait = Math.min(1000 * Math.pow(1.6, this.retry++), 10000);
      clearTimeout(this.rt); this.rt = setTimeout(()=>{ if(!this.closedByUser) this.connect(); }, wait);
    }
    end(){ this.closedByUser = true; clearTimeout(this.rt); clearInterval(this.ping);
      if(this.ws){ try{ this._raw(new Uint8Array([0xE0,0])); this.ws.close(); }catch(e){} } this.connected=false; }
  }
  window.MiniMQTT = MiniMQTT;
})();
