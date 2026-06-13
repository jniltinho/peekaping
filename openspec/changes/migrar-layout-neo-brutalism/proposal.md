## Why

O layout atual do `apps/web` usa um design genérico shadcn/Radix sem personalidade visual forte. A migração para Neo-Brutalism (neobrutalism.dev) vai diferenciar o produto com uma estética marcante, moderna e de alto contraste — bordas grossas pretas, sombras deslocadas (hard shadow) e cores vibrantes, sem cantos arredondados.

## What Changes

- Atualizar design tokens CSS (`index.css`) com bordas pretas espessas, sombras `neo-shadow` e paleta de alto contraste
- Remover todos os `rounded-*` e garantir `border-radius: 0` em todos os componentes UI
- Adicionar variante `neo` nos componentes `Button`, `Card`, `Input`, `Badge`, `Alert` e `Select` com estilo Neo-Brutalism (border 2px black, box-shadow offset)
- Atualizar componentes globais (`app-sidebar`, `app-header`, `nav-main`, `nav-user`) para usar bordas e sombras Neo-Brutalism
- Ajustar modo escuro (`dark`) para manter contraste alto com bordas brancas/cinza-claro e sombras visíveis

## Capabilities

### New Capabilities
- `neo-brutalism-theme`: Sistema de design tokens e utilitários CSS para o estilo Neo-Brutalism aplicado ao `apps/web` (bordas 2px/4px pretas, `neo-shadow`, sem border-radius, paleta high-contrast)

### Modified Capabilities

## Impact

- `apps/web/src/index.css`: novos tokens CSS (neo-shadow, border widths, cores high-contrast)
- `apps/web/src/components/ui/*.tsx`: todos os componentes shadcn/Radix recebem estilos Neo-Brutalism
- `apps/web/src/components/app-sidebar.tsx`, `app-header.tsx`, `nav-main.tsx`, `nav-user.tsx`: componentes globais atualizados
- Sem quebra de API — mudanças são puramente visuais/CSS
