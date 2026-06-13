## Context

O `apps/web` usa Tailwind CSS v4 + Radix UI (shadcn) com tokens de design padrão. O CSS já tem `--radius: 0px` e `rounded-none` nos componentes, mas falta o conjunto completo de características Neo-Brutalism: bordas grossas pretas, sombras hard-offset e paleta de alto contraste. As mudanças são puramente visuais — sem alteração de lógica ou API.

## Goals / Non-Goals

**Goals:**
- Implementar tokens CSS Neo-Brutalism em `index.css` (bordas 2px-4px pretas, `neo-shadow`, cores high-contrast)
- Atualizar todos os componentes `ui/` para aplicar borda e sombra Neo-Brutalism por padrão
- Atualizar componentes globais (sidebar, header, nav) para visual Neo-Brutalism coeso
- Manter compatibilidade com modo escuro (dark theme)
- Zero cantos arredondados (`border-radius: 0` em todo o sistema)

**Non-Goals:**
- Mudanças no `apps/landing` ou `apps/docs`
- Alteração de comportamento funcional dos componentes
- Introdução de nova biblioteca de componentes (continua com shadcn/Radix)
- Sistema de temas alternáveis em runtime (Neo-Brutalism é o único tema)

## Decisions

### D1: Tokens via CSS Custom Properties em `index.css`

Adicionar variáveis CSS dedicadas ao tema Neo-Brutalism diretamente no `:root` do `index.css` do Tailwind v4, usando a diretiva `@theme inline`:

```css
--neo-shadow: 4px 4px 0px 0px oklch(0 0 0);
--neo-shadow-sm: 2px 2px 0px 0px oklch(0 0 0);
--neo-border: 2px solid oklch(0 0 0);
--neo-border-thick: 4px solid oklch(0 0 0);
```

Expor como classes Tailwind via `@theme inline` → `--shadow-neo`, `--shadow-neo-sm`.

**Alternativa considerada**: Criar plugin Tailwind separado. Rejeitado — complexidade desnecessária no Tailwind v4 que já suporta tokens via `@theme`.

### D2: Aplicar estilo Neo-Brutalism diretamente nas classes base dos componentes

Atualizar as `cva` base de cada componente UI (Button, Card, Input, Badge, Alert, Select, Dialog, Sheet, etc.) para incluir `border-2 border-black shadow-neo` por padrão, em vez de criar uma "variante neo" adicional.

**Alternativa considerada**: Adicionar variante `neo` em cada componente. Rejeitado — exigiria atualizar todos os call sites para passar `variant="neo"`, o que gera trabalho proporcional ao número de usos no codebase.

### D3: Paleta de cores high-contrast

Manter a cor primária roxa atual (`oklch(0.6405 0.1889 297.89)`) mas tornar `--border` preto puro (`oklch(0 0 0)`) e `--input` branco com borda preta explícita. Fundo permanece branco/preto (sem cores de fundo vibrantes na base).

**Alternativa considerada**: Paleta Neo-Brutalism full com amarelos e rosas vibrantes como fundo. Rejeitado — riscos de legibilidade e conflito com gráficos de status (verde/vermelho) do dashboard de monitoramento.

### D4: Dark mode com bordas em branco

No modo escuro, substituir `oklch(0 0 0)` nas bordas/sombras por `oklch(1 0 0)` (branco puro) para manter o contraste característico do Neo-Brutalism.

```css
.dark {
  --neo-shadow: 4px 4px 0px 0px oklch(1 0 0);
  --neo-border: 2px solid oklch(1 0 0);
}
```

## Risks / Trade-offs

- [Visibilidade dos shadows no dark mode] → Usar branco puro nas sombras escuras garante contraste; testar visualmente no dashboard
- [Bordas 2px em tabelas densas] → Tabelas com muitas colunas podem ficar pesadas visualmente; usar `border` só em thead/tbody ao invés de cada célula
- [Compatibilidade com tooltips/popovers Radix] → Elementos floating precisam de `shadow-neo` manual; verificar `tooltip.tsx`, `popover.tsx` e `command.tsx`
- [Tamanho de layout com border 2px] → Border adiciona 4px ao box-sizing; garantir que `box-sizing: border-box` (default Tailwind) está em todos os elementos

## Migration Plan

1. Atualizar `index.css` com tokens Neo-Brutalism (D1, D3, D4)
2. Atualizar componentes UI base em `components/ui/` (D2) — começar pelos mais usados: `button`, `card`, `input`, `badge`, `dialog`, `select`
3. Atualizar componentes globais: `app-sidebar`, `app-header`, `nav-main`, `nav-user`
4. Verificar visualmente as páginas principais: dashboard de monitores, login, settings
5. Ajustes finos de dark mode

**Rollback**: Todas as mudanças são CSS/TSX — reverter via `git revert` ou restaurar os arquivos originais.

## Open Questions

- Usar sombra de `4px` ou `2px` como padrão nos cards do dashboard? (cards têm muito conteúdo, 2px pode ser mais adequado)
- Manter ou remover `shadow-sm` dos componentes que já tinham sombra sutil antes da migração?
