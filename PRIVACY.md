# Política de privacidade / Privacy policy

O Ciência LOL **não coleta nem envia dados pessoais para os autores**. Não há telemetria, anúncios
nem contas. As conexões que o app faz, e quando:

| Destino | Quando | O que é enviado |
|---|---|---|
| Cliente do LoL no próprio PC (127.0.0.1) | Ao clicar em *Importar pro LoL*, *Minhas partidas* ou *Pegar estatísticas* | Runas, feitiços e itens a importar; leitura das suas últimas partidas. Nada sai do PC. |
| Riot Games Data Dragon (ddragon.leagueoflegends.com) | Ao abrir o app | Só download de imagens e dados públicos do jogo. |
| Servidores MQTT públicos (broker.emqx.io, broker.hivemq.com, test.mosquitto.org, mqtt.eclipseprojects.io) | Só quando você cria ou entra numa sala online | Nome que você digitou, função, campeão e resultados da sala, **criptografados** (ChaCha20 + HMAC) com uma chave derivada do código e da senha da sala. |
| Google Docs / Google Drive | Ao abrir o app e ao clicar em *Verificar atualização* | Só leitura do número da versão mais nova e, se você aceitar, download dela. |
| Google Apps Script (nuvem do grupo) | Em segundo plano, a cada minuto, nas versões oficiais | Ciências criadas, curtidas, comentários e correções de build, e um backup pessoal (histórico e favoritas) ligado ao nome que você digitou. |
| op.gg (mcp-api.op.gg) | Só ao clicar em *op.gg* numa carta | Nome do campeão e a função. |
| Links externos (Instagram, YouTube, FunLeagueBuilds, op.gg) | Só ao clicar neles | Abre no seu navegador. |

Os dados do app (histórico, favoritas, opções) ficam só no seu PC, em `%APPDATA%\CienciaLoL`.
Para desinstalar, apague o `CienciaLOL.exe` e essa pasta.

*The app collects no personal data and has no telemetry. Official builds sync the group library and a
personal backup (history, favorites) to a Google Apps Script owned by the author and check Google Drive for updates. Online rooms send only what you type
(name, role, champion, results), end-to-end encrypted, through public MQTT brokers, and only when you
create or join a room. op.gg is queried only on demand. App data stays in `%APPDATA%\CienciaLoL`.*
