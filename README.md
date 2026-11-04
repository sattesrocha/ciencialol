# Ciência LOL

Versão oficial do Ciência LOL: roleta de builds fora do comum ("ciências") para League of Legends, com salas online
entre amigos e importação de runas, feitiços e itens direto no cliente do LoL.
App para Windows feito em **Go + WebView2**, com a interface em HTML/CSS/JS.

*An off-meta build roulette for League of Legends with online rooms and one-click import
of runes, summoner spells and item sets into the League client (Windows, Go + WebView2).*

## Download
Baixe pelo site: **https://sattesrocha.github.io/ciencialol/**. Não precisa instalar: é só abrir.
O próprio app avisa quando sai versão nova e se atualiza sozinho.

## O que o app faz
- **🧪 Science**: sala online em que o anfitrião sorteia uma ciência pra cada jogador.
- **✋ Freedom**: sala online em que cada um escolhe o campeão (pode ser surpresa) e puxa a build
  da Biblioteca, das suas últimas partidas ou a mais usada no op.gg.
- **Importar pro LoL**: coloca runas, feitiços e conjunto de itens no cliente do LoL aberto no PC.
- Biblioteca de ciências, histórico, rankings e tema de Halloween.

## Compilar
Requer Go 1.24+ (no Windows, ou cruzado de outro sistema):
```
go install github.com/tc-hib/go-winres@v0.3.3
go-winres make --in winres/winres.json --arch amd64 --out rsrc
set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w -H windowsgui" -o CienciaLOL.exe .
```
Os binários oficiais são compilados pelo GitHub Actions (`.github/workflows/build.yml`).
A nuvem do grupo e a chave das salas entram só na hora de compilar, pelos segredos do repositório;
sem eles o app funciona normal, só sem sincronizar a biblioteca.

## Privacidade e assinatura de código
- [Política de privacidade](PRIVACY.md)
- [Code signing policy](CODE_SIGNING_POLICY.md)

## Autor
Felipe Rocha · Instagram [@sattesrocha](https://www.instagram.com/sattesrocha/)

## Créditos
- Ideias de builds catalogadas a partir de posts públicos do perfil @vinylol1 (Instagram).
- Imagens e dados do jogo: Riot Games Data Dragon, carregados na hora de usar.
- Fontes Cinzel e Inter (SIL Open Font License).
- Sons do tema de Halloween e da roleta de prendas: trechos de terceiros, usados sem fins lucrativos; se você é o dono de algum e quer que seja removido, abra uma issue.

Ciência LOL isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games
or anyone officially involved in producing or managing Riot Games properties. Riot Games and all
associated properties are trademarks or registered trademarks of Riot Games, Inc.
