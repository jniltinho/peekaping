## 1. Design Tokens CSS

- [x] 1.1 Adicionar `--neo-shadow` (`4px 4px 0px 0px oklch(0 0 0)`) e `--neo-shadow-sm` (`2px 2px 0px 0px oklch(0 0 0)`) como tokens no `@theme inline` em `apps/web/src/index.css`
- [x] 1.2 Definir `--border` como preto puro (`oklch(0 0 0)`) no `:root` e branco puro (`oklch(1 0 0)`) no `.dark`
- [x] 1.3 Adicionar override no `.dark` para `--neo-shadow` e `--neo-shadow-sm` usando branco puro
- [x] 1.4 Confirmar `--radius: 0px` em `:root` e verificar que nenhum `--radius-*` produz valor diferente de 0

## 2. Componentes UI Base

- [x] 2.1 `button.tsx`: adicionar `border-2 border-black shadow-neo` na base da `cva`; adicionar `hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-neo-sm transition-all duration-100`; remover `shadow-xs` de todas as variantes
- [x] 2.2 `card.tsx`: adicionar `border-2 border-black shadow-neo`; remover `shadow-sm`
- [x] 2.3 `input.tsx`: adicionar `border-2 border-black shadow-neo-sm`; remover `border` (1px padrão)
- [x] 2.4 `textarea.tsx`: mesmas mudanças do `input.tsx`
- [x] 2.5 `select.tsx`: adicionar `border-2 border-black` no trigger; adicionar `shadow-neo-sm` no content
- [x] 2.6 `badge.tsx`: adicionar `border-2 border-black shadow-neo-sm`
- [x] 2.7 `alert.tsx`: adicionar `border-2 border-black shadow-neo`; remover `border` padrão
- [x] 2.8 `dialog.tsx`: adicionar `border-2 border-black shadow-neo` no content
- [x] 2.9 `sheet.tsx`: adicionar `border-2 border-black` nas bordas relevantes (direita/esquerda dependendo da side)
- [x] 2.10 `popover.tsx`: adicionar `border-2 border-black shadow-neo` no content
- [x] 2.11 `tooltip.tsx`: adicionar `border border-black` e remover `rounded-*`
- [x] 2.12 `command.tsx`: adicionar `border-2 border-black shadow-neo` no dialog wrapper
- [x] 2.13 `dropdown-menu.tsx`: adicionar `border-2 border-black shadow-neo` no content
- [x] 2.14 `checkbox.tsx`: adicionar `border-2 border-black` e remover `rounded-sm`
- [x] 2.15 `switch.tsx`: adicionar `border-2 border-black`; garantir `rounded-none`

## 3. Componentes Globais

- [x] 3.1 `app-sidebar.tsx`: substituir sombra lateral suave por `border-r-2 border-black`
- [x] 3.2 `app-header.tsx`: adicionar `border-b-2 border-black`
- [x] 3.3 `nav-main.tsx`: adicionar `border-l-4 border-black` no item ativo; garantir fundo de destaque high-contrast
- [x] 3.4 `nav-user.tsx`: adicionar `border-2 border-black shadow-neo` no dropdown (via atualização do dropdown-menu.tsx em 2.13)

## 4. Dark Mode

- [x] 4.1 Verificar visualmente em dark mode: cards, buttons, inputs, sidebar e header com bordas brancas visíveis
- [x] 4.2 Ajustar `--border` no `.dark` se necessário para garantir contraste adequado em todos os componentes
- [x] 4.3 Confirmar que sombras no dark mode usam `oklch(1 0 0)` e são visíveis sobre fundos escuros

## 5. Verificação Visual

- [ ] 5.1 Abrir dashboard de monitores e verificar cards, tabelas e botões com estilo Neo-Brutalism
- [ ] 5.2 Verificar página de login (formulário com inputs, botão de submit)
- [ ] 5.3 Verificar página de settings e modal de configuração de monitor
- [ ] 5.4 Verificar página de status pages e notification channels
- [x] 5.5 Rodar `pnpm build` em `apps/web` e confirmar zero erros de TypeScript/build
